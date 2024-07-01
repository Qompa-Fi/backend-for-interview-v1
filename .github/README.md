
# Backend for Technical interview V1

## Deployment

```bash
# In the remote host
sudo mkdir -p /app

# In your local machine
scp build/server interview-7pmwvjf9.inversiones.io:/home/ec2-user/server

# In the remote host again...
sudo mv /home/ec2-user/server /app/server

cat <<EOF | sudo tee /etc/systemd/system/backend.service
[Unit]
Description=Backend for technical interview V1

[Service]
ExecStart=/app/server

Environment="PORT=80"
Environment="API_KEYS=***,***,***"
Environment="TZ=America/Lima"

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable backend
sudo systemctl start backend
```
