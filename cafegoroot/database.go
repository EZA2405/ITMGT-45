package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

/* ===== Types ===== */

type User struct {
	Id       int
	Username string
	Password string
}

type Product struct {
	Id          int
	Name        string
	Price       int
	Description string
}

type Session struct {
	Token  string
	UserId int
}

type CartItem struct {
	Id          int
	UserId      int
	ProductId   int
	Quantity    int
	ProductName string
}

type LineItem struct {
	Id            int
	TransactionId int
	ProductId     int
	Quantity      int
	ProductName   string
	Price         int
}

type Transaction struct {
	Id        int
	UserId    int
	CreatedAt string
	Total     int        // computed sum (Price * Quantity)
	Items     []LineItem // populated when fetching history
}

var database *sql.DB

/* ===== Init DB (connect, migrate, seed) ===== */

func initDB() {
	db, err := sql.Open("sqlite3", "./db")
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	database = db

	queries := []string{
		`CREATE TABLE IF NOT EXISTS cgo_user (username TEXT, password TEXT)`,
		`CREATE TABLE IF NOT EXISTS cgo_product (name TEXT, price INTEGER, description TEXT)`,
		`CREATE TABLE IF NOT EXISTS cgo_session (token TEXT, user_id INTEGER)`,
		`CREATE TABLE IF NOT EXISTS cgo_cart_item (product_id INTEGER, quantity INTEGER, user_id INTEGER)`,
		/* NEW for checkout/history */
		`CREATE TABLE IF NOT EXISTS cgo_transaction (user_id INTEGER, created_at TEXT)`,
		`CREATE TABLE IF NOT EXISTS cgo_line_item (transaction_id INTEGER, product_id INTEGER, quantity INTEGER)`,
	}
	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			log.Fatal(err)
		}
	}

	// Seed users
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM cgo_user`).Scan(&count); err != nil {
		log.Fatal(err)
	}
	if count == 0 {
		seedUsers := []User{
			{Id: 1, Username: "zagreus", Password: "cerberus"},
			{Id: 2, Username: "melinoe", Password: "b4d3ec1"},
		}
		for _, u := range seedUsers {
			if _, err := db.Exec(`INSERT INTO cgo_user (username, password) VALUES (?, ?)`, u.Username, u.Password); err != nil {
				log.Fatal(err)
			}
		}
	}

	// Seed products
	if err := db.QueryRow(`SELECT COUNT(*) FROM cgo_product`).Scan(&count); err != nil {
		log.Fatal(err)
	}
	if count == 0 {
		seedProducts := []Product{
			{Id: 1, Name: "Americano", Price: 100, Description: "Espresso, diluted for a lighter experience"},
			{Id: 2, Name: "Cappuccino", Price: 110, Description: "Espresso with steamed milk"},
			{Id: 3, Name: "Espresso", Price: 90, Description: "A strong shot of coffee"},
			{Id: 4, Name: "Macchiato", Price: 120, Description: "Espresso with a small amount of milk"},
		}
		for _, p := range seedProducts {
			if _, err := db.Exec(`INSERT INTO cgo_product (name, price, description) VALUES (?, ?, ?)`, p.Name, p.Price, p.Description); err != nil {
				log.Fatal(err)
			}
		}
	}
}

/* ===== Data access helpers ===== */

func getUsers() []User {
	var users []User
	rows, err := database.Query(`SELECT rowid, username, password FROM cgo_user`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Id, &u.Username, &u.Password); err != nil {
			log.Fatal(err)
		}
		users = append(users, u)
	}
	return users
}

func getProducts() []Product {
	var products []Product
	rows, err := database.Query(`SELECT rowid, name, price, description FROM cgo_product`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.Id, &p.Name, &p.Price, &p.Description); err != nil {
			log.Fatal(err)
		}
		products = append(products, p)
	}
	return products
}

func setSession(token string, user User) {
	if _, err := database.Exec(`INSERT INTO cgo_session (token, user_id) VALUES (?, ?)`, token, user.Id); err != nil {
		log.Fatal(err)
	}
}

func getUserFromSessionToken(token string) User {
	var u User
	q := `
SELECT
  cgo_session.user_id,
  cgo_user.username,
  cgo_user.password
FROM cgo_session
JOIN cgo_user ON cgo_user.rowid = cgo_session.user_id
WHERE cgo_session.token = ?
LIMIT 1;
`
	err := database.QueryRow(q, token).Scan(&u.Id, &u.Username, &u.Password)
	if err == sql.ErrNoRows {
		return User{}
	} else if err != nil {
		log.Fatal(err)
	}
	return u
}

/* Cart */

func createCartItem(userId int, productId int, quantity int) {
	if _, err := database.Exec(
		`INSERT INTO cgo_cart_item (user_id, product_id, quantity) VALUES (?, ?, ?)`,
		userId, productId, quantity,
	); err != nil {
		log.Fatal(err)
	}
}

func getCartItemsByUser(user User) []CartItem {
	userId := user.Id
	q := `
SELECT
  cgo_cart_item.rowid,
  cgo_cart_item.user_id,
  cgo_cart_item.product_id,
  cgo_cart_item.quantity,
  cgo_product.name
FROM cgo_cart_item
LEFT JOIN cgo_product ON cgo_cart_item.product_id = cgo_product.rowid
WHERE cgo_cart_item.user_id = ?;
`
	rows, err := database.Query(q, userId)
	if err == sql.ErrNoRows {
		return []CartItem{}
	} else if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var result []CartItem
	for rows.Next() {
		var ci CartItem
		if err := rows.Scan(&ci.Id, &ci.UserId, &ci.ProductId, &ci.Quantity, &ci.ProductName); err != nil {
			log.Fatal(err)
		}
		result = append(result, ci)
	}
	return result
}

/* Checkout → Transaction + Line Items */

func checkoutItemsForUser(user User) {
	// Read current cart
	cartItems := getCartItemsByUser(user)
	if len(cartItems) == 0 {
		return
	}

	// Create a transaction
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := database.Exec(`INSERT INTO cgo_transaction (user_id, created_at) VALUES (?, ?)`, user.Id, now)
	if err != nil {
		log.Fatal(err)
	}
	txID, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}

	// Transform each cart item → line item; then clear cart
	for _, ci := range cartItems {
		if _, err := database.Exec(
			`INSERT INTO cgo_line_item (transaction_id, product_id, quantity) VALUES (?, ?, ?)`,
			txID, ci.ProductId, ci.Quantity,
		); err != nil {
			log.Fatal(err)
		}
		if _, err := database.Exec(`DELETE FROM cgo_cart_item WHERE rowid = ?`, ci.Id); err != nil {
			log.Fatal(err)
		}
	}
}

/* Transaction History */

func getTransactionsByUser(user User) []Transaction {
	var txs []Transaction

	// 1) Fetch transactions
	rows, err := database.Query(`SELECT rowid, user_id, created_at FROM cgo_transaction WHERE user_id = ? ORDER BY rowid DESC`, user.Id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.Id, &t.UserId, &t.CreatedAt); err != nil {
			log.Fatal(err)
		}
		// 2) Fetch line items for this transaction
		items := getLineItemsByTransaction(t.Id)
		t.Items = items

		// 3) Compute total (Price * Quantity)
		sum := 0
		for _, it := range items {
			sum += it.Price * it.Quantity
		}
		t.Total = sum

		txs = append(txs, t)
	}

	return txs
}

func getLineItemsByTransaction(txID int) []LineItem {
	q := `
SELECT
  cgo_line_item.rowid,
  cgo_line_item.transaction_id,
  cgo_line_item.product_id,
  cgo_line_item.quantity,
  cgo_product.name,
  cgo_product.price
FROM cgo_line_item
LEFT JOIN cgo_product ON cgo_product.rowid = cgo_line_item.product_id
WHERE cgo_line_item.transaction_id = ?;
`
	rows, err := database.Query(q, txID)
	if err == sql.ErrNoRows {
		return []LineItem{}
	} else if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var items []LineItem
	for rows.Next() 
		var it LineItem
		if err := rows.Scan(&it.Id, &it.TransactionId, &it.ProductId, &it.Quantity, &it.ProductName, &it.Price); err != nil {
			log.Fatal(err)
		}
		items = append(items, it)
	}
	return items
}
