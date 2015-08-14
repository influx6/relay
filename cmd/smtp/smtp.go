package main

import (
	"fmt"
	"log"
	"net/smtp"

	"github.com/influx6/flux"
	somtp "github.com/influx6/smtp"
)

func main() {

	mc := somtp.MetaConfig{Hostname: "localhost"}

	// so, err := somtp.NewSMTP("0.0.0.0:25", ":3040", nil)
	so, err := somtp.NewSMTP(mc, "mailtrap.io:2525", ":3040", nil)

	if err != nil {
		log.Fatal(err)
	}

	flux.GoDefer("SMTPServiceServe", so.Serve)

	auth := smtp.PlainAuth("", "4200103e6055279ac", "52688238e8872b", "")

	// Connect to the remote SMTP server.
	// c, err := smtp.Dial("mail.example.com:25")
	// c, err := smtp.Dial("localhost:431")
	c, err := smtp.Dial(":3040")

	if err != nil {
		log.Fatal("Connecting to SMTP:", err)
	}

	if err := c.Auth(auth); err != nil {
		log.Fatal("Auth:Error ", err)
	}

	// Set the sender and recipient first
	if err := c.Mail("sender@localmail.org"); err != nil {
		log.Fatal("Setting sender", err)
	}
	if err := c.Rcpt("recipient@localmail.net"); err != nil {
		log.Fatal("Setting receiver", err)
	}

	// Send the email body.
	wc, err := c.Data()

	if err != nil {
		log.Fatal("Sending data error: ", err)
	}

	log.Printf("Writing data to io.Writer")
	_, err = fmt.Fprintf(wc, "This is the email body")
	log.Printf("Writen data to io.Writer")

	if err != nil {
		log.Fatal("Writing body", err)
	}

	log.Printf("Closing data to io.Writer")
	err = wc.Close()
	log.Printf("Closed data to io.Writer")

	if err != nil {
		log.Fatal("Closing body", err)
	}

	// Send the QUIT command and close the connection.
	log.Printf("Killing data to io.Writer")
	err = c.Quit()
	log.Printf("Killed data to io.Writer")

	if err != nil {
		log.Fatal(err)
	}
}

// func ExamplePlainAuth() {
// 	// Set up authentication information.
// 	auth := smtp.PlainAuth("", "user@example.com", "password", "mail.example.com")
//
// 	// Connect to the server, authenticate, set the sender and recipient,
// 	// and send the email all in one step.
// 	to := []string{"recipient@example.net"}
// 	msg := []byte("This is the email body.")
// 	err := smtp.SendMail("mail.example.com:25", auth, "sender@example.org", to, msg)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }
