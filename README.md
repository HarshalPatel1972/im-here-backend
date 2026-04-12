# I'm Here

A free, open-source guardian for developers. We never display your keys publicly.

## Setup

1. Copy `.env.example` to `.env` and fill in the values.
2. Run `make run` to start the scanner.
3. Run `make build` to build the binary.
4. Run `make build-linux` to build the Linux binary for Oracle VM deployment.
5. Run `make lint` to lint the code.

## Oracle Cloud Deployment

This project operates seamlessly on an Oracle Cloud Always Free VM (Ubuntu 22.04). Follow these steps to ensure continuous secret scanning:

1. **Deploying via automation:**
    - Provide your OCI Server credentials inside `deploy.sh`.
    - Run `make deploy` to safely send and start the process.

2. **Manual VM Configuration Steps:**
    - Provision an Always Free Oracle AMD VM (running Ubuntu 22.04).
    - Open any required inbound ports in your OCI Security Lists.
    - SSH into your VM and install Git and Go (`sudo apt update && sudo snap install go --classic`).
    - Clone this Repo and populate your `.env`.
    - From the CLI, copy the `.service` instance, then reload the systemctl daemon (`sudo systemctl daemon-reload && sudo systemctl enable im-here && sudo systemctl start im-here`).
    
3. **Validate:**
    - Check the service: `systemctl status im-here`
    - Stream live logs: `journalctl -u im-here -f`

**Note:** This project never stores or displays actual secret values.
