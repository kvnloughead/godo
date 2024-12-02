# Prerequisites for Deploying Godo

These instructions are for preparing a server to run a Go application that uses PostgreSQL.

## Minimum Server Requirements

- Ubuntu 22.04 LTS
- 1GB RAM
- 10GB storage
- Open ports:
  - 22 (SSH)
  - 4000 (API)
  - 5432 (PostgreSQL)

## Required Software

### Go Installation

```bash
# Download and install Go
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
```

### PostgreSQL Installation

```bash
# Install PostgreSQL and CLI tools
sudo apt update
sudo apt install postgresql postgresql-contrib

# Verify installation
psql --version
```

### Example: Setting up a GCP Instance

For GCP free tier requirements, use:

- e2-micro
- us-central1-a
- 30gb standard persistent disk

https://cloud.google.com/free/docs/free-cloud-features#compute

1. Create instance. Replace `NAME_INSTANCE` with your desired name.

   ```bash
   gcloud compute instances create NAME_INSTANCE \
       --machine-type=e2-small \
       --zone=us-central1-a \
       --image-family=ubuntu-2204-lts \
       --image-project=ubuntu-os-cloud
   ```

2. Configure firewall:

   ```bash
   # Allow API access
   gcloud compute firewall-rules create allow-api \
       --direction=INGRESS \
       --network=default \
       --action=ALLOW \
       --rules=tcp:4000 \
       --source-ranges=0.0.0.0/0

   # Allow PostgreSQL from your IP only
   gcloud compute firewall-rules create allow-postgres \
       --direction=INGRESS \
       --network=default \
       --action=ALLOW \
       --rules=tcp:5432 \
       --source-ranges=YOUR_IP/32
   ```

See [DEPLOYMENT.md](DEPLOYMENT.md) for next steps.
