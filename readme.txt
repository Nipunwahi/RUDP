Compile 
go mod init Share
go build 

Run
Server
./Share -PATH filepath -PORT ipaddress

Client
./Share -PORT ipaddress 
ipaddress is of the server

Example
Server
./Share -PATH file.txt -PORT 127.0.0.1:4444

Client
./Share -PORT 127.0.0.1:4444



CHANGES IN PROTOCOL:
We added window support and Selective repeat in our protocol (would still work with the first protocol but added as a feature). The updated protocol pdf is also given.


Members
Nipun Wahi 2018A7PS0966H
Hrithik Kulkarni 2018A7PS0278H
Ameetesh Sharma 2018A7PS0167H
Mir Ameen Mohideen 2018A7PS0487H
Nielless Acharya 2018A7PS0207H
