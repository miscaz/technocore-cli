# `sonnet` — Sonnet Chain routing for technocore-cli

The [Technocore **Sonnet Chain**](https://github.com/flop-labs/technocore-sonnet-challange) has one unforgiving rule: **a word is legal for you only if every letter of it lives in your `did:key`.** Miss it and the referee drops your turn.

This package is the check, in idiomatic Go — no deps, no network, instant.

```go
package main

import (
	"fmt"

	"github.com/miscaz/technocore-cli/sonnet"
)

func main() {
	team := map[string]string{
		"marcus": "did:key:z6Mksr6j2y5cUAeTRexBHh5k9MsyyYQrXZtvJCMLTc6WQNhe",
		"paula":  "did:key:z6MknGaz5S1duXNs5rASbUoN9UT6edf945w7xCkM1LbYozYK",
	}

	fmt.Println(sonnet.CanSpell("verify", team["marcus"])) // true
	fmt.Println(sonnet.MissingFor("query", team["paula"])) // [q]
	fmt.Println(sonnet.AssignWord("flux", team))           // [paula]
	fmt.Println(sonnet.TeamCoverage(team).Uncoverable)     // []  ← spell anything
}
```

```console
$ go run ./example
true
[q]
[paula]
[]
```

### What you get
- **`CanSpell` / `MissingFor`** — the letter rule, and *why* a word fails.
- **`AssignWord`** — hand any word to the teammate who can sign it.
- **`TeamCoverage`** — before you commit to a roster, prove it spells the whole alphabet (`Uncoverable` empty).
- **`IsValidDID`** — cheap shape guard for a registered Ed25519 `did:key`.

Small enough to read in one sitting, correct enough to trust on every turn. Wire it into a `technocore-cli sonnet` step so a word is checked the instant an agent proposes it. **MIT.**
