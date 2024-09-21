# simple-go-chat-api-w-mongo-webscoket
Building a Chat App  Api with GO, Mongodb, Websockets

## 1. Install Go
- https://golang.org/doc/install
- https://golang.org/doc/code.html
- https://golang.org/doc/effective_go.html

## 2. Install Gorilla WebSocket
-   `go get -u github.com/gorilla/websocket`
-  `go get -u github.com/gorilla/mux`
- `go get -u github.com/gorilla/handlers`

## 3. Install Mongodb
- https://docs.mongodb.com/manual/installation/
- https://docs.mongodb.com/manual/tutorial/install-mongodb-on-ubuntu/

## 4. Setup env
- Create a .env file in the root directory
- Add the following variables
```
MONGO_URI=mongodb://localhost:27017
DB_NAME=chat
PORT=8080
etc
```

## 5. Run the server
-   `go run main.go`


References:
- https://dev.to/gbubemi22/building-a-simple-chat-application-with-go-gin-mongodb-and-websocket-2joo
- https://pkg.go.dev/github.com/golang-jwt/jwt/v5#example-Parse-Hmac



