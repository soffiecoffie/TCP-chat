package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
)

//You can add cmd to say number of users and all usernames and private message to user by saying "!private /username/ /message/"
//check if server terminates(ctrl+c) badly and kicks you out

// Reads from Server and returns the message as a string instead of buffer
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
}

// Writes to server from a given string and returns the sent message
func writeToServerStr(c net.Conn, message string) string {
	// Input the message in buf
	buf := []byte(message)
	// Write to server
	_, err := c.Write(buf)
	if err != nil {
		// if err != io.EOF {
		// 	fmt.Println("Error with reading: ", err.Error())
		// }
		// break
		// return //yes or no
		os.Exit(1)
	}
	return message
}

// Writes to server from standard input and returns the sent message
func writeToServer(c net.Conn) string {
	var message string
	fmt.Scanln(&message)
	return writeToServerStr(c, message)
}

// // Writes to the chat from standard input and returns the sent message
// func talkToChat(c net.Conn, usr string) string {
// 	var message string
// 	fmt.Scanf("%s", &message)
// 	return writeToServerStr(c, usr+": "+message)
// }

func main() {
	// fmt.Print("Input the server you want to connect to: ")
	// var addr string
	// fmt.Scanf("%s", &addr)
	// client, err := net.Dial("tcp", addr)

	// client, err := net.Dial("tcp", "192.168.0.3:8080")
	// client, err := net.Dial("tcp", "192.168.0.3:0")
	client, err := net.Dial("tcp", "192.168.0.3:64619")
	if err != nil {
		fmt.Println("Error with connecting: ", err.Error())
		return
	}
	defer client.Close()
	// fmt.Println("Local addr: " + client.LocalAddr().String()) //clients addr
	// fmt.Println("Remote addr: " + client.RemoteAddr().String()) //server addr

	// Writing to the server
	received := readFromCon(client)
	fmt.Print(received)
	writeToServer(client)
	received = readFromCon(client)
	again := received == "again"
	for again {
		received = readFromCon(client)
		fmt.Print(received)
		writeToServer(client)
		received = readFromCon(client)
		again = received == "again"
	}
	username := received
	// fmt.Println("Username is: " + username)

	received = readFromCon(client)
	fmt.Print(received)

	cmd := "!exit"
	run := true

	// Receiving chat messages
	go func() {
		for run {
			received = readFromCon(client)
			fmt.Println(received)
		}
	}()

	//TURNS OUT SCANLN doesnt scan the whole line and ignores ALL whitespaces so use bufio to replace it
	for run {
		// Talking in the chat
		var message string
		fmt.Scanln(&message)
		fmt.Println("YOUR MESSAGE IS:", message)
		if message == cmd {
			run = false
			writeToServerStr(client, username+" exited.")
			writeToServerStr(client, "!exit")
			fmt.Println("Sad to see you go!")
		} else {
			writeToServerStr(client, username+": "+message)
		}
	}
}
