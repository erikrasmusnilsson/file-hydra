# File Hydra
Originally written in Java utilising raw TCP sockets, it is now implemented with Golang as a RESTful API. File-hydra splits a file into several parts and sends each part to one of the cooperating clients. It is up to the clients to then patch the file together over a LAN connection. 

For best effect, use a strong cellular connection to connect to the server and a wired LAN connection to spread the file parts between clients. 

## Interaction
The interaction starts with a master client. The client sends a POST request to /sessions with the following body:
```
{
    "filename": "path/to/file.txt",
    "expectedClients": 3
}
```
The server will then check if the file exists and respond with an UUID corresponding to the created download session. The created session is valid for 5 minutes.
```
{
    "id": "some-uuid-as-a-string",
    "filename": "path/to/file.txt",
    "expectedClients": 3,
    "connectedClients": 0
}
```
Each client then sends a GET request to /sessions/:id where it will wait until all clients are connected and finally respond with the clients corresponding part of the file. The bytes of the file will be encoded as a base64 string as it travels through the wire. An additional HTTP header "X-Partition-Number" is added to the response. The value of this header field represents the partition that any given client has received. 
![Interaction diagram](/doc/out/concept.png)