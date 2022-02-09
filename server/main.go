package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
)

// Removes string element from string array
// func removeElement(s []string, str string) []string {
// 	for i, v := range s {
// 		if v == str {
// 			s = append(s[:i], s[i+1:]...)
// 			break
// 		}
// 	}
// 	return s
// }

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
func main() {
	// Create listener
	listener, err := net.Listen("tcp", "192.168.0.3:")
	if err != nil {
		fmt.Println("Error with listening: ", err.Error())
		return
		//os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Server Address: " + listener.Addr().String())

	//creating an array of existing usernames
	usernames := []string{}

	//for loop to accept multiple incoming connections
	for {
		// fmt.Println("In for loop for incoming connections!")

		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error with listening: ", err.Error())
			//os.Exit(1)
		}
		// fmt.Println("Local addr: " + conn.LocalAddr().String())  // this is mine aka the server
		// fmt.Println("Remote addr: " + conn.RemoteAddr().String()) //this is the clients

		go func(c net.Conn) {
			// conn.Write([]byte("Welcome to the server! Pick a nickname: "))
			// conn.Write([]byte("Welcome to the server! What should we call you: "))

			//Asking the client to create a username
			conn.Write([]byte("Welcome to the server! Create a username: "))

			// username = string(buf[:])
			username := readFromCon(conn)
			//Removes the whitespaces from the username if there are any
			username = removeWhitespaces(username)
			fmt.Println("TEMP picked username: -" + username + "-")

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

			// Adding the chosen username to the slice of usernames
			usernames = append(usernames, username)

			conn.Write([]byte("Welcome to the server " + username + "! Say hi to everyone!"))
			conn.Write([]byte(username + " joined the server! Say hi!"))
			defer c.Close()
		}(conn)
	}
}
