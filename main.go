package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	// 打印程序耗时
	now := time.Now()
	defer func() {
		fmt.Printf("程序耗时： %v\n", time.Since(now))
	}()
	// 打开CSV文件
	file, err := os.Open("data.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	

	// os.Create 默认会覆盖原文件；如果不存在就会新建
	goodFile := mustFile(os.Create("good.csv"))
	defer goodFile.Close()

	badFile := mustFile(os.Create("bad.csv"))
	defer badFile.Close()

	badCsv := csv.NewWriter(badFile)
	defer badCsv.Flush() // 从缓存中刷新到badFile。

	goodCsv := csv.NewWriter(goodFile)
	defer goodCsv.Flush()

	// 写入表头
	// 创建CSV读取器
	reader := csv.NewReader(file)
	record, err := reader.Read()
	if(err!=nil){
		fmt.Println("Error reading CSV:", err)
		return
	}
	badCsv.Write(record)
	goodCsv.Write(record)


	taskCh:= make(chan []string)
	goodCh:= make(chan []string)
	badCh:= make(chan []string)
	var wg sync.WaitGroup

	go func(){
		
		// 逐行读取CSV文件
		for {
			record, err := reader.Read()
			
			if(err!=nil){
				fmt.Println("Error reading CSV:", err)
				break
			}
			taskCh<-record
		}
		close(taskCh)
	}()
	

	//开启100个worker
	for range 100{
		wg.Add(1)
		go Worker(taskCh,goodCh,badCh,&wg)
	}

	goodDone:=make(chan struct{})
	badDone:=make(chan struct{})

	go func(){
		for task:=range goodCh{

			goodCsv.Write(task)
			fmt.Println("good:",task)
			
		}
		close(goodDone)
	}()

	go func(){
		for task:=range badCh{
			badCsv.Write(task)
			fmt.Println("bad:",task)
		}
		close(badDone)

	}()

	wg.Wait()

	close(goodCh)
	close(badCh)

	<-goodDone
	<-badDone
}

func Worker(taskCh chan []string, goodCh chan []string, badCh chan []string, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	for task := range taskCh {
		if check(task[4]){
			goodCh<-task
		}else{
			badCh<-task
		}
	}
}

// 工具函数，简便写法。must设计模式。
func mustFile(f *os.File, err error) *os.File {
	if err != nil {
		panic(err)
	}
	return f
}

// func check(url string)bool{
// 	// 发送HTTP GET请求
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		return false
// 	}
// 	defer resp.Body.Close()

// 	// 检查HTTP响应状态码
// 	if resp.StatusCode == http.StatusOK {
// 		return true
// 	} else {
// 		return false
// 	}
// }
func check(url string) bool {
	// 创建一个 HTTP 客户端并设置超时时间为 3 秒
	client := &http.Client{Timeout: 5 * time.Second}

	// 发送 HTTP GET 请求
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// 检查 HTTP 响应状态码
	if resp.StatusCode == http.StatusOK {
		return true
	} else {
		return false
	}
}


