package main

//importing necessary packages
import (
	"fmt"	// provides I/O functions
	"io/ioutil" // I/O utility functions
	"log" //provides logging functions
	"net/http" //http client-server implementations
	"strings" // string manipulation

	"github.com/gorilla/websocket" //for creating real-time communication
	gubrak "github.com/novalagung/gubrak/v2"
)

type M map[string]interface{}

const MESSAGE_NEW_USER = "New User"
const MESSAGE_CHAT = "Chat"
const MESSAGE_LEAVE = "Leave"

var connections = make([]*WebSocketConnection, 0)

// Declaring structs to support messages, its metadata and websocket connection

type SocketPayload struct {
	Message string //to hold the socket payload message
}

type SocketResponse struct {
	//to hold the response message and its metadata
	From    string
	Type    string
	Message string
}

type WebSocketConnection struct {
	//to manage new connections for every username
	*websocket.Conn
	Username string
}

func main() {

	//http handler to respond to entry point
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		//use the io util to readfile and store the html in the var
		content, err := ioutil.ReadFile("index.html")

		//error handling, cannot read html
		if err != nil {
			http.Error(w, "Could not open requested file", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "%s", content)
	})

	//http handler to establish a new connection for new users
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		currentGorillaConn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)

		// error handling for sign ups
		if err != nil {
			http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		}

		//setting new username, creating a new connection and adding it to the list of active connections
		username := r.URL.Query().Get("username")
		currentConn := WebSocketConnection{Conn: currentGorillaConn, Username: username}
		connections = append(connections, &currentConn)

		go handleIO(&currentConn, connections)
	})

	fmt.Println("Server starting at :8080")
	http.ListenAndServe(":8080", nil)
}

// handles input output of messages using the its own websocket and all others
func handleIO(currentConn *WebSocketConnection, connections []*WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("ERROR", fmt.Sprintf("%v", r))
		}
	}()

	broadcastMessage(currentConn, MESSAGE_NEW_USER, "")

	for {
		payload := SocketPayload{}
		err := currentConn.ReadJSON(&payload)

		//error handling
		if err != nil {
			if strings.Contains(err.Error(), "websocket: close") {
				broadcastMessage(currentConn, MESSAGE_LEAVE, "")
				ejectConnection(currentConn)
				return
			}

			log.Println("ERROR", err.Error())
			continue
		}

		//broadcast chat message
		broadcastMessage(currentConn, MESSAGE_CHAT, payload.Message)
	}
}

// handles sign outs
func ejectConnection(currentConn *WebSocketConnection) {
	filtered := gubrak.From(connections).Reject(func(each *WebSocketConnection) bool {
		return each == currentConn
	}).Result()
	connections = filtered.([]*WebSocketConnection)
}

// function that fires once a user hits send upon typing a msg
// takes in the connection obj, message type and the message
func broadcastMessage(currentConn *WebSocketConnection, kind, message string) {
	for _, eachConn := range connections {
		if eachConn == currentConn {
			continue
		}

		//for every connection other than itself, use the WriteJSON method provided by websocket to publish new messages
		eachConn.WriteJSON(SocketResponse{
			From:    currentConn.Username,
			Type:    kind,
			Message: message,
		})
	}
}
