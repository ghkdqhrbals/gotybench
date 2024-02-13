package benchmark

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
)

var randomKr = []rune("김이박최정강조윤장임한오서신권황안송류전홍고문양손배조백허유남심노정하곽성차주우구신임나전민유진지엄채원천방공강현함변염양변여추노도소신석선설마길주연방위표명기반왕금옥육인맹제모장남탁국여진어은편구용")
const (
	randomEn            = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	YYYYMMDD            = "2006-01-02"
	rangeOfRandomString = 10
	rangeOfRandomInt    = 10000
)

// 랜덤한 영어 문자열을 생성합니다.
func RandStringEn(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = randomEn[rand.Intn(len(randomEn))]
	}
	return string(b)
}

// 랜덤한 한글 문자열을 생성합니다.
func RandStringKr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = randomKr[rand.Intn(len(randomKr))]
	}
	return string(b)
}

// 랜덤 데이터를 생성합니다.
func RandomGeneratedData(jsons map[string]string) []byte {
	data := ""
	for key, val := range jsons {
		if returnRandomByTypes(val) == "null" {
			errorKeyType(key)
			os.Exit(1)
		}
		data += "\"" + key + "\"" + ":" + returnRandomByTypes(val) + ","
	}

	return []byte(`{` + data[:len(data)-1] + `}`)
}

// 타입에 따른 랜덤 데이터를 반환합니다.
func returnRandomByTypes(column_type string) string {
	switch column_type {
	case "int":
		return strconv.Itoa(rand.Intn(rangeOfRandomInt))
	case "double":
		return fmt.Sprintf("%f", rand.Float64())
	case "string":
		return "\"" + RandStringEn(rangeOfRandomString) + "\""
	case "boolean":
		return "false"
	}
	return "null"
}