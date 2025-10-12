#!/bin/bash

# Create Admin User Script
USER_SERVICE_URL="http://localhost:8080"

echo "================================"
echo "Create Admin User"
echo "================================"

echo -e "\n1. Registering admin user..."
RESPONSE=$(curl -s -X POST $USER_SERVICE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "admin123",
    "full_name": "Admin User"
  }')

echo $RESPONSE | jq '.'

USER_ID=$(echo $RESPONSE | jq -r '.id')

if [ "$USER_ID" != "null" ] && [ ! -z "$USER_ID" ]; then
  echo -e "\n✅ Admin user created with ID: $USER_ID"
  echo ""
  echo "⚠️  IMPORTANT: You need to manually change the role to 'admin' in the database:"
  echo ""
  echo "Run this SQL command:"
  echo "--------------------"
  echo "docker exec -it postgres psql -U postgres -d userdb -c \"UPDATE users SET role='admin' WHERE username='admin';\""
  echo ""
  echo "Or using psql directly:"
  echo "psql -h localhost -U postgres -d userdb"
  echo "UPDATE users SET role='admin' WHERE username='admin';"
  echo ""
  echo "After updating the role, you can use these credentials:"
  echo "Username: admin"
  echo "Password: admin123"
else
  echo -e "\n❌ Failed to create admin user"
  echo "User may already exist or there was an error"
fi

echo -e "\n================================"

