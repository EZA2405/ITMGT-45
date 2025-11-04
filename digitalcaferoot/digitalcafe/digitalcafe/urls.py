from django.contrib import admin
from django.urls import path, include
from core import views   # ✅ import views from core app, not from digitalcafe

urlpatterns = [
    path('admin/', admin.site.urls),
    path('', include('core.urls')),   # ✅ core app handles main pages
    path('accounts/login/', views.login_view, name='login_view'),  # ✅ login view in core/views.py
    path("checkout", views.checkout, name="checkout"),
]
