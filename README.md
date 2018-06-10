<img alt="Hera" src="https://s3-us-west-2.amazonaws.com/aschaper-hera/hera.png" width="500px">

### Hera automates the creation of [Argo Tunnels](https://developers.cloudflare.com/argo-tunnel/) to easily and securely expose your local services to the outside world.

Hera is useful for those who run multiple Dockerized services on their home network and want to access them anywhere without the use of port forwarding or other potentially insecure methods. Hera can also be used to manage tunnels as a VPN replacement or to expose local development environments.

Hera monitors the state of your configured services to instantly start a tunnel when the container starts. Tunnel processes are also monitored to ensure persistent connections and to restart them in the event of sudden disconnects or shutdowns. Hera also handles graceful shutdown of active tunnels should their respective containers stop running.

[![Build Status](https://semaphoreci.com/api/v1/aschaper/hera/branches/master/badge.svg)](https://semaphoreci.com/aschaper/hera)
[![](https://images.microbadger.com/badges/version/aschaper/hera.svg)](https://microbadger.com/images/aschaper/hera "Get your own version badge on microbadger.com")
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/aschaper/hera/blob/master/LICENSE)

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
  * [Example](#example)

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
2. Rename the file to `cert.pem`
3. Move the certificate to a directory that can be mounted as a volume (see [Required Volumes](#required-volumes)).

----

## Create a Network

Hera must be able to connect to your containers before it can create a tunnel. During the configuration verification process, Hera checks to ensure it can successfully resolve the hostname of the container. This allows Hera to supply valid addresses to Cloudflare during the tunnel creation process.

It is recommended to create a dedicated network for Hera and attach your desired containers to the new network.

For example, to create a network named `hera`:

`docker network create hera`

----

## Container Configuration

Hera utilizes labels for configuration as a way to let you be explicit about which containers you want enabled for Hera. There are only two labels that need to be defined:

* `hera.hostname` - The hostname is the address you'll use to request the service outside of your home network. This is typically a domain that you own and can either be a domain or subdomain (e.g.: either `mydomain.com` or `coolservice.mydomain.com`).

* `hera.port` - The port your service is running on inside the container.

⚠️ _Note that you can still expose a different port to your host network if desired, but the `hera.port` label value needs to be the internal port within the container._

Here's an example of a container configured for Hera with the `docker run` command:

```
docker run \
  --network=hera \
  -p 8080:80 \
  --label hera.hostname=my.domain.com \
  --label hera.port=80 \
  nginx
```

That's it! If the Hera container were to be running while the above command is executed, a tunnel would be created and accessible via `my.domain.com`.

----

# Running Hera

Now we can start Hera by running the following command:

```
docker run \
  --name=hera \
  --network=hera \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /path/to/cert:/root/.cloudflared \
  aschaper/hera:latest
```

## Required Volumes

* `/var/run/docker.sock` – Attaching the Docker daemon as a volume allows Hera to monitor container events.
* `/path/to/cert` – Allows Cloudflare to authenticate you during tunnel creation. Replace with the path of the certificate obtained earlier.

## Persisting Logs

Logs for Hera itself and active tunnels are sent to both stdout and their own log files located at `/var/log/hera`. You can mount a volume to persist the logs on your host machine:

```
docker run \
  --name=hera \
  --network=hera \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /path/to/cert:/root/.cloudflared \
  -v /path/to/logs:/var/log/hera \
  aschaper/hera:latest
```

* The Hera log file can be found at `/var/log/hera/hera.log`
* Tunnel log files are named by their hostname and can be found at `/var/log/hera/<tunnel hostname>.log`

----

## Example

Now that Hera is up and running, let's run a small nginx container with the proper Hera configuration:

```
docker run \
  --network=hera \
  -p 8080:80 \
  --label hera.hostname=my.domain.com \
  --label hera.port=80 \
  nginx
```

After the container starts you should see similar output to the following in the logs:

```
$ docker logs -f hera

[INFO] Hera v0.1.0 has started
[INFO] Hera is listening
[INFO] Registering tunnel my.domain.com @ 172.21.0.3:80
[INFO] Logging to /var/log/hera/my.domain.com.log
INFO[0000] Applied configuration from /etc/services.d/my.domain.com/config.yml
INFO[0000] Proxying tunnel requests to http://172.21.0.3:80
INFO[0000] Starting metrics server                       addr="127.0.0.1:45603"
INFO[0002] Connected to IAD                              connectionID=0
INFO[0003] Connected to SEA                              connectionID=1
...
```

And just like that, a tunnel is up and running accessible at `my.domain.com`.

Now we'll stop the nginx container which will cause the tunnel to shut down:

```
$ docker stop nginx
$ docker logs -f hera

INFO[0012] Quitting...
INFO[0012] Metrics server stopped
INFO[0043] Initiating graceful shutdown...
[INFO] Stopped tunnel my.domain.com
```
