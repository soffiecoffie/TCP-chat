package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
)

//Add cmd to say number of users and all usernames and private message to user by saying "!private /username/ /message/"
//check if server terminates badly and kicks you out

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
	// return strings.Join(s, " ")
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

func main() {
	fmt.Print("Input the server you want to connect to: ")
	var addr string
	fmt.Scanf("%s", &addr)
	client, err := net.Dial("tcp", addr)

	// client, err := net.Dial("tcp", "192.168.0.3:8080")
	// client, err := net.Dial("tcp", "192.168.0.3:0")
	// client, err := net.Dial("tcp", "192.168.0.3:64619")
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

	//TURNS OUT SCANLN doesnt scan the fucking line and ignores ALL whitespaces
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
	//use this
	// scanner := bufio.NewScanner(os.Stdin)
	// if scanner.Scan() {
	//     line := scanner.Text()
	//     fmt.Printf("Input was: %q\n", line)
	// }

	// Channel to read signals.
	sigs := make(chan os.Signal, 1)

	// Registers the given channel to receive notifications of the specified signals.
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		for {
			sig := <-sigs
			switch sig {
			case syscall.SIGHUP:
				fmt.Println("SIGHUP")
			case syscall.SIGINT:
				fmt.Println("SIGINT")
			case syscall.SIGTERM:
				fmt.Println("SIGTERM")
			case syscall.SIGQUIT:
				fmt.Println("SIGQUIT")
			default:
				fmt.Println("Unknown signal")
			}
		}
	}()

}
