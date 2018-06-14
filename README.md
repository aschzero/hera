<img alt="Hera" src="https://s3-us-west-2.amazonaws.com/aschzero-hera/hera.png" width="500px">

### Hera automates the creation of [Argo Tunnels](https://developers.cloudflare.com/argo-tunnel/) to easily and securely expose your local services to the outside world.

Hera is useful for those who run multiple Dockerized services on their home network and want to access them anywhere without the use of port forwarding or other potentially insecure methods. Hera can also be used to manage tunnels as a VPN replacement or to expose local development environments.

Hera monitors the state of your configured services to instantly start a tunnel when the container starts. Tunnel processes are also monitored to ensure persistent connections and to restart them in the event of sudden disconnects or shutdowns. Hera also handles graceful shutdown of active tunnels should their respective containers stop running.

[![Build Status](https://semaphoreci.com/api/v1/aschzero/hera/branches/master/badge.svg)](https://semaphoreci.com/aschzero/hera)
[![](https://images.microbadger.com/badges/version/aschzero/hera.svg)](https://microbadger.com/images/aschzero/hera "Get your own version badge on microbadger.com")
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/aschzero/hera/blob/master/LICENSE)

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
  * [Examples](#examples)
* [Contributing](#contributing)

----

# Features
* Continuously monitors the state of your services for automated tunnel creation.
* Revives tunnels on running containers when Hera is restarted.
* Uses the s6 process supervisor to ensure active tunnel processes are kept alive.
* Low memory footprint and high performance – services can be accessed through a tunnel within seconds.
* Requires a minimal amount of configuration so you can get up and running quickly.
* Extensive logging for all tunnel activity.

# How Hera Works
Hera attaches to the Docker daemon to watch for changes in state of your configured containers. When a new container is started, Hera checks that it has the proper configuration as well as making sure the container can receive connections. If it passes the configuration checks, Hera spawns a new process to create a persistent tunnel connection.

In the event that a container with an active tunnel has been stopped, Hera gracefully shuts down the tunnel process.

ℹ️ Hera only monitors the state of containers that have been explicitly configured for Hera. Otherwise, containers and their events are completely ignored (see [Configuring Containers](#configuring-containers)).

# Getting Started
## Prerequisites

* Installation of Docker with a client API version of 1.22 or later
* An active domain in Cloudflare with the Argo Tunnel service enabled
* A valid Cloudflare certificate (see [Obtain a Certificate](#obtain-a-certificate))

## Obtain a Certificate

Hera will need a Cloudflare certificate so it can manage tunnels on your behalf.

1. Download a new certificate by visiting https://www.cloudflare.com/a/warp
2. Rename the file to `cert.pem` if it is not already
3. Move the certificate to a directory that can be mounted as a volume (see [Required Volumes](#required-volumes)).

----

## Create a Network

Hera must be able to connect to your containers and resolve their hostnames before it can create a tunnel. This allows Hera to supply a valid address to Cloudflare during the tunnel creation process.

It is recommended to create a dedicated network for Hera and attach your desired containers to the new network.

For example, to create a network named `hera`:

`docker network create hera`

----

## Container Configuration

Hera utilizes labels for configuration as a way to let you be explicit about which containers you want enabled. There are only two labels that need to be defined:

* `hera.hostname` - The hostname is the address you'll use to request the service outside of your home network. It must be the same as the domain you used to configure your certificate and must match either `mydomain.com` or `*.mydomain.com`.

* `hera.port` - The port your service is running on inside the container.

⚠️ _Note: you can still expose a different port to your host network if desired, but the `hera.port` label value needs to be the internal port within the container._

Here's an example of a container configured for Hera with the `docker run` command:

```
docker run \
  --network=hera \
  --label hera.hostname=my.domain.com \
  --label hera.port=80 \
  nginx
```

That's it! Assuming Hera is running, you would immediately be able to see the default nginx welcome page when requesting `my.domain.com`.

----

# Running Hera

Now we can start Hera by running the following command:

```
docker run \
  --name=hera \
  --network=hera \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /path/to/cert:/root/.cloudflared \
  aschzero/hera:latest
```

## Required Volumes

* `/var/run/docker.sock` – Attaching the Docker daemon as a volume allows Hera to monitor container events.
* `/path/to/cert` – Replace with the path of the parent directory of your obtained certificate. This allows Cloudflare to authenticate you during tunnel creation.

## Persisting Logs

Logs for Hera and active tunnels reside in `/var/log/hera` in addition to being printed to stdout. You can mount a volume to persist the logs on your host machine:

```
docker run \
  --name=hera \
  --network=hera \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /path/to/cert:/root/.cloudflared \
  -v /path/to/logs:/var/log/hera \
  aschzero/hera:latest
```

* The Hera log file can be found at `/var/log/hera/hera.log`
* Tunnel log files are named according to their hostname and can be found at `/var/log/hera/<hostname>.log`

----

## Examples

Here are just a couple examples of how a container would be configured for Hera.

An example using Organizr:

```
docker run \
  --name=organizr \
  --network=hera \
  --label=hera.hostname=organizr.domain.com \
  --label=hera.port=80 \
  lsiocommunity/organizr
```

Another example using Kibana:

```
docker run \
  --name=kibana \
  --network=elkstack \
  --network=hera \
  --label hera.hostname=kibana.domain.com \
  --label hera.port=5601
  -p 5000:5601 \
  docker.elastic.co/kibana/kibana:6.2.4
```

* Notice that a container can belong to multiple networks as a means of separating concerns.
* The `hera.port` label points to the port inside of the container
* This port mapping would allow you to access Kibana on your local network via port `5000` in addition to `kibana.domain.com`.

After the container starts you should see similar output to the following in the logs:

```
$ docker logs -f hera

[INFO] Hera v0.1.0 has started
[INFO] Hera is listening
[INFO] Registering tunnel kibana.domain.com @ 172.21.0.3:80
[INFO] Logging to /var/log/hera/kibana.domain.com.log
INFO[0000] Applied configuration from /etc/services.d/kibana.domain.com/config.yml
INFO[0000] Proxying tunnel requests to http://172.21.0.3:5601
INFO[0000] Starting metrics server                       addr="127.0.0.1:45603"
INFO[0002] Connected to IAD                              connectionID=0
INFO[0003] Connected to SEA                              connectionID=1
...
```

And just like that, a tunnel is up and running at `kibana.domain.com`.

A container that stops running will trigger the tunnel to shut down:

```
$ docker stop kibana
$ docker logs -f hera

INFO[0012] Quitting...
INFO[0012] Metrics server stopped
INFO[0043] Initiating graceful shutdown...
[INFO] Stopped tunnel kibana.domain.com
```

# Contributing

* If you'd like to contribute to the project, refer to the [contributing documentation](https://github.com/aschzero/hera/blob/master/CONTRIBUTING.md).
* Read the [Local Development](https://github.com/aschzero/hera/wiki/Local-Development) wiki for information on how to setup Hera for local development.
