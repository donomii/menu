package menu

import (
	"log"

	"fmt"

	//"github.com/donomii/goof"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

//username := goof.CatFile("username")
//password := goof.CatFile("password")
func GetSummaries(maxItems int, username, password string) [][]string {
	max := uint32(maxItems)
	var out [][]string
	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.DialTLS("imap.gmail.com:993", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login(string(username), string(password)); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	/*
		log.Println("Mailboxes:")
		for m := range mailboxes {
			log.Println("* " + m.Name)
		}
	*/
	if err := <-done; err != nil {
		log.Fatal(err)
	}

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}
	//log.Println("Flags for INBOX:", mbox.Flags)

	log.Println("Last", max, "messages:")
	// Get the last 4 messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages >= max {
		// We're using unsigned integers here, only substract if the result is > 0
		from = mbox.Messages - max - 1
		//from = mbox.Messages - 3
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done = make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchUid, imap.FetchRFC822Text, imap.FetchBody, imap.FetchEnvelope, imap.FetchBodyStructure, imap.FetchFlags}, messages)
	}()

	for msg := range messages {
		data := fmt.Sprintf("%+v, %+v", msg.Envelope, msg.BodyStructure)
		for _, v := range msg.Body {
			//fmt.Println("Body: '", k, "'", v)
			data = fmt.Sprintf("%v", v)
		}
		//fmt.Println(data)
		out = append(out, []string{msg.Envelope.Subject, data})
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")
	return out
}
