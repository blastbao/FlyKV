package main

import (
	"BitcaskDB"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

var db *BitcaskDB.DB //DB实例
func init() {
	//初始化db实例
	var err error
	options := BitcaskDB.DefaultOperations
	dir, _ := os.MkdirTemp("", "bitcask-go-http")
	options.DirPath = dir
	db, err = BitcaskDB.Open(options)
	if err != nil {
		panic(fmt.Sprintf("failed to open db: %v", err))
	}

}

func handlePut(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var kv map[string]string

	if err := json.NewDecoder(request.Body).Decode(&kv); err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	for key, value := range kv {
		if err := db.Put([]byte(key), []byte(value)); err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			log.Printf("failed to put kv in db: %v\n", err)
			return
		}
	}
}

func handleGet(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := request.URL.Query().Get("key")

	value, err := db.Get([]byte(key))
	if err != nil && err != BitcaskDB.ErrKeyNotFound {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		log.Printf("failed to get kv in db: %v\n", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(writer).Encode(string(value))
}

func handleDelete(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodDelete {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := request.URL.Query().Get("key")

	err := db.Delete([]byte(key))
	if err != nil && err != BitcaskDB.ErrKeyIsEmpty {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		log.Printf("failed to get kv in db: %v\n", err)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(writer).Encode("OK")
}

func handleListKeys(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	keys := db.ListKeys()
	writer.Header().Set("Content-Type", "application/json")
	var result []string
	for _, k := range keys {
		result = append(result, string(k))
	}
	_ = json.NewEncoder(writer).Encode(result)
}

func handleStat(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stat := db.Stat()
	writer.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(writer).Encode(stat)
}

func main() {
	//注册回调函数
	http.HandleFunc("/bitcask/put", handlePut)
	http.HandleFunc("/bitcask/get", handleGet)
	http.HandleFunc("/bitcask/delete", handleDelete)
	http.HandleFunc("/bitcask/listkeys", handleListKeys)
	http.HandleFunc("/bitcask/stat", handleStat)
	//启动服务并监听
	http.ListenAndServe("localhost:8080", nil)

}
