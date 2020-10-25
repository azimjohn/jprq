package jprq

import (
	cryptorand "crypto/rand"
	"fmt"
	"math/rand"
	"time"
)

func generateToken() string {
	b := make([]byte, 32)
	cryptorand.Read(b)
	return fmt.Sprintf("%x", b)
}

func getRandomAdj() string {
	adjs := []string{
		"amazing", "ambitious", "amusing", "awesome",
		"brave", "bright", "broad-minded",
		"calm", "clever", "charming", "considerate", "confident", "courageous", "creative",
		"dazzling", "decisive", "determined", "diligent", "disciplined",
		"eager", "easygoing", "emotional", "energetic", "enthusiastic", "enchanting",
		"fabulous", "faithful", "fantastic", "fearless", "forceful", "frank", "friendly", "funny",
		"generous", "glorious", "gentle", "good",
		"hard-working", "helpful", "honest", "humorous",
		"imaginative", "independent", "ingenious", "intellectual", "intelligent", "intuitive", "inventive",
		"kind", "loving", "loyal", "modest", "nice", "optimistic",
		"passionate", "patient", "perfect", "persistent", "pioneering", "polite", "powerful", "practical",
		"quick-witted", "quiet",
		"rational", "reliable", "reserved", "resourceful", "romantic",
		"smart", "shy", "sincere", "sociable", "sympathetic",
		"talented", "thoughtful", "understanding", "versatile",
		"warmhearted", "wise", "willing", "witty", "wonderful",
	}
	rand.Seed(time.Now().Unix())
	return adjs[rand.Intn(len(adjs))]
}
