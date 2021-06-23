# biliStreamClient
This package helps establish a websocket connection to the bilibili streaming server.
<br />bilibili直播弹幕的WebSocket协议分析请参考：https://blog.csdn.net/xfgryujk/article/details/80306776

<br />基于这个package，我写了一个朗读弹幕的网页应用
<br />讲解： https://www.bilibili.com/video/BV18X4y1G7Ks/
<br />试用： https://www.danmujun.online/

BiliClient is a struct with the following public methods
```
type BiliClient struct {
	serverConn      *websocket.Conn
	uid             int
	connected       bool
	roomID          int
	mutex           *sync.Mutex
	protocolVersion uint16
	Ch              chan PacketBody
}

func (bili *BiliClient) Connect(rid int) error
func (bili *BiliClient) Disconnect() error
```

After using BiliClient.Connect, a websocket connection between your machine and the Bilibili live-stream server is established. The BiliClient will listen to messages sent back from the live-stream server, parse it into a packBody, and put it into BiliClient.Ch.

We just need to fetch the packBody from BiliClient.Ch.

Currently packBody supports ParseDanmu(), ParseGift() and ParseGiftCombo(). You can also write custom parser to parse the data from the PackBody into your own data structure.

使用的时候我们只需要新建一个BiliClient，然后Connect()，然后从BiliClient.Ch里面不断取出PackBody即可。对于PackBody，目前支持parse 弹幕，礼物和礼物Combo， 分别对应DANMU_MSG，SEND_GIFT和COMBO_SEND。

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

