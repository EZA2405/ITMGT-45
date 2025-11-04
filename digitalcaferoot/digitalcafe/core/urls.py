from django.urls import path
from . import views

urlpatterns = [
    path('', views.index, name='home'),   # ✅ homepage
    path('product/<int:product_id>/', views.product_detail, name='product_detail'),
    path('transactions/', views.transaction_history, name='transaction_history'),  # ✅ ADD THIS LINE
]