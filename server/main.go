package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
)


//check if removeStr or removeconnelement2 are in use before allowing others to use it

// Removes string element from string array
func removeStr(s []string, str string) []string {
	for i, v := range s {
		if v == str {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// Removes element from net.Conn slice
// func removeConnElement(c []net.Conn, i int) []net.Conn {
// 	return append(c[:i], c[i+1:]...)
// }

func removeConnElement2(c []net.Conn, el net.Conn) []net.Conn {
	for i, v := range c {
		if v == el {
			return append(c[:i], c[i+1:]...)
		}
	}
	return c
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

	// listener, err := net.Listen("tcp", ":8080")
	// listener, err := net.Listen("tcp", "192.168.0.3:8080")
	listener, err := net.Listen("tcp", "192.168.0.3:")
	if err != nil {
		fmt.Println("Error with listening: ", err.Error())
		return
		//os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Server Address: " + listener.Addr().String())

	// creating an array of existing usernames
	usernames := []string{}
	// usernames := make([]string, 0)
	clients := []net.Conn{}
	//everytime someone disconnects the username gets deleted
	//for loop to accept multiple incoming connections

	// chat messages
	chatMess := []string{}
	run := true

	// Updating chat in go routine
	go func() {
		for run {
			if len(chatMess) != 0 {
				message := chatMess[0]
				size := len(clients)
				for i := 0; i < size; i++ {
					//dont allow to change the array while doing this
					//and if the array is being changed then wait
					writeFromStr(clients[i], message)
				}
				// Remove front message from the "queue" of messages
				chatMess = chatMess[1:]
			}
		}
	}()

	go func() {
		for run {
			// fmt.Println("In for loop for incoming connections!")

			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("Error with listening: ", err.Error())
				//os.Exit(1)
			}
			// fmt.Println("Local addr: " + conn.LocalAddr().String())  // this is mine aka the server
			// fmt.Println("Remote addr: " + conn.RemoteAddr().String()) //this is the clients

			// fmt.Println("Passed .Accept point!")
			go func(c net.Conn) {
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

				again := contains(usernames, username) || !valid(username)
				//Loop to let the client create a username that is valid and not already in use
				for again {
					conn.Write([]byte("again"))
					// conn.Write([]byte("Uh oh! The username you chose is either taken or contains forbidden symbols.\nAllowed symbols are all letters and digits and '.', '-', '_', '~'.\nTry entering another username: "))
					if contains(usernames, username) {
						conn.Write([]byte("Uh oh! The username you chose is already taken! Enter another username: "))
					}
					if !valid(username) {
						conn.Write([]byte("Uh oh! The username you chose should be longer than 1 character and consist only of letters, numbers, '.', '-', '_' or '~'.\nEnter another username: "))
					}

					username = removeWhitespaces(readFromCon(conn))
					again = contains(usernames, username) || !valid(username)
				}
				conn.Write([]byte(username))

				//Adding the chosen username to the slice of usernames
				usernames = append(usernames, username)
				//Adding the succesfully connected to the chat clients
				clients = append(clients, conn)

				//TEMP remember to make client read this
				conn.Write([]byte("You have successfully connected to the server! To leave just type \"!exit\"\n"))

				// Announcing to everyone in the chat that username has joined
				for i := 0; i < len(clients); i++ {
					conn.Write([]byte(username + " joined the chat! Say hi!\n"))
				}
				
				// Receiving client messeges while the server is running
				for run {
					receive := readFromCon(conn)
					if receive == "!exit" {
						usernames = removeStr(usernames, username)
						clients = removeConnElement2(clients, conn)
						conn.Close()
						break
					} else {
						// Adding new message to the "queue" of all chat messages
						chatMess = append(chatMess, receive)
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
