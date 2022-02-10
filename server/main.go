package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

type ConcurrentSlice struct {
	sync.Mutex
	messages []string
}

func (slice *ConcurrentSlice) Append(str string) {
	slice.messages = append(slice.messages, str)
}

func (slice *ConcurrentSlice) CleanHistory(ind int) {
	slice.Lock()
	defer slice.Unlock()

	slice.messages = slice.messages[ind:]
}

///////////////////////////////
type ConcSliceUsr struct {
	sync.Mutex
	usernames []string
}

func (slice *ConcSliceUsr) Append(usr string) {
	slice.usernames = append(slice.usernames, usr)
}

func (slice *ConcSliceUsr) RemoveUsername(str string) {
	slice.Lock()
	defer slice.Unlock()

	for i, v := range slice.usernames {
		if v == str {
			slice.usernames = append(slice.usernames[:i], slice.usernames[i+1:]...)
			break
		}
	}
}

///////////////////////////////
type ConcSliceIndices struct {
	sync.Mutex
	indices []*int
}

func (slice *ConcSliceIndices) Append(ind *int) {
	slice.indices = append(slice.indices, ind)
}

func (slice *ConcSliceIndices) GetInd(ind int) *int {
	return slice.indices[ind]
}
func (slice *ConcSliceIndices) RemoveInd(ind *int) {
	slice.Lock()
	defer slice.Unlock()

	for i, v := range slice.indices {
		if v == ind {
			slice.indices = append(slice.indices[:i], slice.indices[i+1:]...)
			break
		}
	}
}

// Reads from Client and returns the message as a string instead of buffer
func readFromCon(c net.Conn) string {
	buf := make([]byte, 1024)
	_, err := c.Read(buf)
	if err != nil {
		if err != io.EOF {
			fmt.Println("Error with reading: ", err.Error())
		}
		// break
		// return
		//os.Exit(1)
		fmt.Println("err = io.eof")
	}

	// Removes null characters from buf and puts result in slice
	s := bytes.Trim(buf, "\x00")

	return string(s)
	// return strings.Join(s, " ")
}

// Writes a given string to a given conn and returns the sent message
func writeFromStr(c net.Conn, message string) string {
	// Input the message in buf
	buf := []byte(message)
	// Write to server
	_, err := c.Write(buf)
	if err != nil {
		// if err != io.EOF {
		fmt.Println("Error with writing: ", err.Error())
		// }
		// break
		// return //yes or no
		// os.Exit(1)
	}
	return message
}

// Checks if a slice contains a given string
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// Checks if str has whitespaces
func hasWhitespaces(str string) bool {
	for i := 0; i < len(str); i++ {
		if str[i] == ' ' || str[i] == '\t' || str[i] == '\n' {
			return true
		}
	}
	return false
}

// should i forbid symbols that aren't letters or numbers? i dont like usernames like ____-
// Checks if there are "forbidden" symbols in a given string
func hasForbiddenSymbols(str string) bool {
	// allowed symbols: A-Z, a-z, ., -, _, ~
	for i := 0; i < len(str); i++ {
		if !((str[i] >= 65 && str[i] <= 90) || (str[i] >= 97 && str[i] <= 122) ||
			(str[i] >= 48 && str[i] <= 57) || str[i] == '-' || str[i] == '.' ||
			str[i] == '_' || str[i] == '~') {
			return true
		}
	}
	return false
}

// Returns a string made of the given string without the whitespaces
func removeWhitespaces(str string) string {
	if hasWhitespaces(str) {
		// Makes a slice of the substrings in str between whitespaces
		res := strings.Split(str, " ")
		str = strings.Join(res, "")
		res = strings.Split(str, "\t")
		str = strings.Join(res, "")
		res = strings.Split(str, "\n")
		// Concatenates all the elements present in the slice of string into a single string.
		str = strings.Join(res, "")
	}
	return str
}

// Checks if a username is valid for the server
func valid(str string) bool {
	// The server does not accept names that:
	// - are too short
	// - have whitespaces
	// - have forbidden symbols
	return len(str) > 1 && !hasWhitespaces(str) && !hasForbiddenSymbols(str)
}

// find a way to check if user leaves with ctrl+C so it doesn't break my server

func main() {
	// fmt.Print("Input your local ip address (e.g. 192.168.0.3) to create a server: ")
	// var host string
	// fmt.Scanf("%s", &host)
	// listener, err := net.Listen("tcp", host+":")

	//For TCP networks, if the host in the address parameter is empty Listen listens on all available unicast and anycast IP addresses of the local system.  The address can use a host name, but this is not recommended, because it will create a listener for at most one of the host's IP addresses.
	// listener, err := net.Listen("tcp", ":8080")
	// listener, err := net.Listen("tcp", "192.168.0.3:8080")
	// you need to write ur own address
	listener, err := net.Listen("tcp", "192.168.0.3:")
	if err != nil {
		fmt.Println("Error with listening: ", err.Error())
		return
		//os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Server Address: " + listener.Addr().String())

	// creating an array of existing usernames
	var users ConcSliceUsr

	// chat messages
	var chat ConcurrentSlice
	var indices ConcSliceIndices
	run := true

	go func() {
		for run {
			if len(chat.messages) >= 256 {
				var allread bool
				//lock w
				for i := 0; i < len(indices.indices); i++ {
					if *(indices.indices)[i] < 256 { //whats tha val of *indices.indices[i]
						allread = false
					}
				}
				// Until everyone reads the last 256 messages
				for allread {
					allread = true
					for i := 0; i < len(indices.indices); i++ {
						if *indices.indices[i] < 256 {
							allread = false
						}
					}
				}

				chat.CleanHistory(256)
				for i := 0; i < len(indices.indices); i++ {
					*indices.indices[i] -= 256
					//unlock w
				}
			}
		}
	}()

	go func() {
		for run {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("Error with listening: ", err.Error())
				//os.Exit(1)
			}
			// fmt.Println("Local addr: " + conn.LocalAddr().String())  // this is mine aka the server
			// fmt.Println("Remote addr: " + conn.RemoteAddr().String()) //this is the clients

			// Handling the connection
			go func(conn net.Conn) {
				fmt.Println("Remote address " + conn.RemoteAddr().String() + " connected!")

				// conn.Write([]byte("Welcome to the server! Pick a nickname: "))
				// conn.Write([]byte("Welcome to the server! What should we call you: "))

				//Asking the client to create a username
				conn.Write([]byte("Welcome to the server! Create a username: "))

				// username = string(buf[:])
				username := readFromCon(conn)
				//Removes the whitespaces from the username if there are any
				username = removeWhitespaces(username)
				// fmt.Println("TEMP picked username: -" + username + "-")

				again := contains(users.usernames, username) || !valid(username)
				//Loop to let the client create a username that is valid and not already in use
				for again {
					conn.Write([]byte("again"))
					if contains(users.usernames, username) {
						conn.Write([]byte("Uh oh! The username you chose is already taken! Enter another username: "))
					}
					if !valid(username) {
						conn.Write([]byte("Uh oh! The username you chose should be longer than 1 character and consist only of letters, numbers, '.', '-', '_' or '~'.\nEnter another username: "))
					}

					username = removeWhitespaces(readFromCon(conn))
					again = contains(users.usernames, username) || !valid(username)
				}
				conn.Write([]byte(username))

				//Adding the chosen username to the slice of usernames
				users.Append(username)

				//TEMP remember to make client read this
				conn.Write([]byte("You have successfully connected to the server! To leave just type \"!exit\"\n"))

				// Adding announcement to the chat messages
				chat.Append(username + " joined the chat! Say hi!\n")

				// Index to start reading messages from
				ind := len(chat.messages) - 1

				// Updating chat in go routine
				go func() {
					for run {
						if ind < len(chat.messages) {
							writeFromStr(conn, chat.messages[ind])
							ind++
						}
					}
				}()

				// Receiving client messages while the server is running
				for run {
					receive := readFromCon(conn)
					if receive == "!exit" {
						chat.Append(username + " disconnected.")
						users.RemoveUsername(username)
						indices.RemoveInd(&ind)
						conn.Close()
						break
					} else {
						// Adding new message to the chat messages slice
						chat.Append(receive)
					}
				}

			}(conn)
		}
	}()

	var cmd string
	for {
		fmt.Scanf("%s", cmd)
		if cmd == "stop" {
			run = false
		}
	}
}
