# LightSocks
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fxmapst%2Flightsocks.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fxmapst%2Flightsocks?ref=badge_shield)

support socks4, socks4a, socks5, socks5h, http proxy all in one

## Help

```bash
Support socks4, socks4a, socks5, socks5h, http proxy all in one,

Usage:
  ./lightsocks [flags]
  ./lightsocks [command]

Available Commands:
  client      Support socks4, socks4a, socks5, socks5h, http proxy all in one.
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  server      

Flags:
  -c, --config string   config file path (default "config.yaml")
  -h, --help            help for ./lightsocks

```

## Build

```bash
git clone https://github.com/xmapst/lightsocks.git
cd lightsocks
make
```

## Direct Mode

see `config.yaml` configure

```bash
./lightsocks -c example/config.yaml
```

## Proxy Mode

see `server.yaml` and `client.yaml` configure

```bash
sitsed ./lightsocks server -c example/server.yaml 
sitsed ./lightsocks client -c example/client.yaml
```

## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fxmapst%2Flightsocks.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fxmapst%2Flightsocks?ref=badge_large)