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

### Members
- Nipun Wahi 2018A7PS0966H
- Hrithik Kulkarni 2018A7PS0278H
- Ameetesh Sharma 2018A7PS0167H
- Mir Ameen Mohideen 2018A7PS0487H
- Nielless Acharya 2018A7PS0207H
