# clipu

Yet another tool to share your clipboard amongst devices in your local network.

## Security

Clipboard content is encrypted with [age](https://github.com/FiloSottile/age) using scrypt-based symmetric encryption before being sent over the network. Only peers that share the same `CLIPU_PASSWORD` can decrypt the data.

## Configuration

| Variable | Required | Default | Description |
|---|---|---|---|
| `CLIPU_PASSWORD` | yes | — | Shared secret used for authentication and encryption. Must be set to the same value on every device. |
| `CLIPU_PEER_LIMIT` | no | `1` | Maximum number of peers to discover. With the default of `1` you can connect 2 devices. The higher this number, the longer discovery takes. |
| `CLIPU_LOG_LEVEL` | no | `info` | Log verbosity. Accepted values: `trace`, `debug`, `info`, `warn`, `error`, `fatal`, `panic`. |
| `CLIPU_ALLOW_SELF` | no | unset | When set (any value), allows discovering other instances running on the same machine. Useful for testing. |

## Why?

I was sick of copying links from my work MacBook to my personal Linux station.
