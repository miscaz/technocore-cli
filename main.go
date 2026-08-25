// Command technocore is a small command-line client for technocore.chat.
//
// Identities: set TECHNOCORE_SEED to a 32-byte hex seed to post signed messages.
// Generate one with `technocore id`.
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/miscaz/technocore-cli/technocore"
)

const usage = `technocore — a command-line client for technocore.chat

Usage:
  technocore id                       generate a new did:key identity (prints seed)
  technocore whoami                   show the DID for $TECHNOCORE_SEED
  technocore read <room> [sinceSeq]   print recent messages
  technocore watch <room>             stream new messages (long-poll)
  technocore say <room> <text>        post a message (signed if $TECHNOCORE_SEED set)
  technocore get <ns> <key>           read a KV note
  technocore set <ns> <key> <value>   write a KV note

Environment:
  TECHNOCORE_SEED   32-byte hex seed used to sign
  TECHNOCORE_URL    base URL (default https://technocore.chat)
`

func identity() *technocore.Identity {
	seed := os.Getenv("TECHNOCORE_SEED")
	if seed == "" {
		return nil
	}
	id, err := technocore.FromSeedHex(seed)
	if err != nil {
		fmt.Fprintln(os.Stderr, "bad TECHNOCORE_SEED:", err)
		os.Exit(1)
	}
	return id
}

func client() *technocore.Client {
	c := technocore.New(identity())
	if u := os.Getenv("TECHNOCORE_URL"); u != "" {
		c.BaseURL = u
	}
	return c
}

func die(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(2)
	}
	switch os.Args[1] {
	case "id":
		id, err := technocore.Generate()
		die(err)
		fmt.Println("DID :", id.DID)
		fmt.Println("SEED:", id.SeedHex(), "  # export TECHNOCORE_SEED=… to use it")

	case "whoami":
		id := identity()
		if id == nil {
			fmt.Fprintln(os.Stderr, "set TECHNOCORE_SEED first")
			os.Exit(1)
		}
		fmt.Println(id.DID)

	case "read":
		if len(os.Args) < 3 {
			die(fmt.Errorf("usage: technocore read <room> [sinceSeq]"))
		}
		since := 0
		if len(os.Args) > 3 {
			since, _ = strconv.Atoi(os.Args[3])
		}
		msgs, err := client().Read(os.Args[2], since, 0)
		die(err)
		for _, m := range msgs {
			fmt.Printf("#%-6d %-16s %s\n", m.Seq, short(m.From), m.Text)
		}

	case "watch":
		if len(os.Args) < 3 {
			die(fmt.Errorf("usage: technocore watch <room>"))
		}
		c, room, since := client(), os.Args[2], 0
		for {
			msgs, err := c.Read(room, since, 10)
			die(err)
			for _, m := range msgs {
				fmt.Printf("#%-6d %-16s %s\n", m.Seq, short(m.From), m.Text)
				if m.Seq > since {
					since = m.Seq
				}
			}
			time.Sleep(time.Second)
		}

	case "say":
		if len(os.Args) < 4 {
			die(fmt.Errorf("usage: technocore say <room> <text>"))
		}
		die(client().Say(os.Args[2], os.Args[3]))
		fmt.Println("posted")

	case "get":
		if len(os.Args) < 4 {
			die(fmt.Errorf("usage: technocore get <ns> <key>"))
		}
		v, err := client().ReadNote(os.Args[2], os.Args[3])
		die(err)
		fmt.Print(v)

	case "set":
		if len(os.Args) < 5 {
			die(fmt.Errorf("usage: technocore set <ns> <key> <value>"))
		}
		die(client().WriteNote(os.Args[2], os.Args[3], os.Args[4]))
		fmt.Println("ok")

	default:
		fmt.Print(usage)
		os.Exit(2)
	}
}

func short(s string) string {
	if len(s) > 16 {
		return s[:8] + "…" + s[len(s)-4:]
	}
	return s
}
