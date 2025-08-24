package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"time"
)

func handleConnection(codemap map[int]string, conn net.Conn) {
	defer conn.Close()
	fmt.Println("New connection from", conn.RemoteAddr())

	write_message := func (message string) error {
		_, err := conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Server write error:", err)
			return err
		}
		return nil
	}

	// Create client hostname pattern
    client_hostname_regex := regexp.MustCompile(`HELO (.+)`)
	// Create client FROM pattern
    client_from_regex := regexp.MustCompile(`MAIL FROM:<(.+)>`)
	// Create client TO pattern
    client_to_regex := regexp.MustCompile(`RCPT TO:<(.+)>`)
	// Create DATA <CR LF . CR LF> pattern
    crlf_regex := regexp.MustCompile(`\r\n[.]\r\n`)

	buf := make([]byte, 4096)
	recipients := []string{}
	sender := ""
	var payload strings.Builder
	sequence_ctr := 1
	write_message("220 SMTP postoffice")
	for {
		if sequence_ctr >= 5 {
			// we got to the end of our session with the client.
			// print out some info about the email before quitting
			fmt.Printf("SENDER: %s\n",sender)
			fmt.Println("RECIPIENTS:",recipients)
			fmt.Println("PAYLOAD:")
			fmt.Println(payload.String())
			time.Sleep(1 * time.Second)
			return
		}

		conn.SetDeadline(time.Now().Add(5 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Server connection closed:", err)
			return
		}
		data := string(buf[:n])
		//fmt.Printf("Server received: %s\n", data)

		if data == "QUIT" {
			fmt.Println("Server got quit message.")
			write_message(codemap[221]) // Bye
			sequence_ctr = 5
			continue
		} else if data == "DATA" {
			sequence_ctr = 3
		}

		fmt.Printf("sequence_ctr %d\n",sequence_ctr)

		if sequence_ctr == 1 {
			match := client_hostname_regex.FindStringSubmatch(data)
			client_hostname := match[1]
			write_message(fmt.Sprintf(codemap[251],client_hostname)) // Hello, glad to meet you
		} else if sequence_ctr == 2 {
			match := client_from_regex.FindStringSubmatch(data)
			if match != nil {
				sender = match[1]
			}

			match = client_to_regex.FindStringSubmatch(data)
			if match != nil {
				recipients = append(recipients, match[1])
			}

			write_message(codemap[252]) // Simple 'Ok' response to client data
		} else if sequence_ctr == 3 {
			write_message(codemap[354]) // End data with CR LF . CR LF
			sequence_ctr = 4
			continue
		} else if sequence_ctr == 4 { // Process payload data
			// Check and see if we got a CR LF . CR LF
			if crlf_regex.MatchString(data) {
				fmt.Println("Server got cr lf")
				write_message(codemap[253]) // Ok, queued as 12345
				continue
			}
			payload.WriteString(data)
			write_message(codemap[252])
		}

		// If we're not expecting a range of messages for the sending sequence,
		// move the sequence counter forward.
		if sequence_ctr != 2 && sequence_ctr != 4 {
			sequence_ctr += 1
		}
	}
}

func server(codemap map[int]string, port int) {
	// Listen on TCP port 25 on all interfaces
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d",port))
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Printf("Listening on port %d...\n",port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}
		// Handle each connection concurrently
		go handleConnection(codemap, conn)
		fmt.Println("Created connection")
	}
}

func client(address string) {
	/**/
	// Read the payload information from file
	email_payload, err := os.ReadFile("email_payload.txt")
	if err != nil {
		fmt.Println("Error reading email payload:", err)
		return
	}
	/**/

	buf := make([]byte, 4096)
	sender := "bob@example.org"
	payloads := []string{"HELO postoffice_client"}
	recipients := [2]string{"alice@example.com", "theboss@example.com"}

	from_message := fmt.Sprintf("MAIL FROM:<%s>",sender)
	payloads = append(payloads,from_message)

	for _,itm := range recipients {
		payloads = append(payloads,fmt.Sprintf("RCPT TO:<%s>",itm))
	}
	payloads = append(payloads,"DATA")
	payloads = append(payloads,string(email_payload))
	payloads = append(payloads,"CRLF")
	payloads = append(payloads,"QUIT")

	// Create the connection
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Client connected to:", address)
	// Read the server's initial response
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Client read error:", err)
		return
	}
	fmt.Println("Client received connection response:", string(buf[:n]))

	// By this point, we're ready to actually begin interacting with the server/thread
	for _,payload := range payloads {
		if payload == "CRLF" {
			_, err = conn.Write([]byte{0xD, 0xA, 0x2E, 0xD, 0xA})
			if err != nil {
				fmt.Println("Client write error:", err)
				return
			}
		} else {
			_, err = conn.Write([]byte(payload))
			if err != nil {
				fmt.Println("Client write error:", err)
				return
			}
		}

		fmt.Println("Client sent", payload)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Client read error:", err)
			return
		}

		fmt.Println("Client received:", string(buf[:n]))
		time.Sleep(1 * time.Second)
	}
}

func main() {
	codemap := make(map[int]string)
	port := 25

	codemap[220] = "220 HELO postoffice"
	codemap[221] = "221 Bye"
	// new connection
	codemap[251] = "250 Hello %s, I am glad to meet you"
	// ok response
	codemap[252] = "250 Ok"
	// new email message queued
	codemap[253] = "250 Ok: queued as 12345"
	codemap[354] = "354 End data with <CR><LF>.<CR><LF>"

	go server(codemap,port)
	client(fmt.Sprintf("localhost:%d",port))
}

/*
Sequence:

220 smtp.example.com ESMTP Postfix
C: HELO relay.example.org
S: 250 Hello relay.example.org, I am glad to meet you
C: MAIL FROM:<bob@example.org>
S: 250 Ok
C: RCPT TO:<alice@example.com>
S: 250 Ok
C: RCPT TO:<theboss@example.com>
S: 250 Ok
C: DATA
S: 354 End data with <CR><LF>.<CR><LF>
C: <email_payload.txt>
C: .
S: 250 Ok: queued as 12345
C: QUIT
S: 221 Bye
{The server closes the connection}
*/
