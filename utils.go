package biliStreamClient

import (
	"strings"
	"unicode"
)

func removeRepeated(message string) string {
	notRepeatedWords := "？，.，！ 。"
	msgRunes := []rune(message)
	if len(msgRunes) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteRune(msgRunes[0])
	for i := 1; i < len(msgRunes); i++ {
		if !(strings.ContainsRune(notRepeatedWords, msgRunes[i]) && strings.ContainsRune(notRepeatedWords, msgRunes[i-1])) {
			sb.WriteRune(msgRunes[i])
		}
	}
	return sb.String()
}

// Remove Emojis and repeated punctuations
func Sanitize(message string) string {
	notAllowedWords := "゛(⌒▽)（￣）>==・ω｀´〜△･∀°ﾉ╮╭_: 」∠ゝ←→<>;¬\"\"▔□/ﾟД≡д!?Σ|；`T^｡●ｴεノ≧∇≦-=#へヽ┯━╯口┴—◡♥―〃♡⁄•╬▄︻┻┳═*$&~+'{}[]<>丶。.\\"
	var sb strings.Builder
	for _, r := range message {
		if !strings.ContainsRune(notAllowedWords, r) {
			sb.WriteRune(r)
		}
	}
	temp := sb.String()
	temp = removeRepeated(temp)
	temp = strings.Trim(temp, "()《》，")
	temp = strings.TrimLeft(temp, "？")
	return temp
}

// Determine if the message contain Chinese word or not
func IsContainChineseWord(message string) bool {
	for _, r := range message {
		if unicode.Is(unicode.Scripts["Han"], r) {
			return true
		}
	}
	return false
}
