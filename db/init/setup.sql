-- Set script for creating database and user. Run once per environment.

-- Create database if it doesn't exist
SELECT 'CREATE DATABASE ' || :'db_name'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = :'db_name')\gexec

-- Create user if it doesn't exist
CREATE USER :"db_user" WITH PASSWORD :'db_password';

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE :"db_name" TO :"db_user";
