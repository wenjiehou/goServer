package random_name

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func random(min, max int) int {
	return rand.Intn(max-min) + min
}

func GetRandomName() string {
	n1 := rand.Intn(len(xing))
	n2 := rand.Intn(len(ming))
	s1 := xing[n1]
	s2 := ming[n2]
	return s1 + s2
}
