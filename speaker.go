package biliStreamClient

import (
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tts "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tts/v20190823"
	"math"
	"unicode/utf8"
)

type VoiceConfig struct {
	Endpoint  string
	Region    string
	VoiceCode int64
}

var (
	DefaultBoyVoice = VoiceConfig{
		Endpoint:  "tts.tencentcloudapi.com",
		Region:    "ap-shanghai",
		VoiceCode: 101015,
	}
	DefaultGirlVoice = VoiceConfig{
		Endpoint:  "tts.tencentcloudapi.com",
		Region:    "ap-shanghai",
		VoiceCode: 101016,
	}
)

// 参考腾讯云官方文档： https://cloud.tencent.com/document/product/1073/37995

func GetVoiceFromTencentCloud(SecretID string, SecretKey string, voice VoiceConfig, message string) (string, error) {
	credential := common.NewCredential(
		SecretID,
		SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = voice.Endpoint
	client, _ := tts.NewClient(credential, voice.Region, cpf)

	request := tts.NewTextToVoiceRequest()
	request.Text = common.StringPtr(message)
	request.ModelType = common.Int64Ptr(1)
	request.VoiceType = common.Int64Ptr(voice.VoiceCode)
	request.Volume = common.Float64Ptr(10)
	request.Speed = common.Float64Ptr(math.Min(2.0, float64(utf8.RuneCountInString(message)/8)))
	response, err := client.TextToVoice(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return "", err
	}

	//base64编码的wav/mp3音频数据
	return *response.Response.Audio, nil
}
