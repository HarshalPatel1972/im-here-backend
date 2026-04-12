#!/bin/bash
set -e

# Update these variables for your Oracle Cloud VM
VM_HOST="your-vm-ip-or-hostname"
VM_USER="ubuntu"
DEST_DIR="/home/ubuntu/im-here"

echo "Building Linux binary..."
make build-linux

echo "Creating remote directory..."
ssh ${VM_USER}@${VM_HOST} "mkdir -p ${DEST_DIR}"

echo "Copying binary and service file..."
scp bin/im-here-linux ${VM_USER}@${VM_HOST}:${DEST_DIR}/im-here
scp im-here.service ${VM_USER}@${VM_HOST}:${DEST_DIR}/

echo "Deploying systemd service..."
ssh ${VM_USER}@${VM_HOST} "sudo cp ${DEST_DIR}/im-here.service /etc/systemd/system/im-here.service && \
sudo systemctl daemon-reload && \
sudo systemctl enable im-here && \
sudo systemctl restart im-here"

echo "Deployment complete! Check status with:"
echo "ssh ${VM_USER}@${VM_HOST} 'systemctl status im-here'"
