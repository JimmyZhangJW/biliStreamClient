package biliStreamClient

import (
	"errors"
	"fmt"
)

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

type Sender struct {
	Id      string
	Name    string
	FaceUrl string
}

type Gift struct {
	Sender
	Action   string
	GiftName string
	Price    int
	ComboId  string
}

type GiftCombo struct {
	Gift
	TotalNumber int
	TotalPrice  int
}

type DanmuMsg struct {
	Sender
	Message string
}

func (body PacketBody) ParseGift() (Gift, error) {
	if body.Cmd != "SEND_GIFT" {
		return Gift{}, errors.New("packet is not about a gift")
	}
	gift := Gift{
		Action:   body.Data["action"].(string),
		GiftName: body.Data["giftName"].(string),
		Sender: Sender{
			Id:      fmt.Sprint(int(body.Data["uid"].(float64))),
			Name:    body.Data["uname"].(string),
			FaceUrl: body.Data["face"].(string),
		},
		Price: int(body.Data["price"].(float64)),
	}
	// If contain the batch_combo_id tag, add to the gift
	comboId, ok := body.Data["batch_combo_id"].(string)
	if ok {
		gift.ComboId = comboId
	}
	return gift, nil
}

func (body PacketBody) ParseGiftCombo() (GiftCombo, error) {
	if body.Cmd != "COMBO_SEND" {
		return GiftCombo{}, errors.New("packet is not about a gift combo")
	}
	comboNumber := int(body.Data["total_num"].(float64))
	totalPrice := int(body.Data["combo_total_coin"].(float64))
	combo := GiftCombo{
		Gift: Gift{
			Action:   body.Data["action"].(string),
			GiftName: body.Data["gift_name"].(string),
			Sender: Sender{
				Id:      fmt.Sprint(int(body.Data["uid"].(float64))),
				Name:    body.Data["uname"].(string),
				FaceUrl: "",
			},
			Price:   totalPrice / comboNumber,
			ComboId: body.Data["batch_combo_id"].(string),
		},
		TotalNumber: comboNumber,
		TotalPrice:  totalPrice,
	}
	return combo, nil
}

func (body PacketBody) ParseDanmu() (DanmuMsg, error) {
	if body.Cmd != "DANMU_MSG" {
		return DanmuMsg{}, errors.New("packet is not about a danmu message")
	}
	return DanmuMsg{
		Message: body.Info[1].(string),
		Sender: Sender{
			fmt.Sprint(int(body.Info[2].([]interface{})[0].(float64))),
			body.Info[2].([]interface{})[1].(string),
			"",
		},
	}, nil
}
