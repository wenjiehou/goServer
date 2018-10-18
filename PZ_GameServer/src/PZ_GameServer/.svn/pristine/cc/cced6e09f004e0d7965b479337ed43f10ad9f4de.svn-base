package util

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"
	"math/big"
	rnd "math/rand"
	"time"

	logf "PZ_GameServer/log"
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

//得到Session Id
func GetSID() string {
	uuid, err := genUUID()
	if err != nil {
		logf.Fatal("GenUUID error %s", err)
	}
	return uuid
}

var UserCount = 0

//得到UserID
func GetUID() int {
	UserCount++
	return UserCount //strconv.Itoa(UserCount)
}

//得到UUID
func genUUID() (string, error) {
	uuid := make([]byte, 8)
	n, err := rand.Read(uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// TODO: verify the two lines implement RFC 4122 correctly
	uuid[4] = 0x80 // variant bits see page 5
	uuid[2] = 0x40 // version 4 Pseudo Random, see page 7

	return hex.EncodeToString(uuid), nil
}

// 函　数：生成随机数
// 概　要：
// 参　数：
//      min: 最小值
//      max: 最大值
// 返回值：
//      int : 生成的随机数
func RandInt(min, max int) int {
	if min >= max || max == 0 {
		return max
	}
	r := rnd.New(rnd.NewSource(time.Now().UnixNano()))
	return r.Intn(max-min) + min
}

//随机数
func RandFloat(min, max int64) float32 {
	s := RandInt64(min, max-1)
	r := rnd.New(rnd.NewSource(time.Now().UnixNano()))
	return float32(r.ExpFloat64()) + float32(s)
}

func RandInt64(min, max int64) int64 {
	maxBigInt := big.NewInt(max)
	i, _ := rand.Int(rand.Reader, maxBigInt)
	if i.Int64() < min {
		RandInt64(min, max)
	}
	return i.Int64()
}

//生成32位md5字串
func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

//生成Guid字串
func getGuid() string {
	b := make([]byte, 16)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return GetMd5String(base64.URLEncoding.EncodeToString(b))
}

//
func Encode(data interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

//
func Decode(data []byte, to interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(to)
}

func GetIndex(arr []int, value int) int {
	for i, v := range arr {
		if v == value {
			return i
		}
	}
	return -1
}

func IntArrToString(slices []int32) string {
	var arrStr []string
	if len(slices) == 0 {
		return ""
	}
	for _, v := range slices {
		arrStr = append(arrStr, strconv.Itoa(int(v)))
	}

	if len(arrStr) > 0 {
		return strings.Join(arrStr, ",")
	}

	return ""
}

//调用内部函数
func FunCall(m interface{}, n string, arg []reflect.Value) {
	t := reflect.ValueOf(m)
	f := t.MethodByName(n)
	if f.IsValid() {
		f.Call(arg)
	} else {
		fmt.Println("错误的反射方法 ", n)
	}
}

const (
	zero  = byte('0')
	one   = byte('1')
	lsb   = byte('[') // left square brackets
	rsb   = byte(']') // right square brackets
	space = byte(' ')
)

var uint8arr [8]uint8

// ErrBadStringFormat represents a error of input string's format is illegal .
var ErrBadStringFormat = errors.New("bad string format")

// ErrEmptyString represents a error of empty input string.
var ErrEmptyString = errors.New("empty string")

func init() {
	uint8arr[0] = 128
	uint8arr[1] = 64
	uint8arr[2] = 32
	uint8arr[3] = 16
	uint8arr[4] = 8
	uint8arr[5] = 4
	uint8arr[6] = 2
	uint8arr[7] = 1
}

// append bytes of string in binary format.
func appendBinaryString(bs []byte, b byte) []byte {
	var a byte
	for i := 0; i < 8; i++ {
		a = b
		b <<= 1
		b >>= 1
		switch a {
		case b:
			bs = append(bs, zero)
		default:
			bs = append(bs, one)
		}
		b <<= 1
	}
	return bs
}

// ByteToBinaryString get the string in binary format of a byte or uint8.
func ByteToBinaryString(b byte) string {
	buf := make([]byte, 0, 8)
	buf = appendBinaryString(buf, b)
	return string(buf)
}

// BytesToBinaryString get the string in binary format of a []byte or []int8.
func BytesToBinaryString(bs []byte) string {
	l := len(bs)
	bl := l*8 + l + 1
	buf := make([]byte, 0, bl)
	buf = append(buf, lsb)
	for _, b := range bs {
		buf = appendBinaryString(buf, b)
		buf = append(buf, space)
	}
	buf[bl-1] = rsb
	return string(buf)
}

// regex for delete useless string which is going to be in binary format.
var rbDel = regexp.MustCompile(`[^01]`)

// BinaryStringToBytes get the binary bytes according to the
// input string which is in binary format.
func BinaryStringToBytes(s string) (bs []byte) {
	if len(s) == 0 {
		panic(ErrEmptyString)
	}

	s = rbDel.ReplaceAllString(s, "")
	l := len(s)
	if l == 0 {
		panic(ErrBadStringFormat)
	}

	mo := l % 8
	l /= 8
	if mo != 0 {
		l++
	}
	bs = make([]byte, 0, l)
	mo = 8 - mo
	var n uint8
	for i, b := range []byte(s) {
		m := (i + mo) % 8
		switch b {
		case one:
			n += uint8arr[m]
		}
		if m == 7 {
			bs = append(bs, n)
			n = 0
		}
	}
	return
}
