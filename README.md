# Introduction(gotybench)

![](img/24.gif)

**gotybench(HTTP benchmark tool)은 다음을 목표로 설계 및 제작하였습니다.**

1. **테스트 동시성 보장** : goroutine 경량 멀티 스레드를 사용하였으며, 채널을 통해 통신하도록 설정하였습니다.
2. **다이나믹 Structure 을 통한 Fuzzed Json 오브젝트 생성** : 사용자가 key와 value type들만 설정해주면 자동으로 랜덤한 json 오브젝트를 생성하도록 제작하였습니다.
    * ex) "gotybench -j [userId,string,userAge,int]" : userId의 value를 랜덤한 string으로 설정합니다. 또한 userAge의 value를 랜덤한 int로 설정합니다. 

# Options
| Option | Detail                                                                                                                                                                                                                                                                       |
| ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| -c     | 동시처리가능한 스레드 개수를 해당 옵션으로 설정할 수 있습니다.                                                                                                                                                                                                               |
| -h     | 옵션들의 설명을 확인할 수 있습니다.                                                                                                                                                                                                                                          |
| -j     | 핵심적인 Fuzzing 기능입니다. <br> json object를 해당 옵션으로 key/type을 설정하면, 랜덤한 value의 json obejct가 생성됩니다.<br>Fuzzing이 지원되는 type 은 4가지로 아래와 같습니다.<br>int, float, string, boolean<br>Usage Example<br>ex) `-j "[userId,string,userAge,int]"` |
| -r     | HTTP POST request 개수를 해당 옵션으로 설정할 수 있습니다.                                                                                                                                                                                                                   |
| -t     | 벤치마크 클라이언트의 network connection 의 timeout을 해당 옵션으로 설정할 수 있습니다.                                                                                                                                                                                      |
| -u     | 요청하는 URL을 설정할 수 있습니다.                                                                                                                                                                                                                                           |
| -s     | 로깅된 이전 벤치마크 데이터 및 그래프들을 서버에 띄웁니다.                                                                                                                                                                                                                   |

**이중 특히 `-j` 옵션은 Dynamic Struct를 차용함으로써, 오브젝트의 private 필드는 사용자의 입력값에 따라 구조가 변경됩니다!**

# Before we started, we need to get ...
1. run `go get github.com/fatih/color` for coloring your terminal
2. run `go get -v github.com/gosuri/uilive` for updating process
3. run `go get -u github.com/go-echarts/go-echarts/v2/...` to see graph with responses in timeseries.
4. run `go get github.com/ompluscator/dynamic-struct` to dynamically add field of json structs.

# Usage
1. run `go run main.go` in your terminal and see options

	```bash
	Alloc = 0 MiB	TotalAlloc = 0 MiB	Sys = 8 MiB	NumGC = 0
		Properties
		- Max parallelism : 8
	Usage of /var/folders/h0/_d_zrr0j57x8wmknjb1r6hfm0000gn/T/go-build3252492082/b001/exe/main:
	-c int
			스레드 개수 (default 100)
	-j string
			Json "[KEY1,TYPE1,KEY2,TYPE2,...]" 
	-r int
			요청 개수 (default 10000)
	-t int
			요청 타임아웃(second) (default 30)
	-u string
			URL
	```                                                    
2. choose your options and run

# Example

```bash
$ go run main.go -j "[userId,string,userPw,string,mail,string,userName,string]" -r 10000 -c 1000 -u http://127.0.0.1:8080/auth/user

 [Properties]
- Max parallelism : 8
- Request url : http://127.0.0.1:8080/auth/user
- The number of HTTP Requests : 10000
- The number of threads : 100
Listening server's response .. (10000/10000)

 [Results]
---------------------------------------------------------
| Response Status 	| Count 	| Percent 	|
| 200 			| 10000/10000 	| 100.0%	|
---------------------------------------------------------
- Average response time 	: 110.66 ms
- Max response time     	: 770.32 ms
- Min response time     	: 21.46 ms

 [Memory Usage]
- Heap size = 2 MB
- Cumulative Heap size = 161 MB
- All goroutine size = 22 MB
- GC cycle 횟수 = 48

Finished! ( Total Elapsed Time : 11.4659 seconds ) 
Now you can see response time series graph in local machine => http://localhost:8022 

```

# Results

### **벤치마크 결과 확인**
![img](img/27.png)
### **시간에 따른 응답 RTT 그래프**
![img](img/28.png)
### **벤치마크 로깅**
![img](img/29.png)
