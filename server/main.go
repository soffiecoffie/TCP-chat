package main

import (
	"fmt"
	"io"
	"net"
	"strings"
)

//Removes string element from string array
// func removeElement(s []string, str string) []string {
// 	for i, v := range s {
// 		if v == str {
// 			s = append(s[:i], s[i+1:]...)
// 			break
// 		}
// 	}
// 	return s
// }

// Checks if a slice contains a given string
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
func hasWhitespaces(str string) bool {
	for i := 0; i < len(str); i++ {
		if str[i] == ' ' || str[i] == '\t' || str[i] == '\n' {
			return true
		}
	}
	return false
}

// should i also forbid symbols that aren't letters or numbers? i dont like usernames like ____-
// Checks if there are "forbidden" symbols in a given string
func hasForbiddenSymbols(str string) bool {
	//allowed symbols: A-Z, a-z, ., -, _, ~
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
	//The server does not accept names that:
	// - are too short
	// - have whitespaces
	// - have forbidden symbols
	return len(str) > 1 && !hasWhitespaces(str) && hasForbiddenSymbols(str)
}
func main() {
	//Creating listener
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

	for {
		fmt.Println("In for loop for incoming connections!")

		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error with listening: ", err.Error())
			//os.Exit(1)
		}

		fmt.Println("Passed .Accept point!")
		go func(c net.Conn) {
			fmt.Println("In go routine!")

			// conn.Write([]byte("Welcome to the server! Pick a nickname: "))
			// conn.Write([]byte("Welcome to the server! What should we call you: "))

			//Asking the client to create a username
			conn.Write([]byte("Welcome to the server! Create a username: "))
			var username string
			buf := make([]byte, 1024)
			_, err = conn.Read(buf)
			if err != nil {
				if err != io.EOF {
					fmt.Println("Error with reading: ", err.Error())
				}
				// break
				return
			}

			//temp check ins
			fmt.Println("first read: ", buf)

			//Loop to let the client create a username that is not already in use
			username = string(buf)
      
			//Removes the whitespaces from the username if there are any
			username = removeWhitespaces(username)
			again := contains(usernames, username) && valid(username)
      
			//Loop to let the client create a username that is not already in use
      for again {
				_, err = conn.Read(buf)
				if err != nil {
					if err != io.EOF {
						fmt.Println("Error with reading: ", err.Error())
					}
					return
				}
				username = string(buf[:])

				//temp check ins
				fmt.Println("next username read: ", buf)

				fmt.Println("next username read from buf to string: ", username)

				again = contains(usernames, username)
			}

			//Adding the chosen username to the slice of usernames
			usernames = append(usernames, username)

			//Telling the client their username
			conn.Write([]byte("Welcome to the server " + username + "! Say hi to everyone!"))
			conn.Write([]byte(username + " joined the server! Say hi!"))
		
			defer c.Close()
		}(conn)
	}
}
