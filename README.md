<img alt="Hera" src="https://s3-us-west-2.amazonaws.com/aschzero-hera/hera.png" width="500px">

### Hera automates the creation of [Argo Tunnels](https://developers.cloudflare.com/argo-tunnel/) to easily and securely expose your local services to the outside world.

Hera lets you instantly access services outside of your local network with a custom domain using tunnels and is a more secure alternative than using port forwarding or dynamic DNS.

Hera monitors the state of your configured services to instantly start a tunnel when the container starts. Tunnel processes are also monitored to ensure persistent connections and to restart them in the event of sudden disconnects or shutdowns. Tunnels are automatically restarted when their containers are restarted, or gracefully shutdown if their containers are stopped.

[![Build Status](https://semaphoreci.com/api/v1/aschzero/hera/branches/master/badge.svg)](https://semaphoreci.com/aschzero/hera)
[![GitHub version](https://badge.fury.io/gh/aschzero%2Fhera.svg)](https://badge.fury.io/gh/aschzero%2Fhera)

----

* [Features](#features)
* [How Hera Works](#how-hera-works)
* [Getting Started](#getting-started)
  * [Prerequisites](#prerequisites)
  * [Obtain a Certificate](#obtain-a-certificate)
  * [Create a Network](#create-a-network)
  * [Container Configuration](#container-configuration)
* [Running Hera](#running-hera)
  * [Required Volumes](#required-volumes)
  * [Persisting Logs](#persisting-logs)
  * [Using Multiple Domains](#using-multiple-domains)
  * [Examples](#examples)
* [Contributing](#contributing)

----

# Features
* Continuously monitors the state of your services for automated tunnel creation.
* Revives tunnels on running containers when Hera is restarted.
* Uses the s6 process supervisor to ensure active tunnel processes are kept alive.
* Low memory footprint and high performance – services can be accessed through a tunnel within seconds.
* Requires a minimal amount of configuration so you can get up and running quickly.
* Supports multiple Cloudflare domains.

# How Hera Works
Hera attaches to the Docker daemon to watch for changes in state of your configured containers. When a new container is started, Hera checks that it has the proper configuration as well as making sure the container can receive connections. If it passes the configuration checks, Hera spawns a new process to create a persistent tunnel connection.

In the event that a container with an active tunnel has been stopped, Hera gracefully shuts down the tunnel process.

ℹ️ Hera only monitors the state of containers that have been explicitly configured for Hera. Otherwise, containers and their events are completely ignored.

# Getting Started
## Prerequisites

* Installation of Docker with a client API version of 1.22 or later
* An active domain in Cloudflare with the Argo Tunnel service enabled
* A valid Cloudflare certificate (see [Obtain a Certificate](#obtain-a-certificate))

## Obtain a Certificate

Hera needs a Cloudflare certificate so it can manage tunnels on your behalf.

1. Download a new certificate by visiting https://www.cloudflare.com/a/warp
2. Rename the certificate to match your domain, ending in `.pem`. For example, a certificate for `mysite.com` should be named `mysite.com.pem`.
3. Move the certificate to a directory that can be mounted as a volume (see [Required Volumes](#required-volumes)).

Hera will look for certificates with names matching your tunnels' hostnames and allows the use of multiple certificates. For more info, see [Using Multiple Domains](#using-multiple-domains).

----

## Create a Network

Hera must be able to connect to your containers and resolve their hostnames before it can create a tunnel. This allows Hera to supply a valid address to Cloudflare during the tunnel creation process.

It is recommended to create a dedicated network for Hera and attach your desired containers to the new network.

For example, to create a network named `hera`:

`docker network create hera`

----

## Container Configuration

Hera utilizes labels for configuration as a way to let you be explicit about which containers you want enabled. There are only two labels that need to be defined:

* `hera.hostname` - The hostname is the address you'll use to request the service outside of your home network. It must be the same as the domain you used to configure your certificate and can either be a root domain or subdomain (e.g.: `mysite.com` or `blog.mysite.com`).

* `hera.port` - The port your service is running on inside the container.

⚠️ _Note: you can still expose a different port to your host network if desired, but the `hera.port` label value needs to be the internal port within the container._

Here's an example of a container configured for Hera with the `docker run` command:

```
docker run \
  --network=hera \
  --label hera.hostname=mysite.com \
  --label hera.port=80 \
  nginx
```

That's it! Assuming Hera is running, you would immediately be able to see the default nginx welcome page when requesting `mysite.com`.

----

# Running Hera

Now we can start Hera by running the following command:

```
docker run \
  --name=hera \
  --network=hera \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /path/to/certs:/root/.cloudflared \
  aschzero/hera:latest
```

## Required Volumes

* `/var/run/docker.sock` – Attaching the Docker daemon as a volume allows Hera to monitor container events.
* `/path/to/certs` – The parent directory of your Cloudflare certificates.

## Persisting Logs

Logs for Hera and active tunnels reside in `/var/log/hera` in addition to being printed to stdout. You can mount a volume to persist the logs on your host machine:

```
docker run \
  --name=hera \
  --network=hera \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /path/to/certs:/root/.cloudflared \
  -v /path/to/logs:/var/log/hera \
  aschzero/hera:latest
```

* The Hera log file can be found at `/var/log/hera/hera.log`
* Tunnel log files are named according to their hostname and can be found at `/var/log/hera/<hostname>.log`

----

## Using Multiple Domains

You can use multiple domains as long as there are certificates for each domain with names matching the base hostname of the tunnel. Names are matched according to the pattern `*.domain.tld` and must be placed in the same directory.

For example, tunnels for `mysite.com` or `blog.mysite.com` will use the certificate named `mysite.com.pem`.

If a certificate with a matching domain cannot be found, it will look for `cert.pem` in the same directory as a fallback.

# Examples

Here are just a couple examples of how a container would be configured for Hera.

An example using Organizr:

```
docker run \
  --name=organizr \
  --network=hera \
  --label=hera.hostname=organizr.mysite.com \
  --label=hera.port=80 \
  lsiocommunity/organizr
```

Another example using Kibana:

```
docker run \
  --name=kibana \
  --network=elkstack \
  --network=hera \
  --label hera.hostname=kibana.mysite.com \
  --label hera.port=5601
  -p 5000:5601 \
  docker.elastic.co/kibana/kibana:6.2.4
```

* Notice that a container can belong to multiple networks as a means of separating concerns.
* The `hera.port` label points to the port inside of the container
* This port mapping would allow you to access Kibana on your local network via port `5000` in addition to `kibana.mysite.com`.

After the container starts you should see similar output to the following in the logs:

```
$ docker logs -f hera

[INFO] Found certificate: mysite.com
[INFO] Hera v0.2.0 has started
[INFO] Hera is listening
[INFO] Registering tunnel kibana.mysite.com @ 172.21.0.3:80
[INFO] Logging to /var/log/hera/kibana.mysite.com.log
INFO[0000] Applied configuration from /etc/services.d/kibana.mysite.com/config.yml
INFO[0000] Proxying tunnel requests to http://172.21.0.3:5601
INFO[0000] Starting metrics server                       addr="127.0.0.1:45603"
INFO[0002] Connected to IAD                              connectionID=0
INFO[0003] Connected to SEA                              connectionID=1
...
```

And just like that, a tunnel is up and running at `kibana.mysite.com`.

A container that stops running will trigger the tunnel to shut down:

```
$ docker stop kibana
$ docker logs -f hera

INFO[0012] Quitting...
INFO[0012] Metrics server stopped
INFO[0043] Initiating graceful shutdown...
[INFO] Stopped tunnel kibana.mysite.com
```

# Contributing

* If you'd like to contribute to the project, refer to the [contributing documentation](https://github.com/aschzero/hera/blob/master/CONTRIBUTING.md).
* Read the [Local Development](https://github.com/aschzero/hera/wiki/Local-Development) wiki for information on how to setup Hera for local development.
