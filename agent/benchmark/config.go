// Copyright 2024 @ghkdqhrbals Authors
// This file is part of the gotybench library.
//
// The gotybench library is free software: you can redistribute it and/or modify
// it under the terms of the MIT License.
package benchmark

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
)

var (
	flags *FlagParam
)

type FlagParam struct {
	Help    string
	Url    string
	Info     string
	Method   string
	Requests   int64
	Threads        int64
	RequestTimeout int64
	Body		   string
	OpenServer     bool
}

type Configuration struct {
    Url    string
    Info     string
    Method   string
    Requests   int64
    Threads        int64
	RandomJson     map[string]string
    RequestTimeout int64
}

func GetFlags() *FlagParam {
	return flags
}

func init() {
	flags = &FlagParam{}
	flag.StringVar(&flags.Help, "help", "", "See various options")
	flag.Int64Var(&flags.Requests, "request", 10000, "Request count")
	flag.Int64Var(&flags.Threads, "thread", 100, "Thread count")
	flag.Int64Var(&flags.RequestTimeout, "timeout", 30, "Request timeout(second)")
	flag.StringVar(&flags.Url, "url", "", "URL for request")
	flag.StringVar(&flags.Info, "comment", "", "Comments of this test")
	flag.StringVar(&flags.Body, "body", "", "JSON body contents like \n\"[key1,type1,key2,type2,...]\" "+SupportJsonTypeString)
	flag.BoolVar(&flags.OpenServer,"server", false, "Opening log server")
	flag.Parse()
}

func NewConfiguration() *Configuration {

	if flags.OpenServer {
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
			splitedResults = append(splitedResults, "")
			resultMap[splitedResults[0]] = Results{
				Time:        splitedResults[0],
				Url:         splitedResults[1],
				Threads:     splitedResults[2],
				Requests:    splitedResults[3],
				Result:      splitedResults[4],
				AvgResponse: splitedResults[5],
				MaxResponse: splitedResults[6],
				MinResponse: splitedResults[7],
				Info:        splitedResults[8],
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

	if flags.Help != "" || flags.Url == "" || flags.Body == "" {
		flag.Usage()
		os.Exit(1)
	}

	configuration := initConfiguration()

	setConfiguration(configuration)

	if flags.Body != "" {
		flags.Body = strings.Replace(flags.Body, "[", "", -1)
		flags.Body = strings.Replace(flags.Body, "]", "", -1)
		temps := strings.Split(flags.Body, ",")
		for i := 0; i < len(temps); i += 2 {
			if i > len(temps)-2 {
				errorKeyType(temps[i])
				os.Exit(1)
			}
			configuration.RandomJson[temps[i]] = temps[i+1]
		}
	}

	return configuration
}

func initConfiguration() *Configuration {
	configuration := &Configuration{
		Url:            "",
		Method:         http.MethodPost,
		Requests:       int64(10000),
		Threads:        int64(100),
		RequestTimeout: int64(30),
		RandomJson:     map[string]string{},
	}
	return configuration
}

func setConfiguration(configuration *Configuration) {
	if flags.Info != "" {
		configuration.Info = flags.Info
	}

	if flags.Url != "" {
		configuration.Url = flags.Url
	}

	if flags.Requests != 10000 {
		configuration.Requests = flags.Requests
	}

	if flags.Threads != 100 {
		configuration.Threads = flags.Threads
	}

	if flags.RequestTimeout != 30 {
		configuration.RequestTimeout = flags.RequestTimeout
	}
}