package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
)

func SimpleUp(w http.ResponseWriter, r *http.Request) {

	dump, err := httputil.DumpRequest(r, false)

	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	log.Println(string(dump))
	token := getCookieByName(r.Cookies(), cookieid)
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
		log.Println(username + " is logged in.")
		log.Println("Username " + username + " attempting an upload...")
		if _, err := os.Stat(UploadDIR + "/" + username); os.IsNotExist(err) {
			log.Println("Directory /uploads/" + username + " does not exist so making it...")
			os.Mkdir(UploadDIR+"/"+username, 0770)

		} else {
			log.Println("Directory /uploads/" + username + " already exists.")
		}

		log.Println("Setting max memory for Multipart form.")
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("file")

		if err != nil {
			log.Println(err)
		         w.WriteHeader(http.StatusNoContent)
			return
		}
		if _, err := os.Stat(UploadDIR + "/" + username + "/" + handler.Filename); os.IsNotExist(err) {
			defer file.Close()

			log.Println("File is  " + handler.Filename)
			if handler.Filename == "" {
				log.Println(handler.Filename + "is empty!!!. This is a problem!")
				w.WriteHeader(http.StatusNoContent)
				return
			}

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

		} else {
			w.WriteHeader(http.StatusFound)
		}

	} else if username == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}
	//} else {
	//	w.WriteHeader(http.StatusFound)
	//}

}

func SimpleUpForm(w http.ResponseWriter, r *http.Request) {

	token := getCookieByName(r.Cookies(), cookieid)

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

		uppage := `<html>
<head>
<title>Upload file</title>
</head>
 <body>
  <form enctype="multipart/form-data" action="https://` + r.Host + `/simpleup/" method="post">
   <input type="file" name="uploadfile" />
   <input type="hidden" name="token" value="{{.}}"/>
   <input type="submit" value="upload" />
  </form>
 </body>
</html>`
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
