package util

import (
	"github.com/bwmarrin/snowflake"
)

const base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var node *snowflake.Node

func init() {
	node, _ = snowflake.NewNode(1)
}

func Base62Encode(num int64) string {
	if num == 0 {
		return "0"
	}
	var result []byte
	base := int64(62)
	for num > 0 {
		remainder := num % base
		result = append([]byte{base62Chars[remainder]}, result...)
		num = num / base
	}
	return string(result)
}

func GenerateShortCode() string {
	id := node.Generate().Int64()
	obfuscated := id ^ 0xAAAAAAAAAAAAAAA
	return Base62Encode(obfuscated)
}