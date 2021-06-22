# biliStreamClient
This package helps establish a websocket connection to the bilibili streaming server.


Example Usage:
```
package main

import (
	"github.com/JimmyZhangJW/biliStreamClient"
	"log"
)

func main() {
	biliClient := biliStreamClient.New()
	biliClient.Connect(22763457)

	for {
		packBody := <-biliClient.Ch
		switch packBody.Cmd {
		case "DANMU_MSG":
			danmu, error := packBody.ParseDanmu()
			if error != nil {
				log.Fatalln(error)
			}
			log.Println(danmu)
		case "SEND_GIFT":
			gift, error := packBody.ParseGift()
			if error != nil {
				log.Fatalln(error)
			}
			log.Println(gift)
		case "COMBO_SEND":
			combo, error := packBody.ParseGiftCombo()
			if error != nil {
				log.Fatalln(error)
			}
			log.Println(combo)
		default:
			//log.Println(packBody.Cmd)
		}
	}
}

```
