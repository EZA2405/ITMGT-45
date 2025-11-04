package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type IndexPageData struct {
	Username string
	Products []Product
}

type CartPageData struct {
	CartItems []CartItem
	User      User
}

type TxHistoryPageData struct {
	User         User
	Transactions []Transaction
}

func generateSessionToken() string {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		log.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(raw)
}

/* ===== Handlers ===== */

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./templates/index.html")
	if err != nil {
		log.Fatal(err)
	}

	var token string
	for _, c := range r.Cookies() {
		if c.Name == "cafego_session" {
			token = c.Value
			break
		}
	}
	user := getUserFromSessionToken(token)

	data := IndexPageData{
		Username: user.Username,
		Products: getProducts(),
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Fatal(err)
	}
}

func productHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 3 {
			http.NotFound(w, r)
			return
		}
		idStr := parts[len(parts)-1]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		var prod Product
		for _, p := range getProducts() {
			if p.Id == id {
				prod = p
				break
			}
		}
		if prod == (Product{}) {
			http.NotFound(w, r)
			return
		}

		tmpl, err := template.ParseFiles("./templates/product.html")
		if err != nil {
			log.Fatal(err)
		}
		if err := tmpl.Execute(w, prod); err != nil {
			log.Fatal(err)
		}
		return
	}

	if r.Method == http.MethodPost {
		var token string
		for _, c := range r.Cookies() {
			if c.Name == "cafego_session" {
				token = c.Value
				break
			}
		}
		user := getUserFromSessionToken(token)
		if user == (User{}) {
			fmt.Fprint(w, "You must be logged in to add to cart.")
			return
		}

		pidStr := r.FormValue("product_id")
		qtyStr := r.FormValue("quantity")

		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			log.Fatal(err)
		}
		qty, err := strconv.Atoi(qtyStr)
		if err != nil {
			log.Fatal(err)
		}

		createCartItem(user.Id, pid, qty)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	http.NotFound(w, r)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("./templates/login.html")
		if err != nil {
			log.Fatal(err)
		}
		if err := tmpl.Execute(w, nil); err != nil {
			log.Fatal(err)
		}
		return
	}

	// POST
	username := r.FormValue("username")
	password := r.FormValue("password")

	var user User
	for _, u := range getUsers() {
		if u.Username == username && u.Password == password {
			user = u
			break
		}
	}
	if user == (User{}) {
		fmt.Fprint(w, "Invalid login. Please go back and try again.")
		return
	}

	token := generateSessionToken()
	setSession(token, user)

	http.SetCookie(w, &http.Cookie{
		Name:  "cafego_session",
		Value: token,
		Path:  "/",
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

func cartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFiles("./templates/cart.html")
		if err != nil {
			log.Fatal(err)
		}

		var token string
		for _, c := range r.Cookies() {
			if c.Name == "cafego_session" {
				token = c.Value
				break
			}
		}
		user := getUserFromSessionToken(token)

		var items []CartItem
		if user != (User{}) {
			items = getCartItemsByUser(user)
		}

		data := CartPageData{
			User:      user,
			CartItems: items,
		}
		if err := tmpl.Execute(w, data); err != nil {
			log.Fatal(err)
		}
		return
	}

	if r.Method == http.MethodPost {
		// Checkout via POST /cart/
		var token string
		for _, c := range r.Cookies() {
			if c.Name == "cafego_session" {
				token = c.Value
				break
			}
		}
		user := getUserFromSessionToken(token)
		if user == (User{}) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		checkoutItemsForUser(user)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	http.NotFound(w, r)
}

func txHistoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFiles("./templates/transactions.html")
	if err != nil {
		log.Fatal(err)
	}

	var token string
	for _, c := range r.Cookies() {
		if c.Name == "cafego_session" {
			token = c.Value
			break
		}
	}
	user := getUserFromSessionToken(token)
	if user == (User{}) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	txns := getTransactionsByUser(user)
	data := TxHistoryPageData{
		User:         user,
		Transactions: txns,
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Fatal(err)
	}
}

func main() {
	initDB()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/product/", productHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/login/", loginHandler)
	http.HandleFunc("/cart/", cartHandler)
	http.HandleFunc("/transactions/", txHistoryHandler)

	log.Println("✅ CafeGo running at: http://localhost:3000")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
