package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
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

func randomGeneratedData(jsons map[string]string) []byte {

	data := ""
	for key, val := range jsons {
		if returnRandomByTypes(val) == "null" {
			errorKeyType(key)
			os.Exit(1)
		}
		data += "\"" + key + "\"" + ":" + returnRandomByTypes(val) + ","
	}
	fmt.Println(data)

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

func goid() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

type User struct {
	UserId   string `json:"userId"`
	UserName string `json:"userName"`
	Email    string `json:"email"`
	UserPw   string `json:"userPw"`
}

type RequestAddChatMessageDTO struct {
	RoomId   int    `json:"roomId"`
	Writer   string `json:"writer"`
	WriterId string `json:"writerId"`
	Message  string `json:"message"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
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

func worker(jsons map[string]string, mutex *sync.RWMutex, contexts *sync.Map, wg *sync.WaitGroup, requestURL string, client *http.Client, transferRatePerSecond int, number_worker int) {

	defer wg.Done()
	// 로컬 맵 생성
	m := make(map[int]int)
	instance := dynamicstruct.NewStruct().Build().New()
	for key, val := range jsons {
		temp_instance := dynamicstruct.NewStruct().AddField(strings.ToUpper(key), returnDefaults(val), `json:"`+key+`"`).Build().New()
		instance = dynamicstruct.MergeStructs(instance, temp_instance).Build().New()
	}

	for i := 1; i < (transferRatePerSecond)/2+1; i++ {
		data := randomGeneratedData(jsons)
		err := json.Unmarshal(data, &instance)
		if err != nil {
			log.Fatal(err)
		}

		// data, err = json.Marshal(instance)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// fmt.Println(string(data))

		// ---------- change this data ---------
		s := &User{
			UserId:   "A",
			UserName: RandStringKr(1) + RandStringKr(2),
			Email:    RandStringEn(5) + "@gmail.com",
			UserPw:   RandStringEn(10),
		}
		fmt.Println("instance:", s)

		// s := &RequestAddChatMessageDTO{
		// 	RoomId:   7,
		// 	Writer:   "bbb",
		// 	WriterId: "b",
		// 	Message:  RandStringEn(10),
		// }

		buf, err := json.Marshal(s)
		if err != nil {
			fmt.Println(err)
			return
		}
		// ------------------------------------

		bodyReader := bytes.NewReader(buf)

		// 요청 생성
		req, err := http.NewRequest(http.MethodPost, requestURL, bodyReader)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// json 헤더 설정
		req.Header.Set("Content-Type", "application/json")

		// 실제 요청 전송 및 반환
		res, err := client.Do(req)
		if err != nil {
			fmt.Println("ERROR MESSAGE:", err)
		}

		// 로컬 맵에 삽입
		m[res.StatusCode] += 1
	}

	mutex.Lock()
	for k, v := range m {
		result, ok := contexts.Load(k)
		if ok {
			contexts.Store(k, result.(int)+v)
		} else {
			contexts.Store(k, v)
		}
	}
	mutex.Unlock()
}

func main() {

	var wg sync.WaitGroup
	var mutex = &sync.RWMutex{}
	rand.Seed(time.Now().UnixNano())
	flag.Parse()
	configuration := NewConfiguration()

	// config, err := util.LoadConfig(".")
	// if err != nil {
	// 	log.Fatal("config 가져오기 에러", err)
	// }

	fmt.Println("Request url:", configuration.url)
	fmt.Println("The number of HTTP Requests:", configuration.requests)
	fmt.Println("The number of threads:", configuration.clients)

	// http 전송 url
	requestURL := configuration.url

	// 실행횟수 설정
	transferRatePerSecond := configuration.requests

	// 실행 스레드 개수 설정
	number_worker := configuration.clients

	// NGINX can handle a maximum of 512 concurrent connections. In newer versions, NGINX supports up to 1024 concurrent connections, by default.
	// 이를 잘 확인하고 설정해야합니다.
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 1000    // connection pool 크기
	t.MaxConnsPerHost = 1000 // 호스트 별 최대 할당 connection
	t.MaxIdleConnsPerHost = 100

	// 클라이언트 설정 및 timeout
	client := &http.Client{
		Timeout:   60 * time.Second,
		Transport: t,
	}

	// 스레드 싱크 맵
	var contexts = &sync.Map{}

	startTime := time.Now()
	// 멀티 스레드 http request
	for i := 0; i < int(number_worker); i++ {
		wg.Add(1)
		go worker(configuration.randomJson, mutex, contexts, &wg, requestURL, client, int(transferRatePerSecond/number_worker), i)
	}
	fmt.Println("Proceeding! Please wait until getting all the responses")
	wg.Wait()

	elapsedTime := time.Since(startTime)
	fmt.Println("Elapsed Time:", elapsedTime.Seconds())

	// 성공한 http request 개수 확인
	contexts.Range(func(k, v interface{}) bool {
		fmt.Println("Response status code: ", k, ", How many?: ", v)
		return true
	})
}
