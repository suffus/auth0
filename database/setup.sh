#!/bin/bash

# YubiApp Database Setup Script
# This script creates the database and user for YubiApp

set -e

# Configuration
DB_NAME="yubiapp"
DB_USER="yubiapp"
DB_PASSWORD="T123rain++"  # Change this to a secure password

echo "Setting up YubiApp database..."

# Create database user
echo "Creating database user..."
sudo -u postgres psql -c "CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';" || echo "User already exists or error occurred"

# Create database
echo "Creating database..."
sudo -u postgres psql -c "CREATE DATABASE $DB_NAME OWNER $DB_USER;" || echo "Database already exists or error occurred"

# Grant privileges
echo "Granting privileges..."
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;"

# Run the schema
echo "Running database schema..."
PGPASSWORD=$DB_PASSWORD psql -h localhost -U $DB_USER -d $DB_NAME -f schema.sql

echo "Database setup complete!"
echo ""
echo "Database connection details:"
echo "  Host: localhost"
echo "  Port: 5432"
echo "  Database: $DB_NAME"
echo "  User: $DB_USER"
echo "  Password: $DB_PASSWORD"
echo ""
echo "Update your config.yaml file with these details." 
