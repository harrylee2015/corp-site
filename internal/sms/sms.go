package sms

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
)

type Provider interface {
	Send(phone, code string) error
}

type MockProvider struct{}

func (m *MockProvider) Send(phone, code string) error {
	log.Printf("[SMS Mock] phone=%s", phone)
	return nil
}

type TencentProvider struct {
	secretID   string
	secretKey  string
	sdkAppID   string
	signName   string
	templateID string
}

func NewTencentProvider(secretID, secretKey, sdkAppID, signName, templateID string) *TencentProvider {
	return &TencentProvider{
		secretID:   secretID,
		secretKey:  secretKey,
		sdkAppID:   sdkAppID,
		signName:   signName,
		templateID: templateID,
	}
}

func (t *TencentProvider) Send(phone, code string) error {
	log.Printf("[SMS Tencent] phone=%s", phone)
	return nil
}

func GenerateCode(length int) string {
	if length <= 0 {
		length = 6
	}
	code := make([]byte, length)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		code[i] = byte('0' + n.Int64())
	}
	return fmt.Sprintf("%0*s", length, string(code))
}
