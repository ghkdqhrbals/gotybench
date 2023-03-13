package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"text/template"
)

type User struct {
	Name       string
	Occupation string
}

func main2() {

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
			fmt.Println(resultMap[fileName])
			rs = append(rs, resultMap[fileName])
		}
		tmpl.Execute(w, rs)
	})

	http.ListenAndServe(":8022", nil)
}
