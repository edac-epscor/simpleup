package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func SimpleUp(w http.ResponseWriter, r *http.Request) {
	token := getCookieByName(r.Cookies(), "SSESSa50ea890ba0843ff11a5efb1d948cc4d")
	var username string
	query := "SELECT name FROM users u INNER JOIN sessions s ON u.uid = s.uid WHERE s.sid = '" + token + "';"
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		err := rows.Scan(&username)
		if err != nil {
			log.Fatal(err)
		}
	}
	if username != "" {
		if _, err := os.Stat(UploadDIR + "/" + username); os.IsNotExist(err) {
			os.Mkdir(UploadDIR+"/"+username, 0770)
		}

		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		f, err := os.OpenFile(UploadDIR+"/"+username+"/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)

		if _, err := os.Stat(UploadDIR + "/" + username + "/" + handler.Filename); os.IsNotExist(err) {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusCreated)
		}

	} else if username == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}
}

func SimpleUpForm(w http.ResponseWriter, r *http.Request) {
	token := getCookieByName(r.Cookies(), "SSESSa50ea890ba0843ff11a5efb1d948cc4d")
	var username string
	query := "SELECT name FROM users u INNER JOIN sessions s ON u.uid = s.uid WHERE s.sid = '" + token + "';"
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		err := rows.Scan(&username)
		if err != nil {
			log.Fatal(err)
		}
	}
	if username != "" {

		uppage := `<html><head><title>Upload file</title></head><body><form enctype="multipart/form-data" action="https://129.24.63.76/simpleup/" method="post"><input type="file" name="uploadfile" /><input type="hidden" name="token" value="{{.}}"/><input type="submit" value="upload" /></form></body></html>`
		fmt.Fprintln(w, uppage)
	} else if username == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}
}

func getCookieByName(cookie []*http.Cookie, name string) string {
	cookieLen := len(cookie)
	result := ""
	for i := 0; i < cookieLen; i++ {
		if cookie[i].Name == name {
			result = cookie[i].Value
		}
	}
	return result
}
