// Package sonnet provides helpers for the Technocore "Sonnet Chain" challenge,
// where a team co-writes a 14-line sonnet one signed word per turn and every
// letter of a word must appear in that contributor's did:key.
//
// It answers the routing question a team has on every turn: given a word, which
// teammate is allowed to sign it? Pure standard library, no allocations beyond
// the small letter sets, safe to embed in technocore-cli or import directly.
package sonnet

import (
	"regexp"
	"sort"
	"strings"
)

var didRe = regexp.MustCompile(`^did:key:z6Mk[1-9A-HJ-NP-Za-km-z]{44}$`)

// IsValidDID reports whether did has the registered Ed25519 did:key shape.
func IsValidDID(did string) bool { return didRe.MatchString(did) }

// letterSet returns the set of a–z letters in s, lowercased.
func letterSet(s string) map[rune]bool {
	set := make(map[rune]bool)
	for _, r := range strings.ToLower(s) {
		if r >= 'a' && r <= 'z' {
			set[r] = true
		}
	}
	return set
}

// AllowedLetters returns the a–z letters a did:key can spell with.
func AllowedLetters(did string) map[rune]bool { return letterSet(did) }

// CanSpell reports whether every letter of word appears in did.
func CanSpell(word, did string) bool {
	allowed := letterSet(did)
	for r := range letterSet(word) {
		if !allowed[r] {
			return false
		}
	}
	return true
}

// MissingFor returns the sorted letters word needs that did cannot provide.
// An empty result means the word is playable by that DID.
func MissingFor(word, did string) []string {
	allowed := letterSet(did)
	var miss []string
	for r := range letterSet(word) {
		if !allowed[r] {
			miss = append(miss, string(r))
		}
	}
	sort.Strings(miss)
	return miss
}

// AssignWord returns the sorted labels of teammates who may legally play word.
// team maps a label to that teammate's did:key.
func AssignWord(word string, team map[string]string) []string {
	var who []string
	for label, did := range team {
		if CanSpell(word, did) {
			who = append(who, label)
		}
	}
	sort.Strings(who)
	return who
}

// Coverage summarizes a team's collective letter reach.
type Coverage struct {
	PlayableUnion    []string            // letters at least one member can spell
	Uncoverable      []string            // a–z letters NO member can spell (empty == good)
	PerMemberMissing map[string][]string // per-label gaps
}

// TeamCoverage computes letter coverage across a team.
func TeamCoverage(team map[string]string) Coverage {
	union := make(map[rune]bool)
	per := make(map[string][]string)
	for label, did := range team {
		allowed := letterSet(did)
		var gaps []string
		for r := 'a'; r <= 'z'; r++ {
			if !allowed[r] {
				gaps = append(gaps, string(r))
			}
			if allowed[r] {
				union[r] = true
			}
		}
		per[label] = gaps
	}
	var play, unc []string
	for r := 'a'; r <= 'z'; r++ {
		if union[r] {
			play = append(play, string(r))
		} else {
			unc = append(unc, string(r))
		}
	}
	return Coverage{PlayableUnion: play, Uncoverable: unc, PerMemberMissing: per}
}
