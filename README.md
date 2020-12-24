# Wake on LAN Proxy

![](wol-proxy.png)

Send magic packets to your computer without opening ports at your house.

## Requirements

* One server published to the internet
* One machine inside your house (e.g. Raspberry Pi)
* Installed golang or docker in both machine

## Setup

### 1. Configure a remote web application

Please expose web application to the internet using reverse proxy.  
This application uses websocket to communicate with a client.

* passphrase: Required. You can set any string. Longer one is better.
* default mac address: Set to the mac address of your machine if you wish.

#### Local installation

```shell
go get github.com/kyori19/wol-proxy
wol-proxy server <passphrase> [default mac address]
```

#### Using docker

```shell
docker run -p 3000:3000 kyori/wol-proxy server <passphrase> [default mac address]
```

### 2. Configure a local client application

There's no need to expose a local client to the internet.

* passphrase: Required. Must same as the one you configured in a server.
* `-H <address>` (`--host <address>`): Default `localhost:3000`. Address or domain to your web application.
* `-s` (`--secure`): Default `false`. Use TLS for websocket connection.

#### Local installation

```shell
go get github.com/kyori19/wol-proxy
wol-proxy client <passphrase> [--host <address>] [--secure]
```

#### Using docker

```shell
docker run -p 3000:3000 kyori/wol-proxy client <passphrase> [--host <address>] [--secure]
```

### 3. Access to the application

Open this URL in your browser.

```
http(s)://<address to your remote server>/<passphrase>/wsl
```
