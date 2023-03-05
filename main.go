package main

import (
	"bytes"
	"encoding/json"
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

	"testapi.com/m/util"
)

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

func worker(mutex *sync.RWMutex, contexts *sync.Map, wg *sync.WaitGroup, requestURL string, client *http.Client, transferRatePerSecond int, number_worker int) {

	defer wg.Done()
	// 로컬 맵 생성
	m := make(map[int]int)

	for i := 1; i < transferRatePerSecond+1; i++ {

		// ---------- change this data ---------
		s := &User{
			UserId:   RandStringEn(8),
			UserName: RandStringKr(1) + RandStringKr(2),
			Email:    RandStringEn(5) + "@gmail.com",
			UserPw:   RandStringEn(10),
		}

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

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("config 가져오기 에러", err)
	}
	fmt.Println("Request url:", config.RequestUrl)
	fmt.Println("The number of HTTP Requests:", config.RequestNum)
	fmt.Println("The number of threads:", config.WorkerNum)

	// http 전송 url
	requestURL := config.RequestUrl

	// 실행횟수 설정
	transferRatePerSecond := config.RequestNum

	// 실행 스레드 개수 설정
	number_worker := config.WorkerNum

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
	for i := 0; i < number_worker; i++ {
		wg.Add(1)
		go worker(mutex, contexts, &wg, requestURL, client, int(transferRatePerSecond/number_worker), i)
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
