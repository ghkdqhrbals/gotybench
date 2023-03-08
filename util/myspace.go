package util

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	dynamicstruct "github.com/ompluscator/dynamic-struct"
)

var (
	requests int64
	url      string
	clients  int64
	body     string
	method   string
)

func init() {
	flag.Int64Var(&requests, "r", 10000, "요청 개수")
	flag.Int64Var(&clients, "t", 100, "스레드 개수")
	flag.StringVar(&url, "u", "", "URL")
	flag.StringVar(&body, "j", "", "Json \"[KEY1,TYPE1,KEY2,TYPE2,...]\" ")
}

type JSONTime struct {
	time.Time
}

func (t JSONTime) MarshalJSON() ([]byte, error) {
	//do your serializing here
	stamp := fmt.Sprintf("\"%s\"", t.Format("Mon Jan _2"))
	return []byte(stamp), nil
}

type Configuration struct {
	url        string
	method     string
	requests   int64
	clients    int64
	randomJson map[string]string
}

func NewConfiguration() *Configuration {

	if url == "" {
		flag.Usage()
		os.Exit(1)
	}

	if body == "" {
		flag.Usage()
		os.Exit(1)
	}

	configuration := &Configuration{
		url:        "",
		method:     http.MethodPost,
		requests:   int64((1 << 63) - 1),
		clients:    int64((1 << 63) - 1),
		randomJson: map[string]string{},
	}

	if url != "" {
		configuration.url = url
	}

	if requests != 10000 {
		configuration.requests = requests
	}

	if clients != 100 {
		configuration.clients = clients
	}

	if body != "" {
		body = strings.Replace(body, "[", "", -1)
		body = strings.Replace(body, "]", "", -1)
		temps := strings.Split(body, ",")
		for i := 0; i < len(temps); i += 2 {
			if i > len(temps)-2 {
				errorKeyType(temps[i])
				os.Exit(1)
			}
			configuration.randomJson[temps[i]] = temps[i+1]
		}
	}

	return configuration
}

func returnDefaults(column_type string) interface{} {
	switch column_type {
	case "int":
		return 0
	case "double":
		return 0.0
	case "string":
		return ""
	case "time":
		return JSONTime{time.Now()}
	case "boolean":
		return false
	}
	return "null"
}

func returnRandomByTypes(column_type string) string {
	switch column_type {
	case "int":
		return strconv.Itoa(rand.Intn(100))
	case "double":
		return fmt.Sprintf("%f", rand.Float64())
	case "string":
		return "\"" + RandStringEn(6) + "\""
	case "boolean":
		return "false"
	}
	return "null"
}

func dynamicstructs(jsons map[string]string) interface{} {
	instance := dynamicstruct.NewStruct()
	for key, val := range jsons {
		instance.AddField(strings.ToUpper(key), returnDefaults(val), `json:"`+key+`"`)
		fmt.Println(key, val)
	}
	return instance.Build().New()
}

var randomKr = []rune("김이박최정강조윤장임한오서신권황안송류전홍고문양손배조백허유남심노정하곽성차주우구신임나전민유진지엄채원천방공강현함변염양변여추노도소신석선설마길주연방위표명기반왕금옥육인맹제모장남탁국여진어은편구용")

const randomEn = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringEn(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = randomEn[rand.Intn(len(randomEn))]
	}
	return string(b)
}

func RandStringKr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = randomKr[rand.Intn(len(randomKr))]
	}
	return string(b)
}

func randomGeneratedData(jsons map[string]string) []byte {
	RandStringKr(3)
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

func errorKeyType(key string) {
	fmt.Println("[KEY ERROR] ", key, "의 타입이 없습니다. 아래의 타입 중 하나를 선택해서 표기해주세요.")
	fmt.Println("| - Support Json Type -\t|")
	fmt.Println("| * int \t\t|")
	fmt.Println("| * float \t\t|")
	fmt.Println("| * string \t\t|")
	fmt.Println("| * boolean \t\t|")
	os.Exit(1)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	flag.Parse()
	configuration := NewConfiguration()

	instance := dynamicstruct.NewStruct().Build().New()
	for key, val := range configuration.randomJson {
		temp_instance := dynamicstruct.NewStruct().AddField(strings.ToUpper(key), returnDefaults(val), `json:"`+key+`"`).Build().New()
		instance = dynamicstruct.MergeStructs(instance, temp_instance).Build().New()
	}

	data := randomGeneratedData(configuration.randomJson)

	err := json.Unmarshal(data, &instance)
	if err != nil {
		log.Fatal(err)
	}

	data, err = json.Marshal(instance)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))

	// END

	data = randomGeneratedData(configuration.randomJson)

	err = json.Unmarshal(data, &instance)
	if err != nil {
		log.Fatal(err)
	}

	data, err = json.Marshal(instance)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(data))
	// Out:
	// {"int":123,"someText":"example","double":123.45,"Boolean":true,"Slice":[1,2,3]}
}
