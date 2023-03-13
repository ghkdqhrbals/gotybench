package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/fatih/color"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
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
	hp             string // help flag
	openServer     *bool
	line2web       *charts.Line // Response 시간을 웹에 표시

	infos    = fmt.Sprintf("\n| - Support Json Type -\t|\n| * int \t\t|\n| * float \t\t|\n| * string \t\t|\n| * boolean \t\t|")
	randomKr = []rune("김이박최정강조윤장임한오서신권황안송류전홍고문양손배조백허유남심노정하곽성차주우구신임나전민유진지엄채원천방공강현함변염양변여추노도소신석선설마길주연방위표명기반왕금옥육인맹제모장남탁국여진어은편구용")

	green      = color.New(color.FgGreen)
	boldGreen  = color.New(color.FgGreen).Add(color.Bold)
	yellow     = color.New(color.FgHiYellow)
	boldYellow = color.New(color.FgHiYellow).Add(color.Bold)
	mg         = color.New(color.FgHiMagenta)
	boldMg     = color.New(color.FgHiMagenta).Add(color.Bold)
	red        = color.New(color.FgHiRed)
	boldRed    = color.New(color.FgHiRed).Add(color.Bold)
	boldCyan   = color.New(color.FgCyan).Add(color.Bold)
	cyan       = color.New(color.FgCyan)
)

const (
	randomEn = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	YYYYMMDD = "2006-01-02"
)

type Results struct {
	Time        string
	Url         string
	Threads     string
	Requests    string
	Result      string
	AvgResponse string
	MaxResponse string
	MinResponse string
}

type nonBlocking struct {
	Response    *http.Response
	Error       error
	ElapsedTime float64
}

type Configuration struct {
	url            string
	method         string
	requests       int64
	clients        int64
	randomJson     map[string]string
	client_timeout int64
}

func init() {
	flag.StringVar(&hp, "h", "", "see options")
	flag.Int64Var(&requests, "r", 10000, "The number of request")
	flag.Int64Var(&clients, "c", 100, "The number of thread")
	flag.Int64Var(&client_timeout, "t", 30, "Request Timeout(second)")
	flag.StringVar(&url, "u", "", "Request URL")
	flag.StringVar(&body, "j", "", "Inform us your Json key/type to FUZZING with no space between key and type\n\"[KEY1,TYPE1,KEY2,TYPE2,...]\" "+infos)
	openServer = flag.Bool("s", false, "opening log server")
	rand.Seed(time.Now().UnixNano())
}

func NewConfiguration() *Configuration {

	if *openServer {
		resultMap := make(map[string]Results)

		fs := http.FileServer(http.Dir("public"))
		http.Handle("/public/", http.StripPrefix("/public/", fs))

		files, err := ioutil.ReadDir("./public/graph/")

		var fileNames []string
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			fileNames = append(fileNames, strings.Split(file.Name(), ".")[0])
		}

		content, err := ioutil.ReadFile("./public/results/rs.text")

		if err != nil {
			log.Fatal(err)
		}
		sc := strings.Split(string(content), "\n")
		for _, s := range sc {
			if s == "" {
				continue
			}
			splitedResults := strings.Split(s, "|")
			resultMap[splitedResults[0]] = Results{
				Time:        splitedResults[0],
				Url:         splitedResults[1],
				Threads:     splitedResults[2],
				Requests:    splitedResults[3],
				Result:      splitedResults[4],
				AvgResponse: splitedResults[5],
				MaxResponse: splitedResults[6],
				MinResponse: splitedResults[7],
			}
		}

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			tmpl := template.Must(template.ParseFiles("templates/layout.html"))
			var rs []Results
			for _, fileName := range fileNames {
				rs = append(rs, resultMap[fileName])
			}
			tmpl.Execute(w, rs)
		})

		server := &http.Server{Addr: ":8022", Handler: nil}
		boldYellow.Printf("Local Server Opened! check here : http://localhost%s\n", server.Addr)

		go func() {
			server.ListenAndServe()
		}()

		exit := ""
		fmt.Print("Do you want to exit? [y] ")
		fmt.Scanln(&exit)

		if exit == "y" {
			err := server.Shutdown(context.Background())
			// can't do much here except for logging any errors
			if err != nil {
				log.Printf("error during shutdown: %v\n", err)
			}
			os.Exit(1)
			// notifying the main goroutine that we are done
		}
	}

	if hp != "" {
		flag.Usage()
		os.Exit(1)
	}

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
	case "boolean":
		return false
	}
	return "null"
}

func returnRandomByTypes(column_type string) string {
	switch column_type {
	case "int":
		return strconv.Itoa(rand.Intn(1000))
	case "double":
		return fmt.Sprintf("%f", rand.Float64())
	case "string":
		return "\"" + RandStringEn(10) + "\""
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
	boldRed.Printf("[Key Error] %s 의 타입이 없습니다. 아래의 타입 중 하나를 선택해서 표기해주세요.\n", key)
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
		nb <- nonBlocking{
			Response:    res,
			Error:       err,
			ElapsedTime: elapsedTime,
		}

		if err != nil {
			boldRed.Printf("ERROR MESSAGE:%s\n", err.Error())
			os.Exit(1)
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

func ensureDir(dirName string) error {
	err := os.Mkdir(dirName, 0777)
	if err == nil {
		return nil
	}
	if os.IsExist(err) {
		// check that the existing path is a directory
		info, err := os.Stat(dirName)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return errors.New("path exists but is not a directory")
		}
		return nil
	}
	return err
}

func lineBase(details string, items []opts.LineData, tm []time.Time, numThread string, numRequest string) *charts.Line {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "[gotybench] HTTP Benchmark Graph ( numThread:" + numThread + ", numRequest:" + numRequest + " )", Subtitle: "Response Time (Second) "}),
	)
	line.SetXAxis(tm).
		AddSeries("Response Time ", items).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))
	f, err := os.Create("public/graph/" + details + ".html")
	if err != nil {
		log.Print(err)
	}
	_ = line.Render(f)
	return line
}

func HandleResponse(url string, nb chan nonBlocking, wg *sync.WaitGroup, threads int, requests int) {

	avg := 0.0
	max_response_time := 0.0
	min_response_time := -1.0
	num := 0
	writer := uilive.New()
	writer.Start()

	m := make(map[int]int)
	items := make([]opts.LineData, 0)
	var tm []time.Time
	elaspedTime := 0.0

	for get := range nb {
		num += 1
		if (requests-num)%10 == 0 {
			fmt.Fprintf(writer, "Listening server's response .. (%d/%d)\n", num, requests)
		}
		if get.Error != nil {
			// log.Println(get.Error)
		} else {
			// 웹 그래프용 객체
			items = append(items, opts.LineData{Value: get.ElapsedTime})
			tm = append(tm, time.Now())
			avg = float64(num-1)/float64(num)*avg + 1/float64(num)*float64(get.ElapsedTime) // 평균 필터
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
			writer.Stop()
			writer.Flush()

			boldGreen.Printf("\n [Results]\n")
			fmt.Println("---------------------------------------------------------")
			fmt.Println("| Response Status \t| Count \t| Percent \t|")
			for k, v := range m {
				fmt.Printf("| ")
				if 200 <= k && k < 300 {
					green.Printf("%d \t\t\t", k)
					fmt.Printf("| ")
					green.Printf("%d/%d \t", v, requests)
					fmt.Printf("|")
					fmt.Printf(" %.1f%%\t", float64(v)/float64(requests)*100.0)
					fmt.Printf("|\n")
				} else {
					red.Printf("%d \t\t\t", k)
					fmt.Printf("| ")
					red.Printf("%d/%d \t", v, requests)
					fmt.Printf("|")
					fmt.Printf(" %.1f%%\t", float64(v)/float64(requests)*100.0)
					fmt.Printf("|\n")
				}

			}
			fmt.Println("---------------------------------------------------------")
			green.Printf("- Average response time \t: %.2f ms\n", avg*1000)
			green.Printf("- Max response time     \t: %.2f ms\n", max_response_time*1000)
			green.Printf("- Min response time     \t: %.2f ms\n", min_response_time*1000.0)

			dl := "|"
			details := time.Now().Format("2006-01-02:15:04:05")

			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)
			boldYellow.Printf("\n [Memory Usage]\n")
			yellow.Printf("- Heap size = %v MB\n", bToMb(mem.Alloc))
			yellow.Printf("- Cumulative Heap size = %v MB\n", bToMb(mem.TotalAlloc))
			yellow.Printf("- All goroutine size = %v MB\n", bToMb(mem.Sys))
			yellow.Printf("- GC cycle 횟수 = %v\n\n", mem.NumGC)

			path, err := filepath.Abs("./")

			if err := ensureDir(path + "/public/graph"); err != nil {
				fmt.Println("1 Directory creation failed with error: " + err.Error())
				os.Exit(1)
			}
			if err := ensureDir(path + "/public/results"); err != nil {
				fmt.Println("2 Directory creation failed with error: " + err.Error())
				os.Exit(1)
			}
			line2web = lineBase(details, items, tm, strconv.Itoa(threads), strconv.Itoa(requests))
			f, err := os.OpenFile("public/results/rs.text",
				os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Println(err)
			}
			defer f.Close()
			if _, err := f.WriteString(details + dl +
				url + dl +
				strconv.Itoa(threads) + dl +
				strconv.Itoa(requests) + dl +
				fmt.Sprintf("%v", m) + dl +
				fmt.Sprintf("%.2f", avg*1000) + dl +
				fmt.Sprintf("%.2f", max_response_time*1000) + dl +
				fmt.Sprintf("%.2f", min_response_time*1000) + "\n"); err != nil {
				log.Println(err)
			}

		}
		wg.Done()
	}

}

func generateLineItems() []opts.LineData {
	items := make([]opts.LineData, 0)
	for i := 0; i < 7; i++ {
		items = append(items, opts.LineData{Value: rand.Intn(300)})
	}
	return items
}

func httpserver(w http.ResponseWriter, _ *http.Request) {
	line2web.Render(w)
}

// Local address of the running system
func getLocalAddress() net.IP {
	con, error := net.Dial("udp", "8.8.8.8:80")
	if error != nil {
		log.Fatal(error)
	}
	defer con.Close()

	localAddress := con.LocalAddr().(*net.UDPAddr)

	return localAddress.IP
}

func main() {
	wg := &sync.WaitGroup{}
	var mutex = &sync.RWMutex{}
	rand.Seed(time.Now().UnixNano())
	flag.Parse()

	configuration := NewConfiguration()

	boldCyan.Printf("\n [Properties]\n")
	cyan.Printf("- Max parallelism : %d\n", MaxParallelism())

	cyan.Printf("- Request url : %s\n", configuration.url)
	cyan.Printf("- The number of HTTP Requests : %d\n", configuration.requests)
	cyan.Printf("- The number of threads : %d\n", configuration.clients)

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
	t.MaxIdleConns = 1000    // connection pool 크기
	t.MaxConnsPerHost = 1000 // 호스트 별 최대 할당 connection
	t.MaxIdleConnsPerHost = 100
	t.IdleConnTimeout = time.Duration(configuration.client_timeout) * time.Second

	// 클라이언트 설정 및 timeout
	client := &http.Client{
		Timeout:   time.Duration(configuration.client_timeout) * time.Second,
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

	wg.Add(int(configuration.requests))
	go HandleResponse(requestURL, nb, wg, int(number_worker), int(configuration.requests))
	runtime.GC()
	wg.Wait()

	elapsedTime := time.Since(startTime).Seconds()

	var exit string
	laddr := getLocalAddress().String()
	fmt.Printf("Finished! ( Total Elapsed Time : %.4f seconds ) \nNow you can see response time series graph in local machine => http://"+laddr+":8022 \n\n", elapsedTime)

	wg.Add(1)

	http.HandleFunc("/", httpserver)
	server := &http.Server{Addr: ":8022", Handler: nil}

	go func() {
		server.ListenAndServe()
	}()

	fmt.Print("Do you want to exit? [y] ")
	fmt.Scanln(&exit)

	if exit == "y" {
		err := server.Shutdown(context.Background())
		// can't do much here except for logging any errors
		if err != nil {
			log.Printf("error during shutdown: %v\n", err)
		}
		// notifying the main goroutine that we are done
		wg.Done()
	}

	wg.Wait()
}
