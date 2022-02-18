package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var receivingMsgs = true
var inChat = false
var run = true

// Reads from Server and returns the message as a string instead of buffer
func readFromCon(c net.Conn) string {
	buf := make([]byte, 1024)
	_, err := c.Read(buf)
	if err != nil {
		if err != io.EOF {
			fmt.Println("Error with reading: ", err.Error())
			os.Exit(1)
			return ""
		}
		fmt.Println("err = io.eof") 
		os.Exit(1)                  
		return ""
	}

	// Removes null characters from buf and puts result in slice
	s := bytes.Trim(buf, "\x00")
	ifServerClosed(string(s))

	return string(s)
}

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

// Writes to server from a given string and returns the sent message
func writeToServerStr(c net.Conn, message string) string {
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

// Handles Signals
func handleSignals(client net.Conn, usr string, chatting bool) {
	// Making a channel to read signals.
	sigs := make(chan os.Signal, 1)
	fmt.Printf("IN HANDLE SIG WITH chatting= %t\n", chatting)
	// Registers the given channel to receive notifications of the specified signals.
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	stop := false
	for !stop {
		fmt.Printf("IN HANDLE SIG WITH chatting= %t\n", chatting)

		sig := <-sigs
		if chatting != inChat {
			return
		}
		switch sig {
		case syscall.SIGHUP:
			fmt.Println("SIGHUP")
			if chatting {
				writeToServerStr(client, usr+" exited.")
				run = false
				writeToServerStr(client, "!exit")
			} else {
				writeToServerStr(client, "!exit")
				os.Exit(0)
			}
			stop = true
			for receivingMsgs {
				time.Sleep(0 * time.Second)
			}
		case syscall.SIGINT:
			fmt.Println("SIGINT")
			if chatting {
				writeToServerStr(client, usr+" exited.")
				run = false
				writeToServerStr(client, "!exit")

			} else {
				writeToServerStr(client, "!exit")
			}
			stop = true
		case syscall.SIGTERM:
			fmt.Println("SIGTERM")
			if chatting {
				writeToServerStr(client, usr+" exited.")
				run = false
				writeToServerStr(client, "!exit")

			} else {
				writeToServerStr(client, "!exit")
			}
			stop = true
		case syscall.SIGQUIT:
			fmt.Println("SIGQUIT")
			if chatting {
				writeToServerStr(client, usr+" exited.")
				run = false
				writeToServerStr(client, "!exit")

			} else {
				writeToServerStr(client, "!exit")
			}
			stop = true
		default:
			fmt.Println("Unknown signal")
		}
	}
}

func ifServerClosed(msg string) {
	if msg == "Chat is closing now! See you next time!" {
		fmt.Println(msg)
		os.Exit(0)
	}
}

func main() {
	go handleSignals(nil, "", false)
	fmt.Print("Input the server you want to connect to: ")
	addr := readFromStdin()

	client, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Error with connecting: ", err.Error())
		os.Exit(1)
	}
	defer client.Close()

	inChat = true
	go handleSignals(client, client.RemoteAddr().String(), true)

	received := readFromCon(client)
	fmt.Print(received)

	username := readFromStdin()

	writeToServerStr(client, username)
	received = readFromCon(client)

	// Check if server has closed
	ifServerClosed(received)
	again := received == "again" || received == "!exit"
	for again {
		if received == "!exit" {
			fmt.Println("Sad to see you go!")
			os.Exit(0)
		}
		received = readFromCon(client)
		fmt.Print(received)
		username = readFromStdin()
		writeToServerStr(client, username)
		received = readFromCon(client)
		again = received == "again" || received == "!exit"

		// Check if server has closed
		ifServerClosed(received)
	}
	username = received

	// Receive the "You have successfully connected to the server! To leave just type \"!exit\"\n" message
	received = readFromCon(client)
	fmt.Print(received)

	receivingMsgs := true

	ch := make(chan int)
	// Receiving chat messages
	go func() {
		for receivingMsgs {
			received = readFromCon(client)

			// Client recieves "!exit" from the server when it's ok to exit
			if received == "!exit" {
				fmt.Println("Sad to see you go!")

				receivingMsgs = false
				ch <- 1
				return
			} else {
				fmt.Println(received)
			}
		}
	}()

	var message string
	for run || receivingMsgs {
		// Talking in the chat
		if run {
			message = readFromStdin() //blocks and waits for input
			if message == "!exit" {
				run = false
				writeToServerStr(client, "!exit")
			} else {
				writeToServerStr(client, username+": "+message)
			}
		} else {
			fin := <-ch
			if fin == 1 {
				break
			}
		}
	}
}
