#!/bin/bash

# Product Service API Test Script
BASE_URL="http://localhost:8081"

echo "================================"
echo "Product Service API Test"
echo "================================"

# Health Check
echo -e "\n1. Health Check..."
curl -s $BASE_URL/health | jq '.'

# Create Product
echo -e "\n2. Creating a product..."
PRODUCT_ID=$(curl -s -X POST $BASE_URL/api/products \
  -H "Content-Type: application/json" \
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
echo -e "\n3. Creating another product..."
curl -s -X POST $BASE_URL/api/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "iPhone 15 Pro",
    "description": "Latest iPhone with A17 Pro chip",
    "price": 1199.99,
    "stock": 100,
    "category": "Electronics",
    "sku": "IPH15-PRO-001",
    "is_active": true
  }' | jq '.'

# List Products
echo -e "\n4. Listing products..."
curl -s "$BASE_URL/api/products?limit=10&offset=0" | jq '.'

# Get Product by ID
if [ ! -z "$PRODUCT_ID" ]; then
  echo -e "\n5. Getting product by ID: $PRODUCT_ID..."
  curl -s $BASE_URL/api/products/$PRODUCT_ID | jq '.'
  
  # Update Product
  echo -e "\n6. Updating product..."
  curl -s -X PUT $BASE_URL/api/products/$PRODUCT_ID \
    -H "Content-Type: application/json" \
    -d '{
      "name": "MacBook Pro M3 (Updated)",
      "description": "Updated description",
      "price": 2399.99,
      "stock": 50,
      "category": "Electronics",
      "sku": "MBP-M3-001",
      "is_active": true
    }' | jq '.'
  
  # Update Stock
  echo -e "\n7. Updating stock..."
  curl -s -X PATCH $BASE_URL/api/products/$PRODUCT_ID/stock \
    -H "Content-Type: application/json" \
    -d '{
      "stock": 75
    }' | jq '.'
  
  # Get updated product
  echo -e "\n8. Getting updated product..."
  curl -s $BASE_URL/api/products/$PRODUCT_ID | jq '.'
fi

# List by Category
echo -e "\n9. Listing products by category..."
curl -s "$BASE_URL/api/products?category=Electronics&limit=10" | jq '.'

# Metrics
echo -e "\n10. Checking metrics endpoint..."
curl -s $BASE_URL/metrics | head -n 20

echo -e "\n================================"
echo "Test completed!"
echo "================================"

