package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// fungsi menerima pesan dari server
func receiveMessages(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for { //looping digunakan untuk menerima pesan secara terus menerus dari server
		message, err := reader.ReadString('\n') //membaca pesan hingga menemukan newline
		if err != nil {
			fmt.Println("Connection closed")
			os.Exit(1)
		}
		fmt.Print(message) //tampilkan pesan di terminal
	}
}

func main() {
	conn, err := net.Dial("tcp", ":9090")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot connect to server: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()
	fmt.Println("Connected to server!")

	// go routine untuk menerima pesan
	go receiveMessages(conn)

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter name: ") // prompt memasukan nama (tidak di handle di server karena selalu prompt pertama)

	// input pesan, client hanya mengirimkan pesan tanpa mengetahui context dari pesan itu, alur dan logic diatur semua di server
	for {
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)
		if message != "" {
			fmt.Fprintln(conn, message)
		}
	}
}
