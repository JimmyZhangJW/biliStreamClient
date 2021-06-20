package biliStreamClient

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	roomInitUrl    string = "http://api.live.bilibili.com/room/v1/Room/room_init?id=%d"
	serverAddr     string = "broadcastlv.chat.bilibili.com:2244"
	chanBufferSize int    = 200
)

type roomInitResult struct {
	Code int `json:"code"`
	Data struct {
		Encrypted   bool `json:"encrypted"`
		HiddenTill  int  `json:"hidden_till"`
		IsHidden    bool `json:"is_hidden"`
		IsLocked    bool `json:"is_locked"`
		LockTill    int  `json:"lock_till"`
		NeedP2p     int  `json:"need_p2p"`
		PwdVerified bool `json:"pwd_verified"`
		RoomID      int  `json:"room_id"`
		ShortID     int  `json:"short_id"`
		UID         int  `json:"uid"`
	} `json:"data"`
	Message string `json:"message"`
	Msg     string `json:"msg"`
}

type PacketBody struct {
	Cmd   string                 `json:"cmd"`
	Info  []interface{}          `json:"info"`
	Data  map[string]interface{} `json:"data"`
	Count int                    `json:"count"`
}

type Packet struct {
	packetLen int
	headerLen int
	version   int
	op        int
	seq       int
	body      []PacketBody
}

type BiliClient struct {
	serverConn      *websocket.Conn
	uid             int
	connected       bool
	roomID          int
	mutex           *sync.Mutex
	protocolVersion uint16
	Ch              chan PacketBody
}

func New() *BiliClient {
	return &BiliClient{
		serverConn:      nil,
		uid:             0,
		connected:       false,
		roomID:          0,
		mutex:           &sync.Mutex{},
		protocolVersion: 1,
		Ch:              make(chan PacketBody, chanBufferSize), //Give a little bit of buffer to prevent blocking in case of reading slower than writing
	}
}

func getRealRoomID(rid int) (realID int, err error) {
	resp, err := http.Get(fmt.Sprintf(roomInitUrl, rid))
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()
	var res roomInitResult
	jBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	if err = json.Unmarshal(jBytes, &res); err != nil {
		return 0, err
	}
	if res.Code == 0 {
		return res.Data.RoomID, nil
	}
	return 0, fmt.Errorf(res.Message)
}

func decode(blob []byte) (Packet, error) {
	result := Packet{
		packetLen: int(binary.BigEndian.Uint32(blob[0:4])),
		headerLen: int(binary.BigEndian.Uint16(blob[4:6])),
		version:   int(binary.BigEndian.Uint16(blob[6:8])),
		op:        int(binary.BigEndian.Uint32(blob[8:12])),
		seq:       int(binary.BigEndian.Uint32(blob[12:16])),
		body:      make([]PacketBody, 0),
	}

	if result.op == 5 {
		offset := 0
		for offset < len(blob) {
			packetLen := int(binary.BigEndian.Uint32(blob[offset : offset+4]))
			if result.version == 2 {
				// If the data is zipped by zlib, we need to first unzip it
				data := blob[offset+result.headerLen : offset+packetLen]
				r, err := zlib.NewReader(bytes.NewReader(data))
				if err != nil {
					return Packet{}, err
				}

				var newBlob []byte
				if newBlob, err = ioutil.ReadAll(r); err != nil {
					return Packet{}, err
				}

				if err = r.Close(); err != nil {
					return Packet{}, err
				}

				// Then read the content of decompressed child
				var child Packet
				if child, err = decode(newBlob); err != nil {
					return Packet{}, err
				}
				result.body = append(result.body, child.body...)
			} else {
				// If the data is not zipped
				data := blob[offset+result.headerLen : offset+packetLen]
				var body PacketBody
				if err := json.Unmarshal(data, &body); err != nil {
					return Packet{}, err
				}
				result.body = append(result.body, body)
			}
			offset += packetLen
		}
	} else if result.op == 3 {
		result.body = append(result.body, PacketBody{
			Cmd:   "COUNTS_UPDATE",
			Count: int(binary.BigEndian.Uint32(blob[16:20])),
		})
	}

	return result, nil
}

// heartbeatLoop keep heartbeat every 5 seconds with bilibili live-stream server and stay online
func (bili *BiliClient) heartbeatLoop() {
	for bili.CheckConnect() {
		err := bili.sendSocketData(0, 16, bili.protocolVersion, 2, 1, "")
		if err != nil {
			bili.SetConnect(false)
			log.Printf("heartbeatError:%s\r\n", err.Error())
			return
		}
		time.Sleep(time.Second * 5)
	}
}

// sendSocketData: write message to the bilibili server
func (bili *BiliClient) sendSocketData(packetlength uint32, magic uint16, ver uint16, action uint32, param uint32, body string) error {
	bodyBytes := []byte(body)
	if packetlength == 0 {
		packetlength = uint32(len(bodyBytes) + 16)
	}
	headerBytes := new(bytes.Buffer)
	var data = []interface{}{
		packetlength,
		magic,
		ver,
		action,
		param,
	}
	for _, v := range data {
		err := binary.Write(headerBytes, binary.BigEndian, v)
		if err != nil {
			return err
		}
	}
	socketData := append(headerBytes.Bytes(), bodyBytes...)
	err := bili.serverConn.WriteMessage(websocket.TextMessage, socketData)
	return err
}

// joining the channel and listens to the bilibili server broadcast
func (bili *BiliClient) sendJoinChannel(channelID int) error {
	bili.uid = rand.Intn(2000000000) + 1000000000
	body := fmt.Sprintf("{\"roomid\":%d,\"uid\":%d}", channelID, bili.uid)
	return bili.sendSocketData(0, 16, bili.protocolVersion, 7, 1, body)
}

// update the connected state
func (bili *BiliClient) SetConnect(connected bool) {
	bili.mutex.Lock()
	bili.connected = connected
	bili.mutex.Unlock()
}

// check whether connected
func (bili *BiliClient) CheckConnect() bool {
	bili.mutex.Lock()
	defer bili.mutex.Unlock()
	return bili.connected
}

// Connect the bili client to the bilibili live-stream server
func (bili *BiliClient) Connect(rid int) error {
	roomID, err := getRealRoomID(rid)
	if err != nil {
		log.Println("client Connect " + err.Error())
		return errors.New("无法获取房间真实ID: " + err.Error())
	}

	u := url.URL{Scheme: "ws", Host: serverAddr, Path: "/sub"}

	if bili.serverConn, _, err = websocket.DefaultDialer.Dial(u.String(), nil); err != nil {
		log.Println("client Connect " + err.Error())
		return errors.New("Cannot dail websocket: " + err.Error())
	}
	bili.roomID = roomID

	if err := bili.sendJoinChannel(roomID); err != nil {
		return errors.New("Cannot send join channel: " + err.Error())
	}
	bili.SetConnect(true)

	go bili.heartbeatLoop()
	go bili.receiveMessages()
	return nil
}

func (bili *BiliClient) Disconnect() error {
	if bili.CheckConnect() {
		bili.SetConnect(false)
		if err := bili.serverConn.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (bili *BiliClient) receiveMessages() {
	for bili.CheckConnect() {
		_, message, err := bili.serverConn.ReadMessage()
		if err != nil {
			log.Fatalln("client receiveMessages " + err.Error())
		}
		packet, err := decode(message)
		if err != nil {
			log.Fatalln("client receiveMessages " + err.Error())
		}
		for _, body := range packet.body {
			bili.Ch <- body
		}
	}
}
