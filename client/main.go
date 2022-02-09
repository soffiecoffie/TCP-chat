package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

func readFromCon(buf []byte, c net.Conn) string {
	_, err := c.Read(buf)
	if err != nil {
		if err != io.EOF {
			fmt.Println("Error with reading: ", err.Error())
		}
		os.Exit(1)
		fmt.Println("err = io.eof")
	}
	return string(buf)
}
func main() {
	// client, err := net.Dial("tcp", "192.168.0.3:8080")
	// client, err := net.Dial("tcp", "192.168.0.3:0")
	client, err := net.Dial("tcp", "192.168.0.3:49916")
	if err != nil {
		fmt.Println("Error with connecting: ", err.Error())
		return
	}
	defer client.Close()
	fmt.Println("Local addr: " + client.LocalAddr().String())
	fmt.Println("Remote addr: " + client.RemoteAddr().String())

	go func(c net.Conn) {
		buff := make([]byte, 1024)
		for {
			reader := readFromCon(buff, c)
			fmt.Print(reader)
		}
		// _, err = c.Read(buff)
		// if err != nil {
		// 	if err != io.EOF {
		// 		fmt.Println("Error with reading: ", err.Error())
		// 	}
		// 	// break
		// 	// return
		// 	fmt.Println("err = io.eof")
		// }
		// fmt.Println(string(buff))
	}(client)

	//Asking the client to create a username
	// client.Write([]byte("Welcome to the server! Create a username: "))
	reader := bufio.NewReader(os.Stdin) //reader.ReadString('\n') //returns err!=nil only when it doesnt end in \n
	// fmt.Print(reader.ReadString('\n'))
	buf := make([]byte, 1024)
	_, err = reader.Read(buf)
	if err != nil {
		if err != io.EOF {
			fmt.Println("Error with reading: ", err.Error())
		}
		return // y/n
	}

	_, err = client.Write(buf)
	if err != nil {
		// if err != io.EOF {
		// 	fmt.Println("Error with reading: ", err.Error())
		// }
		return // y/n
	}
}
