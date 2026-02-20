# IBKR Client Portal Gateway Container

Run the [Interactive Brokers Client Portal Gateway](https://www.interactivebrokers.com/campus/ibkr-api-page/cpapi-v1/) in a container using Podman or Docker.

## Prerequisites

- [Podman](https://podman.io/) or [Docker](https://www.docker.com/) installed
- An active [Interactive Brokers](https://www.interactivebrokers.com/) account

## Quick Start

```bash
# Build the container image
./gateway.sh build

# Start the gateway
./gateway.sh run

# Open the authentication page
xdg-open https://localhost:5000
```

Log in with your IBKR credentials. The gateway session remains active until you stop the container or the session times out.

## Commands

| Command | Description |
|---------|-------------|
| `./gateway.sh build` | Build the container image |
| `./gateway.sh run` | Start the gateway (detached) |
| `./gateway.sh stop` | Stop and remove the container |
| `./gateway.sh logs` | Follow container logs |
| `./gateway.sh status` | Show container status |

## Configuration

The gateway is configured via `conf.yaml`, which is bind-mounted into the container at runtime. Edit it without rebuilding:

```yaml
# conf.yaml — key settings
listenPort: 5000          # Internal port (map externally via HOST_PORT)
listenSsl: true           # HTTPS enabled by default

ips:
  allow:
    - 127.0.0.1           # Localhost
    - 172.*               # Container default network
    - 10.*                # Common private range
    - 192.168.*           # Common private range
```

### Custom Port

Set the `IBKR_CPGW_PORT` environment variable to change the host-side port:

```bash
IBKR_CPGW_PORT=8443 ./gateway.sh run
```

### Container Runtime

The script auto-detects `podman` (preferred) or `docker`. Override with:

```bash
CONTAINER_RUNTIME=docker ./gateway.sh run
```

> **Note:** Do not change `listenPort` in `conf.yaml`. Use the environment variable to remap the host port instead.

## Architecture

```
Host                          Container
─────────────────────────────────────────────
localhost:5000  ──►  5000/tcp  (gateway)
                     ▲
conf.yaml (bind mount, read-only)
```

The Containerfile uses a multi-stage build:
1. **Build stage** — downloads and extracts the gateway zip from IBKR
2. **Runtime stage** — minimal `eclipse-temurin:21-jre` image, runs as non-root `gateway` user

## Security Notes

- The default `conf.yaml` IP allowlist (`172.*`, `10.*`, `192.168.*`) is permissive for ease of use. For production, tighten these to match your actual network ranges (e.g., `172.16.*` through `172.31.*` for RFC 1918).
- The gateway uses a self-signed certificate (`vertx.jks`) shipped by IBKR. Replace it with your own for production use.
- Never commit IBKR credentials to source control.

## Troubleshooting

**"Container already exists"** — Run `./gateway.sh stop` first, then `./gateway.sh run`.

**Cannot connect** — Verify the gateway is running with `./gateway.sh status`. Check that your IP is in the `ips.allow` list in `conf.yaml`.

**SSL certificate warning** — Expected. The gateway uses a self-signed certificate. Accept the warning in your browser or use `curl -k`.

## References

- [IBKR Client Portal API Docs](https://interactivebrokers.github.io/cpwebapi/)
- [IBKR API Campus](https://www.interactivebrokers.com/campus/ibkr-api-page/cpapi-v1/)
