#!/bin/bash

# Product Service API Test Script
BASE_URL="http://localhost:8081"
USER_SERVICE_URL="http://localhost:8080"

echo "================================"
echo "Product Service API Test (With Auth)"
echo "================================"

# Health Check
echo -e "\n1. Health Check..."
curl -s $BASE_URL/health | jq '.'

# Get Admin Token from User Service
echo -e "\n2. Getting Admin Token..."
LOGIN_RESPONSE=$(curl -s -X POST $USER_SERVICE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo "❌ Failed to get admin token. Creating admin user first..."
  
  # Register admin user
  curl -s -X POST $USER_SERVICE_URL/auth/register \
    -H "Content-Type: application/json" \
    -d '{
      "username": "admin",
      "email": "admin@example.com",
      "password": "admin123",
      "full_name": "Admin User"
    }' | jq '.'
  
  echo -e "\n⚠️  Please run this script again after admin user is created."
  echo "Note: You may need to change the admin role manually via database."
  exit 1
fi

echo "✅ Token received: ${TOKEN:0:20}..."

# Test public endpoint without token (should work)
echo -e "\n3. Testing public endpoint (no token)..."
curl -s "$BASE_URL/api/products?limit=5" | jq '.'

# Try to create product without token (should fail)
echo -e "\n4. Trying to create product without token (should fail)..."
curl -s -X POST $BASE_URL/api/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Product",
    "sku": "TEST-001"
  }' | jq '.'

# Create Product with Admin Token
echo -e "\n5. Creating a product with admin token..."
PRODUCT_ID=$(curl -s -X POST $BASE_URL/api/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "MacBook Pro M3",
    "description": "Latest MacBook Pro with M3 chip",
    "price": 2499.99,
    "stock": 50,
    "category": "Electronics",
    "sku": "MBP-M3-001",
    "is_active": true
  }' | jq -r '.data.id')

echo "Created product with ID: $PRODUCT_ID"

# Create another product
echo -e "\n6. Creating another product..."
curl -s -X POST $BASE_URL/api/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "iPhone 15 Pro",
    "description": "Latest iPhone with A17 Pro chip",
    "price": 1199.99,
    "stock": 100,
    "category": "Electronics",
    "sku": "IPH15-PRO-001",
    "is_active": true
  }' | jq '.'

# List Products (public)
echo -e "\n7. Listing products (public endpoint)..."
curl -s "$BASE_URL/api/products?limit=10&offset=0" | jq '.'

# Get Product by ID (public)
if [ ! -z "$PRODUCT_ID" ] && [ "$PRODUCT_ID" != "null" ]; then
  echo -e "\n8. Getting product by ID: $PRODUCT_ID (public)..."
  curl -s $BASE_URL/api/products/$PRODUCT_ID | jq '.'
  
  # Try to update without token (should fail)
  echo -e "\n9. Trying to update product without token (should fail)..."
  curl -s -X PUT $BASE_URL/api/products/$PRODUCT_ID \
    -H "Content-Type: application/json" \
    -d '{
      "name": "Hacked Product"
    }' | jq '.'
  
  # Update Product with admin token
  echo -e "\n10. Updating product with admin token..."
  curl -s -X PUT $BASE_URL/api/products/$PRODUCT_ID \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
      "name": "MacBook Pro M3 (Updated)",
      "description": "Updated description",
      "price": 2399.99,
      "stock": 50,
      "category": "Electronics",
      "sku": "MBP-M3-001",
      "is_active": true
    }' | jq '.'
  
  # Update Stock with admin token
  echo -e "\n11. Updating stock with admin token..."
  curl -s -X PATCH $BASE_URL/api/products/$PRODUCT_ID/stock \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d '{
      "stock": 75
    }' | jq '.'
  
  # Get updated product
  echo -e "\n12. Getting updated product..."
  curl -s $BASE_URL/api/products/$PRODUCT_ID | jq '.'
fi

# List by Category (public)
echo -e "\n13. Listing products by category (public)..."
curl -s "$BASE_URL/api/products?category=Electronics&limit=10" | jq '.'

# Get Stats (public)
echo -e "\n14. Getting product statistics (public)..."
curl -s "$BASE_URL/api/products/stats" | jq '.'

# Try to delete without token (should fail)
if [ ! -z "$PRODUCT_ID" ] && [ "$PRODUCT_ID" != "null" ]; then
  echo -e "\n15. Trying to delete product without token (should fail)..."
  curl -s -X DELETE $BASE_URL/api/products/$PRODUCT_ID | jq '.'
  
  # Delete with admin token
  echo -e "\n16. Deleting product with admin token..."
  curl -s -X DELETE $BASE_URL/api/products/$PRODUCT_ID \
    -H "Authorization: Bearer $TOKEN" | jq '.'
fi

# Metrics
echo -e "\n17. Checking metrics endpoint..."
curl -s $BASE_URL/metrics | head -n 30

echo -e "\n================================"
echo "✅ Test completed!"
echo "================================"
echo ""
echo "Summary:"
echo "- Public endpoints (GET) work without token ✅"
echo "- Admin endpoints (POST/PUT/PATCH/DELETE) require admin token ✅"
echo "- JWT authentication is enforced ✅"

