# A QUIC based Proxy！(Just a toy)

# Usage

## Install

```
# client peer
go get -u -v github.com/freedomio/fio-go/cmd/fioc 
# server peer
wget -O fiod https://github.com/freedomio/fio-go/releases/download/0.0.3/fiod
chmod +x fiod
# or with golang envrionment on you server computer
go get -u -v github.com/freedomio/fio-go/cmd/fiod
```

### Run

```
# On your server computer
fiod -l :9093
# On your pc
fioc -l 127.0.0.1:8987(socks5 proxy listen) -r your_server_ip:9093 -lh 127.0.0.1:8988(http proxy listen)
```

Enjoy it!