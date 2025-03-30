## Containers

To enable easier local testing, multiple rVASP images can be run concurrently using docker compose. This requires a setup step to generate some local fixtures and a build step to build the images.

### Generating the fixtures

```
$ ./scripts/generate-fixtures.sh
```

This creates a `fixtures` directory at the root of the repository which contains three subdirectories:

1. A `certs` directory that contains certificates for the `alice`, `bob`, and `evil` rVASPs. The `evil` rVASP is derived from a different certificate authority than `alice` and `bob` and therefore won't be able to authenticate with them.
2. A `vasps` directory that contains JSON files representing rVASP records in a directory service.
3. A `db` directory that contains an embedded directory service database.

These fixtures enable the rVASPs to look each other up by name using a local directory service.

### Building the images

```
$ ./containers/build.sh
```

This is a script that automates the build process by first building the rVASP base image and then the other images defined in `containers/docker-compose.yml`. Alternatively, this can be done as two separate steps.

```
$ docker build -t trisa/rvasp:latest -f ./containers/rvasp/Dockerfile .
$ docker compose -f ./containers/docker-compose.yml build
```

### Starting the containers

```
$ docker compose -f ./containters/docker-compose.yml up
```

This starts the local directory service, initializes a shared postgres database for the rVASPs, and starts the `alice`, `bob`, and `evil` rVASPs.

Additionally you can use the `skaffold build` command to build and push the images to Dockerhub.

## Interacting with the rVASPs

The rVASP CLI can be used to interact with the rVASPs running in the containers. By default, `alice` listens on `localhost:5434`, `bob` listens on `localhost:6434`, and `evil` listens on `localhost:7434`. For example, one might initiate a transfer between two accounts using the `transfer` command:

```
$ go run ./cmd/rvasp transfer -e localhost:6434 -a robert@bobvasp.co.uk -b mary@alicevasp.us -B api.alice.vaspbot.com -d 42.99 -E
```

The `resetdb` command can be used to reset the rVASP database without restarting the containers, using a fixtures directory containing a `vasps.json` and `wallets.json` (defaults to `pkg/rvasp/fixtures`).

```
$ go run ./cmd/rvasp resetdb -d postgres://postgres:postgres@localhost:5433/rvasp?sslmode=disable -f pkg/rvasp/fixtures
```