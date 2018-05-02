# A QUIC based Proxy！(Just a toy)

# Usage

## Install

```
# client peer
go get -u -v github.com/freedomio/fio-go/cmd/fioc 
# server peer
go get -u -v github.com/freedomio/fio-go/cmd/fiod
```

### Run

```
# On your server computer
fiod -l :9093
# On your pc
fioc -l 127.0.0.1:8987 -r your_server_ip:9093
```

Enjoy it!