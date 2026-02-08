package auth

import (
	"bytes"

	"github.com/yeqown/go-qrcode"
)

func GenerateQRCodePNG(content string) ([]byte, error) {
	qrc, err := qrcode.New(content)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := qrc.SaveTo(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
