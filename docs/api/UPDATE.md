# Updating the Application

## Updating the Service Configuration

1. Make changes to your local service file
2. Copy the updated service file to the server:
   ```bash
   make deploy/copy FILE=path/to/godo.service DEST=/tmp/
   ```

3. On the server, move the file and reload:
   ```bash
   sudo mv /tmp/godo.service /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl restart godo
   ```

4. Verify the service is running:
   ```bash
   sudo systemctl status godo
   ```

## Updating the Application Binary

There are two ways to update the application binary:

### Method 1: Using the deploy/gcp command

This is the recommended method as it handles everything automatically:

```bash
make deploy/gcp
```

This command will:
1. Build the Linux binary
2. Stop the service
3. Copy the new binary
4. Start the service

### Method 2: Manual Update

If you need more control over the update process:

1. Build the Linux binary:
   ```bash
   make build/linux
   ```

2. Copy the binary to the server:
   ```bash
   make deploy/copy FILE=build/release/godo-linux-amd64 DEST=/opt/godo/
   ```

3. SSH into the server:
   ```bash
   make deploy/ssh
   ```

4. Restart the service:
   ```bash
   sudo systemctl restart godo
   ```

5. Verify the update:
   ```bash
   sudo systemctl status godo
   sudo journalctl -u godo -n 50 --no-pager
   ```

## Troubleshooting

If you encounter issues after an update:

1. Check the service logs:
   ```bash
   sudo journalctl -u godo -f
   ```

2. Verify the service configuration:
   ```bash
   sudo systemctl cat godo
   ```

3. Roll back if necessary:
   - Keep previous versions of the binary with date suffixes
   - Use symlinks to switch between versions
   - Example:
     ```bash
     sudo mv /opt/godo/godo-linux-amd64 /opt/godo/godo-linux-amd64.$(date +%Y%m%d)
     sudo ln -sf /opt/godo/godo-linux-amd64.20240315 /opt/godo/godo-linux-amd64
     sudo systemctl restart godo
     ```