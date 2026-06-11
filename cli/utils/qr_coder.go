package qr_coder

import (
	"fmt"

	qrcode "github.com/skip2/go-qrcode"
)

func GenerateQRStr(content string) error {

	q, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		fmt.Printf("\033[38;5;208mError: could not generate QRCode: %v\033[0m\n", err) // color is orange
		return err
	}
	fmt.Println(q.ToSmallString(false))
	fmt.Println()
	return nil

}
