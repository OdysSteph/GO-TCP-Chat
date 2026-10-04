package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

type Client struct { //client sebagai struct
	conn net.Conn
	name string
	room string
}

var (
	clients = make(map[net.Conn]*Client)                                     //map digunakan untuk menyimpan seluruh client dengan index net.Conn
	rooms   = []string{"general", "games", "hobby", "study", "chill", "qna"} //rooms berisi seluruh room yang dapat digunakan untuk chat
	mutex   = &sync.Mutex{}                                                  //mutex digunakan untuk akses konkuren ke shared resource yaitu clients dan room
)

func listRooms() string { // menampilkan list room yang client dapat bergabung
	result := "\nAvailable rooms [/leave to exit room]:\n"
	for _, room := range rooms {
		result += fmt.Sprintf(". %s\n", room)
	}
	result += "Choose room:\n"
	return result
}

func broadcastMessage(message string, sender net.Conn, room string) {
	mutex.Lock()         //mengunci akses ke shared resource(clients)
	defer mutex.Unlock() //mutex akan di unlock ketika fungsi selesai di execute

	for conn, client := range clients { //looping ke seluruh user yang ada di clients, untuk mengirimkan pesan
		if conn != sender && client.room == room { //tambahkan kondisi untuk mengirim pesan ke seluruh client yang ada di room yang sama dan tidak mengirim ke diri sendiri
			fmt.Fprint(conn, message) //kirim pesan ke client
		}
	}
}

func isNameTaken(name string) bool { //func mengecek apakah nama client sudah ada yang pakai
	mutex.Lock() // mutex karena setiap koneksi mengakses clients
	defer mutex.Unlock()

	loweCaseName := strings.ToLower(name) //ubah nama menjadi lowercase karena capitalization tidak berpengaruh
	for _, client := range clients {      // check satu per satu di clients
		if strings.ToLower(client.name) == loweCaseName { // return true jika ditemukan nama yang sama
			return true
		}
	}
	return false
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	var name string // get nama client
	for {           // loop sampai client memiliki nama yg unik
		nameInput, err := reader.ReadString('\n')
		if err != nil { //jika error (misal forced close) akan return dan func terhentikan
			fmt.Printf("Failed to read name: %v\n", err)
			return
		}
		name = strings.TrimSpace(nameInput) //hapus space di depan dan belakang nama
		if isNameTaken(name) {              // jika nama sudah ada yang pakai, loop kembali dan client memasukan nama baru
			fmt.Fprintf(conn, "Name unavailable, enter a different name:\n") // prompt masukan nama baru
		} else { //jika unik, hentikan loop
			break
		}
	}
	// init client baru, mula2 belum memiliki room
	client := &Client{
		conn: conn,
		name: name,
		room: "",
	}

	mutex.Lock() //mutex karena banyak conn dapat mengakses clients sekaligus
	clients[conn] = client
	mutex.Unlock()

	roomSelect: for { // loop untuk pilih room
		fmt.Fprintf(conn, "%s", listRooms())     // kirim room list ke client
		roomName, err := reader.ReadString('\n') // get nama room
		if err != nil {
			fmt.Printf("Failed to read room name: %v\n", err)
			break
		}
		roomName = strings.TrimSpace(roomName)
		validRoom := false // validasi room, check input room di rooms
		for _, room := range rooms {
			if strings.EqualFold(room, roomName) {
				roomName = room
				validRoom = true
				break
			}
		}
		if !validRoom { // jika room yang diinput client tidak ada, client diminta memasukan nama room kembali
			fmt.Fprintf(conn, "\nInvalid room name, try again\n")
			continue
		}
		client.room = roomName //set room milik client

		fmt.Printf("%s joined %s room\n", name, roomName)
		fmt.Fprintf(conn, "\nSuccesfully joined %s.\n\nNew chat:\n", roomName)
		broadcastMessage(fmt.Sprintf("%s just joined.\n", name), conn, roomName) //broadcast ke room tersebut bahwa client telah join

		for { // menerima message client terus menerus
			message, err := reader.ReadString('\n')
			if err != nil {
				break roomSelect
			}
			message = strings.TrimSpace(message)
			if message == "/leave" { // message /leave menandakan client ingin keluar room
				fmt.Fprintf(conn, "Left %s.\n", client.room)                                         // pesan untuk client, menandakan bahwa ia sudah keluar room
				broadcastMessage(fmt.Sprintf("%s left %s.\n", name, client.room), conn, client.room) //notifikasi bahwa client keluar room di room tersebut
				client.room = ""                                                                     // hilangkan room dari client
				break                                                                                //hentikan loop menerima message, berpindah ke loop untuk memilih room
			}
			if message != "" { //jika message tidak kosong, broadcast message pada room tersebut
				broadcastMessage(fmt.Sprintf("%s: %s\n", name, message), conn, client.room)
			}
		}
	}

	mutex.Lock()          //mutex karena mengakses clients
	delete(clients, conn) // hapus client jika disconnect
	mutex.Unlock()

	if client.room != "" { // jika client berada dalam room, notifikasi client meninggalkan room tersebut
		fmt.Printf("%s left %s\n", name, client.room) // notif di server
		broadcastMessage(fmt.Sprintf("%s left %s.\n", name, client.room), conn, client.room)
	}
}

func main() {
	ln, err := net.Listen("tcp", ":9090")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to listen: %v\n", err)
		os.Exit(1)
	}
	defer ln.Close() //listener akan ditutup ketika fungsi main telah di execute
	fmt.Println("Listening on port 9090")
	fmt.Println("Available rooms:", rooms)

	for {
		conn, err := ln.Accept() //menunggu dan menerima koneksi dari client
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to receive connection: %v\n.", err)
			continue
		}
		go handleClient(conn) //menggunakan go routine untuk menangani client secara independen sehingga dapat dikerjakan konkuren(multi thread)
	}
}
