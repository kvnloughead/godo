# Deploying Godo

## Prerequisites

Your server needs:

- Ubuntu 22.04 or later
- PostgreSQL 14 or later
- Open ports:
  - 4000 (API)
  - 5432 (PostgreSQL)
- Systemd for service management
- User with sudo privileges

See [PREREQUISITES.md](PREREQUISITES.md) for more details.

## Database Setup

1. Configure PostgreSQL for remote connections:

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

   Comment out:

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
   host    godo_db         godo_user        YOUR_IP/32              scram-sha-256
   ```

   Restart PostgreSQL:

   ```bash
   sudo systemctl restart postgresql
   ```

## Application Setup

1. Create application directory:

   ```bash
   sudo mkdir -p /opt/godo
   sudo chown $USER:$USER /opt/godo
   ```

2. Create systemd service:

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
   User=$USER
   Group=$USER
   WorkingDirectory=/opt/godo
   ExecStart=/opt/godo/godo-linux-amd64 \
       -port=4000 \
       -env=production \
       -db-dsn="postgresql://godo_user:your_password@localhost/godo_db?sslmode=disable"

   Restart=always
   RestartSec=5

   [Install]
   WantedBy=multi-user.target
   ```

## Local Development Setup

1. Create `.env.production`:

   ```env
   DB_NAME=godo_db
   DB_USER=godo_user
   DB_PASSWORD=your_password
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

   # Deploy binary
   make deploy/gcp
   ```

## Service Management

```bash
# Start and enable service
sudo systemctl enable --now godo

# Check status
sudo systemctl status godo

# View logs
sudo journalctl -u godo -n 100 --reverse # Last 100 lines
sudo journalctl -u godo -f &    # Follow in background
sudo journalctl -u godo -f    # Follow in foreground
```

## Domain Setup

1. Install Nginx:

```bash
sudo apt update
sudo apt install nginx
```

2. Create Nginx configuration:

```bash
sudo nano /etc/nginx/sites-available/godo
```

Add:

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        # Forward requests to your Go API running on port 4000
        proxy_pass http://localhost:4000;

        # Preserve the original host, port, and protocol headers from the client
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-Port $server_port;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Preserve the original client IP, even if behind another proxy.
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;

        # Timeouts
        proxy_connect_timeout 10s;
        proxy_send_timeout 15s;
        proxy_read_timeout 60s;
    }
}
```

3. Enable the site:

```bash
sudo ln -s /etc/nginx/sites-available/godo /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

4. Install certbot for SSL:

Follow the instructions [here](https://certbot.eff.org) for NGINX and Linux
(snap). Then run:

```bash
# Prepare certbot command
sudo ln -s /snap/bin/certbot /usr/bin/certbot

# Run certbot and follow the instructions
sudo certbot --nginx
```

5. Update firewall rules:

```bash
# Allow HTTP/HTTPS
gcloud compute firewall-rules create allow-web \
    --direction=INGRESS \
    --network=default \
    --action=ALLOW \
    --rules=tcp:80,tcp:443 \
    --source-ranges=0.0.0.0/0
```

6. Configure Network Tags

Your VM instance needs the correct network tags to allow HTTP traffic:

```bash
# Add http-server tag to allow HTTP traffic
gcloud compute instances add-tags YOUR_INSTANCE_NAME \
    --tags=http-server \
    --zone=YOUR_ZONE
```

The `http-server` tag is required for the default HTTP firewall rule to take effect.
