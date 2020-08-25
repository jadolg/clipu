# clipu

Yet another tool to share your clipboard amongst devices in your local network.

## Security Note

Clipu is absolutely NOT safe to use! 

Your data travels in plain text over the network!

Do not use it if you don't trust the network you are connected to!

## Configuration

Set `CLIPU_PASSWORD` to the same value on every device you want to use.
Set `CLIPU_PEER_LIMIT` if you want more than 1 other station (default: 1)
With CLIPU_PEER_LIMIT=1 you can connect 2 devices. 
The higher this number, the longer it takes to discover devices.

## Why?

I was sick of copying links from my work MacBook to my personal Linux station.
