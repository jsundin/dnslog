# dnslog

Will refuse queries to any sub-/domain not matching the configured domain. Will respond with NXDOMAIN to all legitimate requests.

## Usage
The server needs to be registered as the dns server for a zone, such as `mydomain.com` used in the examples below.

Start:
```
dnslog                      \
    -dns 53                 \
    -domain mydomain.com.   \
    -http 8080              \
    -level info
```

Make a request that ends up in the log:
```
nslookup server-a.mytoken.mydomain.com
```

Check the log:
```
curl http://server:8080/mytoken
```

## Local test example
```
./dnslog -dns 1053 -domain dnslog.lab. -http 8080 -level info
```

```
nslookup -port=1053 server-a.mytoken.dnslog.lab
```

```
curl http://localhost:8080/mytoken
```
