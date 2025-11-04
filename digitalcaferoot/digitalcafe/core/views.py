from django.http import HttpResponse
from django.template import loader
from django.shortcuts import get_object_or_404, redirect
from django.contrib.auth.decorators import login_required
from django.contrib.auth import login, logout, authenticate
from django.contrib import messages
import datetime as dt

from .models import Product, CartItem, Transaction, LineItem

@login_required
def index(request):
    template = loader.get_template("core/index.html")
    products = Product.objects.all()
    
    # Get cart item count for the current user
    cart_count = CartItem.objects.filter(user=request.user).count()
    
    context = {
        "user": request.user,
        "product_data": products,
        "cart_count": cart_count,
    }
    return HttpResponse(template.render(context, request))


@login_required
def product_detail(request, product_id):
    product = get_object_or_404(Product, id=product_id)
    
    if request.method == 'GET':
        template = loader.get_template("core/products_detail.html")
        context = {
            "product": product
        }
        return HttpResponse(template.render(context, request))

    elif request.method == 'POST':
        submitted_quantity = int(request.POST['quantity'])
        user = request.user

        # Check if cart item already exists for this user and product
        cart_item, created = CartItem.objects.get_or_create(
            user=user,
            product=product,
            defaults={'quantity': submitted_quantity}
        )
        if not created:
            cart_item.quantity += submitted_quantity
            cart_item.save()

        messages.add_message(
            request,
            messages.INFO,
            f'Added {submitted_quantity} of {product.name} to your cart.'
        )
        return redirect('home')


def login_view(request):
    if request.method == 'GET':
        template = loader.get_template("core/login_view.html")
        return HttpResponse(template.render({}, request))

    elif request.method == 'POST':
        submitted_username = request.POST['username']
        submitted_password = request.POST['password']
        user_object = authenticate(
            username=submitted_username,
            password=submitted_password
        )
        if user_object is None:
            messages.add_message(request, messages.INFO, 'Invalid login.')
            return redirect(request.path_info)
        login(request, user_object)
        return redirect('home')

    
@login_required
def checkout(request):
    if request.method == 'GET':
        template = loader.get_template("core/checkout.html")
        cart_items = CartItem.objects.filter(user=request.user)
        
        # Calculate total
        total = sum(item.product.price * item.quantity for item in cart_items)
        
        context = {
            'cart_items': cart_items,
            'total': total,
        }
        return HttpResponse(template.render(context, request))
    
    elif request.method == 'POST':
        # Handle checkout submission
        cart_items = CartItem.objects.filter(user=request.user)
        
        # Check if cart is empty
        if not cart_items.exists():
            messages.warning(request, 'Your cart is empty.')
            return redirect('home')
        
        # Create transaction
        transaction = Transaction.objects.create(
            user=request.user,
            created_at=dt.datetime.now()
        )
        
        # Create line items from cart
        for cart_item in cart_items:
            LineItem.objects.create(
                transaction=transaction,
                product=cart_item.product,
                quantity=cart_item.quantity
            )
        
        # Clear the cart
        cart_items.delete()
        
        messages.success(request, f'Order #{transaction.id} placed successfully!')
        return redirect('home')
    
@login_required
def transaction_history(request):
    # Get all transactions for the current user, ordered by most recent first
    transactions = Transaction.objects.filter(user=request.user).order_by('-created_at')
    
    template = loader.get_template("core/transaction_history.html")
    context = {
        'transactions': transactions,
    }
    return HttpResponse(template.render(context, request))