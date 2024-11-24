# Deployment Guide

## Server Setup

1. Install Go:

   ```bash
   wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
   sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz

   # Add Go to PATH
   echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
   source ~/.bashrc
   ```

2. Install and configure PostgreSQL:

   ```bash
   sudo apt update
   sudo apt install postgresql postgresql-contrib
   ```

3. Configure PostgreSQL for remote connections:

   Edit postgresql.conf:

   ```bash
   sudo nano /etc/postgresql/*/main/postgresql.conf
   ```

   Add:

   ```ini
   listen_addresses = '*'
   ```

   Edit pg_hba.conf:

   ```bash
   sudo nano /etc/postgresql/*/main/pg_hba.conf
   ```

   Comment out the line:

   ```
   # local   all             postgres                                peer
   ```

   Add:

   ```
   # Allow postgres superuser access
   local   all             postgres                                trust
   host    all             postgres        127.0.0.1/32            trust
   host    all             postgres        ::1/128                 trust

   # Allow remote connections from your development machine
   host    godo_db         godo_user        YOUR_LOCAL_IP/32        scram-sha-256
   ```

4. Restart PostgreSQL:
   ```bash
   sudo systemctl restart postgresql
   ```

## Application Setup

1. Create application directory:

   ```bash
   sudo mkdir -p /opt/godo
   sudo chown $USER:$USER /opt/godo
   ```

2. Copy binary to server:

   ```bash
   scp build/godo-linux-amd64 your_server:/opt/godo/
   ```

3. Create systemd service:

   ```bash
   sudo nano /etc/systemd/system/godo.service
   ```

   Add:

   ```ini
   [Unit]
   Description=Godo API Service
   After=network.target postgresql.service

   [Service]
   Type=simple
   Environment=ENV=production
   User=kevin
   Group=kevin
   WorkingDirectory=/opt/godo
   ExecStart=/opt/godo/godo-linux-amd64 \
       -port=4000 \
       -env=production \
       -db-dsn="postgresql://godo_user:your_secure_password@localhost/godo_db?sslmode=disable"

   Restart=always
   RestartSec=5

   [Install]
   WantedBy=multi-user.target
   ```

## Database Management

All database operations are performed from your local development machine. The server only needs PostgreSQL configured to accept remote connections.

1. Create `.env.production` on your local machine:

   ```env
   DB_NAME=godo_db
   DB_USER=godo_user
   DB_PASSWORD=your_secure_password
   DB_HOST=your_server_ip
   DB_PORT=5432
   DB_DSN=postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable
   ```

2. Run database operations from your local machine:

   ```bash
   # Setup database
   ENV=production make db/setup

   # Run migrations
   ENV=production make db/migrations/up

   # Verify connection
   ENV=production make db/psql
   ```

## Service Management

```bash
# Start the service
sudo systemctl start godo

# Enable service on boot
sudo systemctl enable godo

# Check service status
sudo systemctl status godo

# View logs
sudo journalctl -u godo -f
```
