package main

import (
	"encoding/json"
	"fmt"

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

		fmt.Println(string(b))
	}
}
