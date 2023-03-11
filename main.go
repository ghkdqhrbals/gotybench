package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/gosuri/uilive"
	dynamicstruct "github.com/ompluscator/dynamic-struct"
)

var (
	requests       int64
	url            string
	clients        int64
	body           string
	method         string
	client_timeout int64
)

func init() {
	flag.Int64Var(&requests, "r", 10000, "요청 개수")
	flag.Int64Var(&clients, "c", 100, "스레드 개수")
	flag.Int64Var(&client_timeout, "t", 30, "요청 타임아웃(second)")
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
	url            string
	method         string
	requests       int64
	clients        int64
	randomJson     map[string]string
	client_timeout int64
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
		url:            "",
		method:         http.MethodPost,
		requests:       int64(10000),
		clients:        int64(100),
		client_timeout: int64(30),
		randomJson:     map[string]string{},
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

	if client_timeout != 30 {
		configuration.client_timeout = client_timeout
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
	// fmt.Println(data)

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

type nonBlocking struct {
	Response    *http.Response
	Error       error
	ElapsedTime float64
}

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func worker(nb chan nonBlocking, jsons map[string]string, mutex *sync.RWMutex, contexts *sync.Map, wg *sync.WaitGroup, requestURL string, client *http.Client, transferRatePerSecond int, number_worker int, results *sync.Map) {
	defer wg.Done()
	// 로컬 맵 생성
	m := make(map[int]int)
	instance := dynamicstruct.NewStruct().Build().New()

	for key, val := range jsons {
		temp_instance := dynamicstruct.NewStruct().AddField(strings.ToUpper(key), returnDefaults(val), `json:"`+key+`"`).Build().New()
		instance = dynamicstruct.MergeStructs(instance, temp_instance).Build().New()
	}

	for i := 1; i < (transferRatePerSecond)+1; i++ {

		data := randomGeneratedData(jsons)
		err := json.Unmarshal(data, &instance)
		if err != nil {
			log.Fatal(err)
		}

		data, err = json.Marshal(instance)
		if err != nil {
			log.Fatal(err)
		}

		bodyReader := bytes.NewReader(data)

		// 요청 생성
		req, err := http.NewRequest(http.MethodPost, requestURL, bodyReader)
		if err != nil {
			fmt.Println(err)
			req.Body.Close()
		}

		// json 헤더 설정
		req.Header.Set("Content-Type", "application/json")

		startTime := time.Now()
		// 실제 요청 전송 및 반환
		res, err := client.Do(req)
		elapsedTime := time.Since(startTime).Seconds()
		// fmt.Println("TREAHD: ", number_worker)
		nb <- nonBlocking{
			Response:    res,
			Error:       err,
			ElapsedTime: elapsedTime,
		}

		if err != nil {
			fmt.Println("ERROR MESSAGE:", err)
			// fmt.Println("에러 스테이터스 코드 : ", res.StatusCode)
			continue
		} else {

			m[res.StatusCode] += 1
			res.Body.Close()
		}

	}

}

func MaxParallelism() int {
	maxProcs := runtime.GOMAXPROCS(0)
	numCPU := runtime.NumCPU()
	if maxProcs < numCPU {
		return maxProcs
	}
	return numCPU
}

func HandleResponse(nb chan nonBlocking, wg *sync.WaitGroup, requests int) {

	// avg := 0.0
	max_response_time := 0.0
	min_response_time := -1.0
	num := 0
	writer := uilive.New() // writer for the first line
	writer.Start()

	m := make(map[int]int)
	elaspedTime := 0.0

	for get := range nb {
		num += 1
		if get.Error != nil {
			// log.Println(get.Error)
		} else {
			// fmt.Println(time.Now(), get.Response.StatusCode)
			// avg = float64(num-1)/float64(num)*avg + 1/float64(num)*float64(get.ElapsedTime*1000) // 평균 필터
			if min_response_time < 0 {
				min_response_time = get.ElapsedTime
			}
			max_response_time = math.Max(max_response_time, get.ElapsedTime)
			min_response_time = math.Min(min_response_time, get.ElapsedTime)

			// log.Println(get.Response.Status)
			m[get.Response.StatusCode] += 1
			elaspedTime += get.ElapsedTime
		}

		if num == requests {
			for k, v := range m {
				fmt.Println("\t[RESULTS] Response status code: ", k, ", How many?: ", v)
				fmt.Println(max_response_time, min_response_time)
			}
		}
		wg.Done()
	}

}

func main() {
	PrintMemUsage()

	// success := color.New(color.Bold, color.FgGreen).FprintlnFunc()
	// Or just add them to New()
	d := color.New(color.FgCyan, color.Bold)

	// Mix up foreground and background colors, create new mixes!
	green := color.New(color.FgGreen)
	boldGreen := green.Add(color.Bold)

	d.Printf("\t Properties\n")
	color.Cyan("\t- Max parallelism : %d", MaxParallelism())

	wg := &sync.WaitGroup{}
	var mutex = &sync.RWMutex{}
	rand.Seed(time.Now().UnixNano())
	flag.Parse()
	configuration := NewConfiguration()

	// config, err := util.LoadConfig(".")
	// if err != nil {
	// 	log.Fatal("config 가져오기 에러", err)
	// }

	color.Cyan("\t- Request url : %s", configuration.url)
	color.Cyan("\t- The number of HTTP Requests : %d", configuration.requests)
	color.Cyan("\t- The number of threads : %d", configuration.clients)

	// http 전송 url
	requestURL := configuration.url

	// 실행횟수 설정
	transferRatePerSecond := configuration.requests

	// 실행 스레드 개수 설정
	number_worker := configuration.clients

	// 전체 실행 개수
	nb := make(chan nonBlocking, configuration.requests)

	// NGINX can handle a maximum of 512 concurrent connections. In newer versions, NGINX supports up to 1024 concurrent connections, by default.
	// 이를 잘 확인하고 설정해야합니다.
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100    // connection pool 크기
	t.MaxConnsPerHost = 100 // 호스트 별 최대 할당 connection
	t.MaxIdleConnsPerHost = 100
	t.IdleConnTimeout = 30 * time.Second

	// 클라이언트 설정 및 timeout
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: t,
	}

	// 스레드 싱크 맵
	var contexts = &sync.Map{}
	var results = &sync.Map{}

	startTime := time.Now()
	// 멀티 스레드 http request
	for i := 0; i < int(number_worker); i++ {
		wg.Add(1)
		go worker(nb, configuration.randomJson, mutex, contexts, wg, requestURL, client, int(transferRatePerSecond/number_worker), i, results)
	}
	fmt.Println("\t=> Proceeding... Please wait until getting all the responses", int(configuration.requests))

	wg.Add(int(configuration.requests))
	go HandleResponse(nb, wg, int(configuration.requests))
	wg.Wait()

	elapsedTime := time.Since(startTime)

	avg := 0.0
	max_response_time := 0.0
	min_response_time := 0.0

	results.Range(func(k, v interface{}) bool {
		sums := v.([]float64)
		avg += sums[0]
		max_response_time += sums[1]
		min_response_time += sums[2]
		return true
	})

	boldGreen.Printf("\t[RESULTS] Elapsed Time : %s\n", elapsedTime)
	// avg_str := fmt.Sprintf("%.2f", avg/float64(number_worker))
	// max_resp_str := fmt.Sprintf("%.2f", max_response_time*1000/float64(number_worker))
	// min_resp_str := fmt.Sprintf("%.2f", min_response_time*1000/float64(number_worker))

	// fmt.Println("\t[RESULTS] AVG : ", avg_str, "MAX_RESPONSE_TIME : ", max_resp_str, "MIN_RESPONSE_TIME : ", min_resp_str)

	// // 성공한 http request 개수 확인
	// contexts.Range(func(k, v interface{}) bool {
	// 	fmt.Println("\t[RESULTS] Response status code: ", k, ", How many?: ", v)
	// 	return true
	// })
	PrintMemUsage()
	runtime.GC()
}
