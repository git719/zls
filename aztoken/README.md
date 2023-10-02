# Azure Token
Sample code for acquiring an Azure token.


## Getting Started
Working Docker and Docker Compose environment. Note that below examples are using version 2.

References:
  - Installing Docker: <https://docs.docker.com/engine/install/rhel/>
  - Docker Compose: <https://docs.docker.com/compose/compose-file/compose-file-v3/>

Docker Socket Issues:
If experience `permissions denied` issues with the docker daemon, try these commands: 

```
sudo usermod -aG docker your_username     # Ubuntu
sudo usermod -aG podman your_username     # RHEL
sudo systemctl restart docker
docker run hello-world                    # To confirm it is fixed
```

Alternatively, try:
```
sudo chown your_username:docker /var/run/docker.sock
sudo chmod 660 /var/run/docker.sock
```

## aztoken.js
Node JS example.

First, make sure you define/export the 3 required environment variables: 

```
export MAZ_TENANT_ID="tenant-id-uuid-string"
export MAZ_CLIENT_ID="your-client-ID-uuid-string"
export MAZ_CLIENT_SECRET="client-secret-string"
```

Then you can build and run for the first time, or run subsequent times.

- `docker compose up --build`: To build and run for the first time.
- `docker compose up`: To run subsequent times.

You can edit `aztoken.js` file to play with different behavior, like using a different scope and so on.


## Other Languages
TODO

