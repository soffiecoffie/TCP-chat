package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

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

// Writes to server from standard input and returns the sent message
func writeToServer(c net.Conn) string {
	reader := bufio.NewReader(os.Stdin)
	message, _ := reader.ReadString('\n') //returns err!=nil only when it doesnt end in \n
	// Remove '\n' from the message and the ascii 13 character that breaks everything.. 
	message = strings.TrimSuffix(message, "\n")
	message = strings.TrimSuffix(message, string(13))
	// Input the message in buf
	buf := []byte(message)
	fmt.Println("length of usr ", len(message))
	fmt.Println("length of buf ", len("toast"))
	fmt.Println(message)
	// Write to server
	_, err := c.Write(buf)
	if err != nil {
		// if err != io.EOF {
		// 	fmt.Println("Error with reading: ", err.Error())
		// }
		// break
		// return //y/n
		os.Exit(1)
	}
	return message
}

func main() {
	// client, err := net.Dial("tcp", "192.168.0.3:8080")
	// client, err := net.Dial("tcp", "192.168.0.3:0")
	client, err := net.Dial("tcp", "192.168.0.3:52429")
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
	fmt.Println("Username is: " + username)

	// for {
	// 	received = readFromCon(client)
	// 	fmt.Print(received)
	// }

}
