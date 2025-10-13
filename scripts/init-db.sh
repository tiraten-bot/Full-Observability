#!/bin/bash
set -e

# Create multiple databases
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Create userdb if not exists
    SELECT 'CREATE DATABASE userdb'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'userdb')\gexec

    -- Create productdb if not exists
    SELECT 'CREATE DATABASE productdb'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'productdb')\gexec

    -- Create inventorydb if not exists
    SELECT 'CREATE DATABASE inventorydb'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'inventorydb')\gexec

    -- Create paymentdb if not exists
    SELECT 'CREATE DATABASE paymentdb'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'paymentdb')\gexec

    -- Grant privileges
    GRANT ALL PRIVILEGES ON DATABASE userdb TO postgres;
    GRANT ALL PRIVILEGES ON DATABASE productdb TO postgres;
    GRANT ALL PRIVILEGES ON DATABASE inventorydb TO postgres;
    GRANT ALL PRIVILEGES ON DATABASE paymentdb TO postgres;
EOSQL

echo "Databases created successfully!"

