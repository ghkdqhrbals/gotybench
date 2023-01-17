# Test with multiple http requests!(대량 HTTP request 전송 테스트)

This program is written in Golang. **You can send multiple HTTP requests by simply running `docker-compose up`!**
(해당 프로젝트는 Golang으로 제작되었으며, `docker-compose up`를 통해 다량의 http 패킷을 전송할 수 있습니다)

# Test Steps
1. Edit `main.go` with your own json body!
```golang
...
// For your own json body, edit here!(보내고 싶은 json 구조에 맞게 수정하시면 됩니다)
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
		// ---------- For your own json body, edit here!(랜덤값을 설정하는 부분입니다) ---------
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
2. Run `docker-compose up`

# Test Results

```
test-multiple-http-request  | Request url: http://127.0.0.1:8080/auth/user
test-multiple-http-request  | The number of HTTP Requests: 10000
test-multiple-http-request  | The number of threads: 100
test-multiple-http-request  | Proceeding! Please wait until getting all the responses
test-multiple-http-request  | Elapsed Time: 30.533003028
test-multiple-http-request  | Response status code:  200 , How many?:  10000
```

As you can see here, we send 10K http request to our server and get responses with status code 200 within 30 seconds.
(총 10K개의 rest api call 을 진행하였고, 100개의 go-routine으로 진행한 결과입니다. 약 30.5초가 소요되었으며, 전부 200 status code를 반환 받은 것을 확인할 수 있습니다)

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
When your server run in docker container and expose port through host, this docker setting will be fine.(프록시 서버가 컨테이너로 실행되고, 호스트한테 프록시 포트가 expose되어있다면, 그대로 `docker-compose up`을 실행하시면 됩니다)

However, when your server expose port inside the container, you should change compose setting by delete `network_mode: "host"` and set alias with your server.(하지만 내부 포트로 expose되어 있다면, 같은 네트워크로 묶어줘야합니다)