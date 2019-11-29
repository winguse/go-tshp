# Tunnel of Secure HTTP Proxy

Usage as ssh proxy:

```
# .ssh/config

Host host-via-tshp
 HostName your-host-name
 User your-username
 ProxyCommand go-tshp --proxy-host=your-proxy-server:443 --target-host=%h:%p

```
