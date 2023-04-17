# LightSocks
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