# dnslog

Small DNS tokenized logger capable of forwarding. Designed for on-prem DNS token-based testing.
Will, by default, refuse queries to any sub-/domain not matching the configured domain. Will respond with NXDOMAIN to all legitimate requests.

Obviously if this is used non-locally this must be registered as a nameserver for a zone.

HTTP server supports `application/json` or `application/html` (default), depending on `Accept` header.

## Usage
The server needs to be registered as the dns server for a zone, such as `mydomain.com` used in the examples below.

### Arguments:
| Parameter | Description | Default
|-|-|-|
| `-dns <port>` | Listening port for DNS queries | `1053`
| `-domain <domain>` | Token domain, must end with a dot (will never be forwarded) | `dnslog.lab.`
| `-forward` | Enable forwarding for non-token domains | `false`
| `-upstream <server>` | DNS server to use for forwarding, with or without port | Whatever is in `/etc/resolv.conf`
| `-http <port>` | Listening port for HTTP queries | `8080`
| `-level <loglevel>` | Logging level (`panic`, `fatal`, `error`, `warn`, `info`, `debug`, `trace`) | `info`
||||

### Example:
```
# Start server
dnslog -dns 53 -domain dnslog.mydomain.com. -http 8080 -forward

# Make a request
nslookup server-a.mytoken.dnslog.mydomain.com

# Check the log
curl -H "Accept: application/json" http://server:8080/mytoken
```

## Local test example
```
# Start server
./dnslog -dns 1053 -domain dnslog.lab. -http 8080 -level info

# Make a request
nslookup -port=1053 server-a.mytoken.dnslog.lab

# Check token
curl -H "Accept: application/json" http://localhost:8080/mytoken
```
