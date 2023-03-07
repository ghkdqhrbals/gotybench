# Introduction(gotybench)

저는 불편한 툴들을 자동화하는 것을 좋아합니다 :)

여러가지 툴들을 살펴보다가 랜덤한 json body를 가지는 

# Test your server with multiple http requests(대량 HTTP request를 통한 서버 부하 테스트)

해당 프로젝트는 Golang으로 제작되었으며, `docker-compose up`를 통해 다량의 http 패킷을 전송할 수 있습니다

(This program is written in Golang. **You can send multiple HTTP requests by simply running `docker-compose up`**)


# Usage
1. 전송url, http 전송개수, 실행 스레드 개수를 `app.env`에서 수정하세요(Edit `app.env` RequestUrl, reqeust frequency, the number of worker process)

```
RequestURL=http://127.0.0.1:8080/auth/user
RequestNum=10000
WorkerNum=100
```

1. `main.go`에서 전송하고자 하는 json struct와 랜덤값들을 설정해주시면 됩니다(Edit `main.go` with your own json body)
```golang
...
// 보내고 싶은 json 구조에 맞게 수정하시면 됩니다(For your own json body, edit here!)
type User struct {
	UserId   string `json:"userId"`
	UserName string `json:"userName"`
	Email    string `json:"email"`
	UserPw   string `json:"userPw"`
}
// ----------------------------------
...
func worker(contexts *sync.Map, wg *sync.WaitGroup, requestURL string, client *http.Client, transferRatePerSecond int, number_worker int) {
	...
	for i := 0; i < transferRatePerSecond; i++ {
		// ---------- 랜덤값을 설정하는 부분입니다(For your own json body, edit here!) ---------
		s := &User{
			UserId:   RandStringEn(8),
			UserName: RandStringKr(1) + RandStringKr(2),
			Email:    RandStringEn(5) + "@gmail.com",
			UserPw:   RandStringEn(10),
		}
        // ------------------------------------------------------------------------------
		...
	}
    ...
}
```
1. 루트 디렉토리에서 `docker-compose up`을 실행하시고 결과를 확인해보세요(Run `docker-compose up` in root directory and see what happen!)

# Test Results

```
test-multiple-http-request  | Request url: http://127.0.0.1:8080/auth/user
test-multiple-http-request  | The number of HTTP Requests: 10000
test-multiple-http-request  | The number of threads: 100
test-multiple-http-request  | Proceeding! Please wait until getting all the responses
test-multiple-http-request  | Elapsed Time: 30.533003028
test-multiple-http-request  | Response status code:  200 , How many?:  10000
```

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