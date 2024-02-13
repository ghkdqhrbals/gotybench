// Copyright 2024 @ghkdqhrbals Authors
// This file is part of the gotybench library.
//
// The gotybench library is free software: you can redistribute it and/or modify
// it under the terms of the MIT License.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"gonum.org/v1/gonum/stat"

	"testapi.com/m/benchmark"
	"github.com/fatih/color"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/gosuri/uilive"
	
	dynamicstruct "github.com/ompluscator/dynamic-struct"

)

var (
	flags *benchmark.FlagParam
	line2web       *charts.Line // Response 시간을 웹에 표시

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

type Results struct {
	Time        string
	Url         string
	Threads     string
	Requests    string
	Result      string
	AvgResponse string
	MaxResponse string
	MinResponse string
	Info        string
}

type nonBlocking struct {
	Response    *http.Response
	Error       error
	ElapsedTime float64
}

func init() {
	flags = benchmark.GetFlags()
	rand.Seed(time.Now().UnixNano())
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

// byte to MegaByte
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// 요청을 전송하는 스레드 함수
func worker(nb chan nonBlocking, jsons map[string]string, mutex *sync.RWMutex, contexts *sync.Map, wg *sync.WaitGroup, requestURL string, client *http.Client, transferRatePerSecond int) {
	defer wg.Done()
	// 로컬 맵 생성
	m := make(map[int]int)
	instance := dynamicstruct.NewStruct().Build().New()

	for key, val := range jsons {
		temp_instance := dynamicstruct.NewStruct().AddField(strings.ToUpper(key), returnDefaults(val), `json:"`+key+`"`).Build().New()
		instance = dynamicstruct.MergeStructs(instance, temp_instance).Build().New()
	}

	for i := 1; i < (transferRatePerSecond)+1; i++ {

		data := benchmark.RandomGeneratedData(jsons)
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
		req.Close = true
		if err != nil {
			fmt.Println(err)
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
			// os.Exit(1)
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

func reverseFloat64Slice(slice []float64) {
	for i := len(slice)/2 - 1; i >= 0; i-- {
		opp := len(slice) - 1 - i
		slice[i], slice[opp] = slice[opp], slice[i]
	}
}


// html 그래프 생성
func lineBase(details string, items []opts.LineData, tm []time.Time, numThread string, numRequest string) *charts.Line {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "[gotybench] HTTP Benchmark Graph ( numThread:" + numThread + ", numRequest:" + numRequest + " )", Subtitle: "Response Time (ms) "}),
	)

	line.SetXAxis(tm).
		AddSeries("Response Time ", items).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))
	line.SetGlobalOptions(charts.WithYAxisOpts(opts.YAxis{GridIndex: 0, Name: "", Type: "value", Show: true, Scale: true, SplitLine: &opts.SplitLine{Show: true}}),
		charts.WithTooltipOpts(opts.Tooltip{Show: true, Trigger: "axis", AxisPointer: &opts.AxisPointer{Type: "cross"}}))

	//SetXAxisOptions(charts.WithXAxisOpts(opts.XAxis{AxisLabel: &opts.AxisLabel{Interval: "auto", Rotate: 45, Show: true, Formatter: nil}, Position: "bottom"}))
	f, err := os.Create("public/graph/" + details + ".html")
	if err != nil {
		log.Print(err)
	}

	_ = line.Render(f)
	return line
}

// 각 HTTP 요청스레드에서 받은 채널응답을 처리하는 함수
func HandleResponse(url string, nb chan nonBlocking, wg *sync.WaitGroup, threads int, requests int, info string) {
	avg := 0.0 // 평균 응답시간
	max_response_time := 0.0 // 최대 응답시간
	min_response_time := -1.0 // 최소 응답시간
	num := 0 // 요청 누적 수
	writer := uilive.New() // 실시간으로 터미널 출력해주는 객체
	writer.Start()
	responseTimeList := make([]float64, 0) // 응답시간 리스트
	// tps 모니터링 boolean 채널
	done := make(chan bool)

	responseStatusMap := make(map[int]int) // 응답 상태 코드 카운트
	items := make([]opts.LineData, 0) // 웹 그래프용 객체
	var tm []time.Time
	elaspedTime := 0.0 // 총 응답시간
	tpsList := make([]float64, 0) // TPS 리스트
	//tpsTimer := time.NewTicker(time.Second)
	//startTime := time.Now()
	transaction := 0

	go func() {
		tpsTimer := time.NewTicker(time.Second)
		defer tpsTimer.Stop()
		for {
			select {
			case <-tpsTimer.C:
				// 1초마다 실행될 코드 작성
				// 예를 들어, 특정 메시지를 출력하거나 필요한 작업을 수행할 수 있습니다.
				tpsList = append(tpsList, -float64(transaction))
				transaction = 0
			case <-done:
				return // 타이머 종료
			}
		}
	}()

	for get := range nb {
		num += 1
		transaction += 1
		if (requests-num)%10 == 0 {
			fmt.Fprintf(writer, "Listening server's response .. (%d/%d)\n", num, requests)
		}
		if get.Error != nil {
		} else {
			// 웹 그래프용 객체
			items = append(items, opts.LineData{Value: get.ElapsedTime*1000})
			tm = append(tm, time.Now())
			avg = float64(num-1)/float64(num)*avg + 1/float64(num)*float64(get.ElapsedTime) // 평균 필터
			if min_response_time < 0 {
				min_response_time = get.ElapsedTime
			}
			max_response_time = math.Max(max_response_time, get.ElapsedTime)
			min_response_time = math.Min(min_response_time, get.ElapsedTime)

			// log.Println(get.Response.Status)
			responseStatusMap[get.Response.StatusCode] += 1
			elaspedTime += get.ElapsedTime
			responseTimeList = append(responseTimeList, get.ElapsedTime)
		}


		if num == requests {
			writer.Stop()
			writer.Flush()
			done<- true // 타이머 종료

			boldGreen.Printf("\n [Results]\n")
			fmt.Println("---------------------------------------------------------")
			fmt.Println("| Response Status \t| Count \t| Percent \t|")
			for k, v := range responseStatusMap {
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

			sort.Float64s(responseTimeList)
			p90 := stat.Quantile(0.90, stat.Empirical, responseTimeList, nil)
			p95 := stat.Quantile(0.95, stat.Empirical, responseTimeList, nil)
			p99 := stat.Quantile(0.99, stat.Empirical, responseTimeList, nil)
			p50 := stat.Quantile(0.5, stat.Empirical, responseTimeList, nil)

			fmt.Println("---------------------------------------------------------")
			green.Printf("- Average response time \t: %.2f ms\n", avg*1000)
			green.Printf("- Max response time     \t: %.2f ms\n", max_response_time*1000)
			green.Printf("- Min response time     \t: %.2f ms\n", min_response_time*1000.0)
			green.Printf("- Response Time 50th percentile       \t: %.2f ms\n", p50*1000)
			green.Printf("- Response Time 90th percentile       \t: %.2f ms\n", p90*1000)
			green.Printf("- Response Time 95th percentile       \t: %.2f ms\n", p95*1000)
			green.Printf("- Response Time 99th percentile       \t: %.2f ms\n", p99*1000)


			// TPS 퍼센타일 계산
			sort.Float64s(tpsList)
			tpsP95 := stat.Quantile(0.95, stat.Empirical, tpsList, nil)
			tpsP99 := stat.Quantile(0.99, stat.Empirical, tpsList, nil)
			tpsP90 := stat.Quantile(0.90, stat.Empirical, tpsList, nil)
			tpsP50 := stat.Quantile(0.50, stat.Empirical, tpsList, nil)
			fmt.Printf("- TPS 50th percentile       \t: %.2f\n", -tpsP50)
			fmt.Printf("- TPS 90th percentile       \t: %.2f\n", -tpsP90)
			fmt.Printf("- TPS 95th percentile       \t: %.2f\n", -tpsP95)
			fmt.Printf("- TPS 99th percentile       \t: %.2f\n", -tpsP99)

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
				fmt.Sprintf("%v", responseStatusMap) + dl +
				fmt.Sprintf("%.2f", avg*1000) + dl +
				fmt.Sprintf("%.2f", max_response_time*1000) + dl +
				fmt.Sprintf("%.2f", min_response_time*1000) + dl +
				info + "\n"); err != nil {
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

	configuration := benchmark.NewConfiguration()
	// http 전송 url
	requestURL := configuration.Url
	// 실행횟수 설정
	transferRatePerSecond := configuration.Requests
	// 실행 스레드 개수 설정
	number_worker := configuration.Threads

	printProperties(requestURL, transferRatePerSecond, number_worker)

	// 전체 실행 개수
	nb := make(chan nonBlocking, number_worker)


	// 스레드 간 공유 맵
	var contexts = &sync.Map{}
	startTime := time.Now()

	wg.Add(int(transferRatePerSecond))
	go HandleResponse(requestURL, nb, wg, int(number_worker), int(transferRatePerSecond), configuration.Info)

	// 멀티 스레드 http request
	for i := 0; i < int(number_worker); i++ {
		wg.Add(1)
		go worker(nb, configuration.RandomJson, mutex, contexts, wg, requestURL,
			httpPropertiesSetting(configuration), int(transferRatePerSecond/number_worker))
	}
	// 세마포어 대기열 추가
	//wg.Add(int(transferRatePerSecond))

	runtime.GC()
	// 세마포어 대기
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

func httpPropertiesSetting(configuration *benchmark.Configuration) *http.Client {
	t := &http.Transport{
		MaxIdleConns:        5,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     time.Duration(configuration.RequestTimeout) * time.Second,
	}

	// 클라이언트 설정 및 timeout
	client := &http.Client{
		Timeout:   time.Duration(configuration.RequestTimeout) * time.Second,
		Transport: t,
	}
	return client
}

func printProperties(requestURL string, transferRatePerSecond int64, number_worker int64) {
	boldCyan.Printf("\n [Properties]\n")
	cyan.Printf("- Max parallelism : %d\n", MaxParallelism())
	cyan.Printf("- Request url : %s\n", requestURL)
	cyan.Printf("- The number of HTTP Requests : %d\n", transferRatePerSecond)
	cyan.Printf("- The number of threads : %d\n", number_worker)
}
