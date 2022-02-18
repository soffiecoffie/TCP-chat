package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

var (
	run             = true
	sendingMessages = true
	ok              = false
	cleaningHistory = false
	chat            ConcurrentSlice
	ind             ConcSliceIndices
)

type ConcurrentSlice struct {
	sync.Mutex
	messages []string
}

// Appends string to the messages of a ConcurrentSlice
func (slice *ConcurrentSlice) Append(str string) int {
	slice.Lock()
	defer slice.Unlock()

	slice.messages = append(slice.messages, str)
	return len(slice.messages) - 1
}

// Removes the last ind messages from the messages of a ConcurrentSlice
func (slice *ConcurrentSlice) CleanHistory(ind int) {
	slice.Lock()
	defer slice.Unlock()
	cleaningHistory = true
	slice.messages = slice.messages[ind:]
	cleaningHistory = false
}

type ConcSliceUsr struct {
	sync.Mutex
	usernames []string
}

// Append a username to the usernames slice of a ConcSliceUser
func (slice *ConcSliceUsr) Append(usr string) {
	slice.Lock()
	defer slice.Unlock()

	slice.usernames = append(slice.usernames, usr)
}

// Removes username from usernames slice of a ConcSliceUser
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

type ConcSliceIndices struct {
	sync.Mutex
	indices []*int
}

// Appends an index to the indices slice of a ConcSliceIndices
func (slice *ConcSliceIndices) Append(ind *int) {
	slice.Lock()
	defer slice.Unlock()

	slice.indices = append(slice.indices, ind)
}

// Removes index from the indices slice of a ConcSliceIndices
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
		if !run {
			os.Exit(0)
		}
		if err != io.EOF {
			fmt.Println("Error with reading: ", err.Error())
			return ""
		}
		fmt.Println("err = io.eof")
		return ""
	}

	// Removes null characters from buf and puts result in slice
	s := bytes.Trim(buf, "\x00")

	return string(s)
}

// Writes a given string to a given connection and returns the sent message
func writeFromStr(c net.Conn, message string) string {
	// Input the message in buf
	buf := []byte(message)
	// Write to server
	_, err := c.Write(buf)
	if err != nil {
		fmt.Println("Error with writing: ", err.Error())
		return ""
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
	for _, v := range str {
		if v == ' ' || v == '\t' || v == '\n' {
			return true
		}
	}
	return false
}

// Checks if there are "forbidden" symbols in a given string
func hasForbiddenSymbols(str string) bool {
	// allowed symbols: A-Z, a-z, ., -, _, ~
	for _, v := range str {
		if !((v >= 65 && v <= 90) || (v >= 97 && v <= 122) ||
			(v >= 48 && v <= 57) || v == '-' || v == '.' ||
			v == '_' || v == '~') {
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

// Handles signals
func handleSignals() {
	// Channel to read signals.
	sigs := make(chan os.Signal, 1)
	// Registers the given channel to receive notifications of the specified signals.
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	for {
		sig := <-sigs
		switch sig {
		case syscall.SIGHUP:
			fmt.Println("SIGHUP")
			closingServer()
			os.Exit(0)
		case syscall.SIGINT:
			fmt.Println("SIGINT")
			closingServer()
			os.Exit(0)
		case syscall.SIGTERM:
			fmt.Println("SIGTERM")
			closingServer()
			os.Exit(0)
		case syscall.SIGQUIT:
			fmt.Println("SIGQUIT")
			closingServer()
			os.Exit(0)
		default:
			fmt.Println("Unknown signal")
		}
	}
}

// Closes server
func closingServer() {
	fmt.Println("Stopping server...")
	run = false

	chat.Append("Chat is closing now! See you next time!")
	if !ok {
		checkOk()
	}
	sendingMessages = false
	fmt.Println("Server has closed!")
}

// Reads from standard input and returns the message
func readFromStdin() string {
	var reader = bufio.NewReader(os.Stdin)
	message, _ := reader.ReadString('\n')

	if len(message) > 2 {
		// Removes last 2 characters which are '\n' and ascii(13) ""
		message = message[:len(message)-2]
	} else if len(message) == 1 && message[0] == '\n' ||
		len(message) == 2 && message[0] == 13 && message[1] == '\n' {
		return ""
	}
	return message
}

// Checks if all messages have been read
func checkOk() {
	allread := false
	// Until every client reads all messages
	for !allread {
		allread = true
		for i := 0; i < len(ind.indices); i++ {
			if *ind.indices[i] < len(chat.messages) {
				allread = false
			}
		}
	}
	ok = true
}

func main() {
	fmt.Print("Input your local ip address (e.g. 192.168.0.3) to create a server: ")
	var host string
	fmt.Scanf("%s", &host)
	listener, err := net.Listen("tcp", host+":")

	if err != nil {
		fmt.Println("Error with listening: ", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	ok = true

	fmt.Println("Server Address: " + listener.Addr().String())
	fmt.Println("To stop server write \"stop\"")

	// Creating an array of existing usernames
	var users ConcSliceUsr

	go handleSignals()

	go func() {
		for run {
			if len(chat.messages) >= 256 {
				var allread bool
				for i := 0; i < len(ind.indices); i++ {
					if *(ind.indices)[i] < 255 {
						allread = false
						fmt.Println("INSIDE HISTORY CLEAN! ")
					}
				}
				// Until everyone reads the last 256 messages
				for allread {
					allread = true
					for i := 0; i < len(ind.indices); i++ {
						if *ind.indices[i] < 255 {
							allread = false
						}
					}
				}

				chat.CleanHistory(256)
				for i := 0; i < len(ind.indices); i++ {
					*ind.indices[i] -= 255
				}
			}
		}
	}()

	go func() {
		for run {
			conn, err := listener.Accept()
			if err != nil {
				if run {
					fmt.Println("Error with listening: ", err.Error())
				} else {
					return
				}
			}
			ok = false

			go func(conn net.Conn) {
				fmt.Println("Remote address " + conn.RemoteAddr().String() + " connected!")

				// Asking the client to create a username
				conn.Write([]byte("Welcome to the server! Create a username: "))

				defer conn.Close()
				username := readFromCon(conn)

				// Removes the whitespaces from the username if there are any
				username = removeWhitespaces(username)
				if username == "!exit" {
					fmt.Println(conn.RemoteAddr().String() + " exited.")
					// Telling client it's ok to exit
					writeFromStr(conn, "!exit")
					return
				}

				again := contains(users.usernames, username) || !valid(username)

				//Loop to let the client create a username that is valid and not already in use
				for again {
					if username == "!exit" {
						fmt.Println(conn.RemoteAddr().String() + " exited.")
						conn.Close()
						return
					}

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

				conn.Write([]byte("You have successfully connected to the server! To leave just type \"!exit\"\n"))

				// Index to start reading messages from
				readInd := chat.Append(username + " joined the chat! Say hi!\n")
				ind.Append(&readInd)

				sendMsgToCon := true
				// The index to the last message that needs to be read
				var lastInd int
				// Updating chat in go routine
				go func() {
					for sendingMessages && sendMsgToCon {
						for readInd < len(chat.messages) {
							writeFromStr(conn, chat.messages[readInd])
							readInd++
						}
					}
				}()

				// Receiving client messages while the server is running
				for run {
					receive := readFromCon(conn)
					if receive == "!exit" {
						sendMsgToCon = false
						lastInd = chat.Append(username + " disconnected.")
						if cleaningHistory {
							lastInd -= 255
						}

						// Assuring all messages up to the one above will be read
						for readInd <= lastInd {
							writeFromStr(conn, chat.messages[readInd])
							readInd++
						}

						// Telling client it's ok to exit
						writeFromStr(conn, "!exit")
						users.RemoveUsername(username)
						ind.RemoveInd(&readInd)
						return
					} else {
						// Adding new message to the chat messages slice
						chat.Append(receive)
					}
				}
			}(conn)
		}
	}()

	var cmd string
	for run {
		cmd = readFromStdin()
		if cmd == "stop" {
			if len(users.usernames) == 0 {
				ok = true
			}
			closingServer()
			return
		}
	}
}
