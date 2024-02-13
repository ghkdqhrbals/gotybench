package benchmark

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var SupportJsonTypeString = fmt.Sprintf("\n| - Support Json Type -\t|\n| * int \t\t|\n| * float \t\t|\n| * string \t\t|\n| * boolean \t\t|")

// 키에 대한 타입이 없을 때 info 를 출력하고 에러를 반환합니다.
func errorKeyType(key string) {
	boldRed.Printf("[Key Error] %s 의 타입이 없습니다. 아래의 타입 중 하나를 선택해서 표기해주세요.\n", key)
	fmt.Println("| - Support Json Type -\t|")
	fmt.Println("| * int \t\t|")
	fmt.Println("| * float \t\t|")
	fmt.Println("| * string \t\t|")
	fmt.Println("| * boolean \t\t|")
	os.Exit(1)
}

// 현재 고루틴의 고유 ID를 반환합니다.
func Goid() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

