package cipher

import (
	"crypto/md5"
	"fmt"
	"net"
	"strings"

	"github.com/Dreamacro/go-shadowsocks2/core"
	"github.com/Dreamacro/go-shadowsocks2/shadowstream"
)

type Cipher interface {
	StreamConn(net.Conn) (net.Conn, error)
	PacketConn(net.PacketConn) (net.PacketConn, error)
}

type shadowSocksRStreamCipher struct {
	shadowstream.Cipher
	Key []byte
}

func (ciph *shadowSocksRStreamCipher) StreamConn(c net.Conn) (net.Conn, error) {
	return NewShadowSocksRStreamConn(c, *ciph)
}

func (ciph *shadowSocksRStreamCipher) PacketConn(c net.PacketConn) (net.PacketConn, error) {
	return nil, fmt.Errorf("packet conn for ssr not implemented")
}

type none struct{}

func (none) StreamConn(c net.Conn) (net.Conn, error) {
	return c, nil
}
func (none) PacketConn(c net.PacketConn) (net.PacketConn, error) {
	return c, nil
}

var streamList = map[string]struct {
	KeySize int
	New     func(key []byte) (shadowstream.Cipher, error)
}{
	"RC4-MD5-6":     {16, RC4MD56},
	"RC4-MD5":       {16, shadowstream.RC4MD5},
	"AES-128-CFB":   {16, shadowstream.AESCFB},
	"AES-192-CFB":   {24, shadowstream.AESCFB},
	"AES-256-CFB":   {32, shadowstream.AESCFB},
	"AES-128-CTR":   {16, shadowstream.AESCTR},
	"AES-192-CTR":   {24, shadowstream.AESCTR},
	"AES-256-CTR":   {32, shadowstream.AESCTR},
	"BF-CFB":        {16, BFCFB},
	"CAST5-CFB":     {16, Cast5CFB},
	"DES-CFB":       {8, DESCFB},
	"SALSA20":       {32, Salsa20},
	"CHACHA20":      {32, Chacha20},
	"CHACHA20-IETF": {32, shadowstream.Chacha20IETF},
}

func PickCipher(name string, key []byte, password string) (Cipher, error) {
	name = strings.ToUpper(name)

	if name == "NONE" {
		return &none{}, nil
	}

	if choice, ok := streamList[name]; ok {
		if len(key) == 0 {
			key = kdf(password, choice.KeySize)
		}
		if len(key) != choice.KeySize {
			return nil, shadowstream.KeySizeError(choice.KeySize)
		}
		ciph, err := choice.New(key)
		return &shadowSocksRStreamCipher{ciph, key}, err
	}

	return nil, core.ErrCipherNotSupported
}

// Copied from the go-shadowsocks2 project because this function is not made public
// https://github.com/shadowsocks/go-shadowsocks2
// Licensed under Apache License 2.0
// key-derivation function from original Shadowsocks
func kdf(password string, keyLen int) []byte {
	var b, prev []byte
	h := md5.New()
	for len(b) < keyLen {
		h.Write(prev)
		h.Write([]byte(password))
		b = h.Sum(b)
		prev = b[len(b)-h.Size():]
		h.Reset()
	}
	return b[:keyLen]
}
