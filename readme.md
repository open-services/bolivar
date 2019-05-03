# Bolivar
> Federated Open-Registry proxy for p2p sharing of packages


[![Imagelink](https://i.imgur.com/g9jEN6T.png)](https://www.youtube.com/watch?v=VA1HibxXOx0)

## Introduction

Bolivar is a prototype to move the registry closer to wherever you are, and
enable local network sharing of already downloaded data.

This means:

- Faster performance as mirrors are closer to you
- You can easily install a registry proxy in your local network
- You'll be able to install modules from your friends next to you, even
  without a upstream internet connection

## Usage

Install bolivar.

Then you just run `bolivar` and it's up and running!

Next step is pointing your package manager to it:

```bash
# replace npm with yarn or pnpm if you use those
npm config set registry http://localhost:8080
```

You can see all the different options by doing `bolivar --help`

```
$ bolivar --help
usage: bolivar-vlatest-linux-amd64 [<flags>]

A federated version of Open-Registry (https://open-registry.dev)

Flags:
      --help                Show context-sensitive help (also try --help-long and --help-man).
      --federate="/dns4/npm.open-registry.dev/tcp/4001/ipfs/QmbNjMCXkwt7fyBBWc3R8mZzrX8KebkCo4Qv67wVmCH5Aa"
                            Multiaddr of primary federation server to read from
  -s, --share               Whetever to share downloaded packages to others
      --update-type="http"  How to receive updates to root hash.
  -o, --offline             Disable any external connections
  -v, --verbose             Verbose mode.
  -l, --listen-addr="/ip4/0.0.0.0/tcp/4005"
                            What address to set the libp2p node to listen at
  -r, --repo-path="ipfs-test-repo"
                            What path to use for the repository
  -a, --http-address="localhost"
                            What address to use for HTTP server
  -p, --http-port="8080"    What port to use for HTTP server
      --http-endpoint="https://npm.open-registry.dev"
                            What the endpoint for registry index is
      --version             Show application version.
```

Example output when running:

```
$ bolivar
Starting libp2p host
libp2p listening on the following addresses:
/ip4/127.0.0.1/tcp/4005
/ip4/192.168.2.238/tcp/4005
/ip4/172.17.0.1/tcp/4005
Connecting to open-registry.dev libp2p node...
Connected
Think we protected a bootstrap peer from being killed too
Starting server localhost:8080...
0 B|0 B [0 B/s|0 B/s] (Peers: 1)
19 kB|832 B [19 kB/s|831 B/s] (Peers: 1)
19 kB|832 B [7.0 kB/s|306 B/s] (Peers: 1)
19 kB|832 B [2.6 kB/s|112 B/s] (Peers: 1)
19 kB|832 B [942 B/s|41 B/s] (Peers: 1)
19 kB|832 B [346 B/s|15 B/s] (Peers: 3)
```

## Contributing

### Testing

- `make test` - runs bolivar and tries to download dependencies for webpack/webpack via bolivar

### Building

- `make linux` / `make windows` / `make darwin` builds binary for linux/windows/darwin

### Linting

- `make lint` runs gometalinter over all the .go source files in the repo

## License

MIT 2019 - Open-Registry
