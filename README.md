# Introduction(gotybench)

저는 불편한 툴들을 자동화하는 것을 좋아합니다. 여러 부하 분산을 테스팅할 수 있는 툴을 찾다가 **랜덤한 값을 가지는 json으로 다량의 request를 전송하는 툴**을 찾기 힘들었습니다. 즉, 기존 툴들은 Fuzzing하기 힘들었어요.

그래서 만들었어요!

해당 프로젝트는 Golang으로 제작되었으며, 다량의 랜덤한 http 패킷을 전송할 수 있습니다.

# Usage
1. run `go run main.go` in your terminal

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

1. Run `go run main.go -j "[userId,string,userPw,string,mail,string,userName,string]" -r 10000 -c 1000 -u http://127.0.0.1:8080/auth/user`

	```
	Alloc = 0 MiB	TotalAlloc = 0 MiB	Sys = 8 MiB	NumGC = 0
		Properties
	- Max parallelism : 8
	- Request url : http://127.0.0.1:8080/auth/user
	- The number of HTTP Requests : 10000
	- The number of threads : 1000
	=> Proceeding... Please wait until getting all the responses 10000
	[RESULTS] Response status code:  200 , How many?:  10000
	```




# Test Results

총 10K개의 rest api call 을 진행하였고, 100개의 go-routine으로 진행한 결과입니다. 약 30.5초가 소요되었으며, 전부 200 status code를 반환 받은 것을 확인할 수 있습니다

(As you can see here, we send 10K http request to our server and get responses with status code 200 within 30 seconds.)

* Flow of example test

```
configurate app.env
--(Viper)--> go build
--> build docker images
--> run docker container(network:host)
--> nginx
--> auth-server
```

# **Notice!**
프록시 서버가 컨테이너로 실행되고, 호스트한테 프록시 포트가 expose되어있다면, 그대로 `docker-compose up`을 실행하시면 됩니다(When your server run in docker container and expose port through host, this docker setting will be fine.)

하지만 내부 포트로 expose되어 있다면, 같은 네트워크로 묶어줘야합니다(However, when your server expose port inside the container, you should change compose setting by delete `network_mode: "host"` and set alias with your server.)