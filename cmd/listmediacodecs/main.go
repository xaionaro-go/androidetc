package main

import (
	"encoding/json"

	"github.com/xaionaro-go/androidetc"
)

func main() {
	mediaCodecs, err := androidetc.ParseMediaCodecs()
	if err != nil {
		panic(err)
	}

	for _, mc := range mediaCodecs {
		b, err := json.Marshal(mc)
		if err != nil {
			panic(err)
		}

		println(string(b))
	}
}
