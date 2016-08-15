package client

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log"
	"pay/common"
	"sort"
	"strings"
)

var defaultAliAppClient *AliAppClient

type AliAppClient struct {
	partnerID   string //合作者ID
	sellerID    string
	AppID       string // 应用ID
	callbackURL string
	privateKey  *rsa.PrivateKey
	publicKey   *rsa.PublicKey
	queryURL    string // 查询订单接口地址
}

// DefaultAliAppClient 得到默认支付宝app客户端
func DefaultAliAppClient() *AliAppClient {
	return defaultAliAppClient
}

func (aa *AliAppClient) Pay(charge *common.Charge) (string, error) {
	data := make(map[string]string)
	data["service"] = "mobile.securitypay.pay"
	data["partner"] = aa.partnerID
	data["_input_charset"] = "utf-8"
	data["notify_url"] = aa.callbackURL

	data["out_trade_no"] = charge.TradeNum
	data["subject"] = charge.Describe
	data["payment_type"] = "1"
	data["seller_id"] = aa.sellerID
	data["total_fee"] = fmt.Sprintf("%.2f", float64(charge.MoneyFee)/float64(100))
	data["body"] = charge.Describe

	sign, err := aa.GenSign(data)
	if err != nil {
		return "", err
	}
	data["sign"] = sign
	data["sign_type"] = "RSA"

	var re []string
	for k, v := range data {
		re = append(re, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(re, "&"), nil
}

// GenSign 产生签名
func (aa *AliAppClient) GenSign(m map[string]string) (string, error) {
	delete(m, "sign_type")
	delete(m, "sign")
	var data []string
	for k, v := range m {
		if v == "" {
			continue
		}
		data = append(data, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(data)
	signData := strings.Join(data, "&")
	s := sha1.New()
	_, err := s.Write([]byte(signData))
	if err != nil {
		log.Println(err)
	}
	hashByte := s.Sum(nil)
	signByte, err := aa.privateKey.Sign(rand.Reader, hashByte, crypto.SHA1)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(signByte), nil
}

// CheckSign 检测签名
func (aa *AliAppClient) CheckSign(data string, sign string) error {
	signByte, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return err
	}
	s := sha1.New()
	_, err = s.Write([]byte(data))
	if err != nil {
		return err
	}
	hash := s.Sum(nil)
	return rsa.VerifyPKCS1v15(aa.publicKey, crypto.SHA1, hash[:], signByte)
}