# RUDP
A reliable UDP protocol implemented in Go following selective repeat paradigm


## Compile 
```
go mod init Share
go build
```
## Run
### Server
```
./Share -PATH filepath -PORT ipaddress
```

### Client
```
./Share -PORT ipaddress 
```
ipaddress is of the server

### Example
### Server
```
./Share -PATH file.txt -PORT 127.0.0.1:4444
```

### Client
```
./Share -PORT 127.0.0.1:4444
```

