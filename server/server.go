package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/faiface/pixel"
)

type player struct {
	pos         pixel.Vec
	canMove     bool
	currDir     int // Current direction of moving, 0 == up, 1 == right, 2 == down, 3 == left
	clientID    int
	isConnected bool
}

const (
	maxConnections = 2 //Maximum allowed connections
)

var (
	connections [maxConnections]net.Addr //Create a net address array of maxConnections
)

func main() {
	fmt.Println("Initialized The Boys Make A Game The Movie - Server")
	ln, _ := net.ListenPacket("udp", ":5669") //Begin listening
	fmt.Println("Starting server on port:5669")

	var players []player

	for i := 0; i < maxConnections; i++ { //Create a list of empty players
		players = append(players, player{pixel.ZV, true, 0, i, false}) //<---- Seperate this in a different method at some point in the future
	}

	go handleIncomingPackets(ln, players)
	//	var printChannelString = make(chan string)

	//List of players
	//go handleConnections(ln, printChannelString, players) //Create a goroutine handling connections
	x := 0.0

	for { // no you
		fmt.Println(x)
		<-time.After(1 * time.Second)
		x++
	}
}

/**func handleConnections(ln net.Listener, cs chan string, players []player) {
	for i := 0; i < maxConnections; i++ { //Create a list of empty players
		players = append(players, player{pixel.ZV, true, 0, i, false})
	}
	for {
		conn, err := ln.Accept()
		connections = append(connections, conn)
		currID := maxConnections + 1 //Used to detect extra connections
		for i := 0; i < len(players); i++ {
			if !players[i].isConnected { //There is a free slot available
				currID = i
				players[i].isConnected = true
				fmt.Println("Player connected")
				registerClient(int8(currID), conn)
				go handleIncomingPackets(conn, int8(currID), players) //begin goroutine to handle incoming packets
				i = len(players)
			}
		}
		if currID > maxConnections { //Too many players! Terminate the connection
			conn.Close()
			fmt.Println("Terminated excess connection")
		}

		if err != nil { //D'aw fuck I can't believe you've done this
			fmt.Println("Something broke:" + err.Error())
		}

	}

}**/
func handleIncomingPackets(conn net.PacketConn, players []player) {
	for {
		recvBuf := make([]byte, 1024) //Packet Recieved buffer
		playerID := int8(0)
		size, addr, err := conn.ReadFrom(recvBuf) //Size of packet
		for i := 0; i < len(connections); i++ {
			if connections[i] != nil {
				if connections[i].String() == addr.String() {
					playerID = int8(i)
					i = len(connections)
				}
			}
		}
		if err != nil {
			//conn.Close()
			log.Println(err)
			handleDisconnection(playerID, players)
			return
		}
		_ = size
		packetType := recvBuf[0] //First byte of packet
		switch packetType {      //Execute different functions depending on packet type
		case 0: //Player Move packet
			handlePacketMove(conn, playerID, players, recvBuf)
			break
		case 2: //Player info packet
			handlePacketPlayerInfo(conn, playerID, players, recvBuf)
			break
		case 3: //Connect/Disconnect packet
			handlePacketConnect(addr, playerID, players, recvBuf, conn)
		}

	}
}
func playerCalculations(clientID int8, players []player) {
	for {
		if !players[clientID].isConnected {
			return
		}
	}
}
func handlePacketConnect(addr net.Addr, playerID int8, players []player, recvBuf []byte, packetConn net.PacketConn) {
	log.Println("Recieved a disconnection/connection packet from: ", addr.String(), " the type being: ", recvBuf[1])
	if recvBuf[1] == byte(1) {

		currID := maxConnections + 1 //Used to detect extra connections
		for i := 0; i < len(players); i++ {
			if !players[i].isConnected { //There is a free slot available
				connections[i] = addr
				currID = i
				players[i].isConnected = true
				fmt.Println("Player connected")
				registerClient(int8(currID), packetConn, addr, players)
				i = len(players)
			}
		}
		if currID > maxConnections { //Too many players! Terminate the connection
			//	conn.Close()
			fmt.Println("Terminated excess connection")
		}
	} else {
		handleDisconnection(playerID, players)
	}
}
func handlePacketPlayerInfo(conn net.PacketConn, playerID int8, players []player, recvBuf []byte) {
	//log.Println("asdsad")
	players[playerID].currDir = int(recvBuf[1])
	packet := []byte{byte(2)}
	packet = append(packet, byte(playerID))
	packet = append(packet, recvBuf[1])
	relayPacketsToAllConnections(packet, conn)
}
func handleDisconnection(connectionID int8, players []player) {
	log.Println("Player: ", connectionID, " has disconnected")

	players[connectionID].isConnected = false //Set the player as disconnected, used when seeing if a free player exists
	connections[connectionID] = nil
	log.Println("Connections remaining: ", len(connections))
}
func handlePacketMove(conn net.PacketConn, playerID int8, players []player, recvBuf []byte) {
	x := recvBuf[1:9] //<-- this is where it is hardcoded future me 
	y := recvBuf[9:]
	floatX := byte2Float64(x)                             //Convert x byte array into float
	floatY := byte2Float64(y)                             //Convert y  byte array into float
	players[playerID].pos = pixel.V(floatX, floatY)       //Save the local position serverside
	sendPacketMove(conn, playerID, floatX, floatY, false) //Begin the relaying process
}
func sendPacketMove(conn net.PacketConn, playerID int8, x float64, y float64, teleport bool) {
	//	fmt.Println(playerID)
	byteTeleport := byte(0)
	if teleport {
		byteTeleport = byte(1)
	}
	packet := []byte{byte(0)}               //Packet ID
	packet = append(packet, byte(playerID)) //Add the playerID
	packet = append(packet, byteTeleport)
	packet = append(packet, float642Byte(x)...) //Add X movement
	packet = append(packet, float642Byte(y)...) //Add Y Movement
	//conn.Write(packet)
	relayPacketsToAllConnections(packet, conn) //Send to everyone
}
func registerClient(clientID int8, conn net.PacketConn, addr net.Addr, players []player) { //Tells the client what ID it has
	sendPacketMove(conn, clientID, players[clientID].pos.X, players[clientID].pos.Y, true)
	packet := []byte{byte(1)}
	packet = append(packet, byte(clientID))
	conn.WriteTo(packet, addr)
}
func handleStringPacket(conn net.PacketConn, cs chan string) { //Deprecated, may be used in future in case of sending chat messages
	for {
		//	message, _ := bufio.NewReader(conn).ReadString('\n')
		//	cs <- fmt.Sprintf("Packet Recieved, %s", string(message))
	}
}
func relayPacketsToAllConnections(packet []byte, conn net.PacketConn) { //Sends a packet to all currently connected clients
	for i := 0; i < len(connections); i++ {
		if connections[i] != nil {
			conn.WriteTo(packet, connections[i])
		}
	}
}
