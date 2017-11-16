package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Exists struct {
	Exists bool
}


func SimpleUp(w http.ResponseWriter, r *http.Request) {

	_, err := httputil.DumpRequest(r, false)
	overwrite := r.URL.Query().Get("overwrite")

	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

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
		log.Println("Username " + username + " attempting an upload...")
		if _, err := os.Stat(UploadDIR + "/" + username); os.IsNotExist(err) {
			log.Println("Directory /uploads/" + username + " does not exist so making it...")
			os.Mkdir(UploadDIR+"/"+username, 0777)

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
		defer file.Close()
                log.Println(file)
		ID := r.FormValue("dsetid")
		if ID != "" {
			if _, err := os.Stat(UploadDIR + "/" + username + "/" + handler.Filename); os.IsNotExist(err) {

				log.Println("File is  " + handler.Filename)
				if handler.Filename == "" {
					log.Println(handler.Filename + "is empty!!!. This is a problem!")
					w.WriteHeader(http.StatusNoContent)
					return
				}
				f, err := os.OpenFile(UploadDIR+"/"+username+"/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0777)
				if err != nil {
					fmt.Println(err)
					return
				}
				defer f.Close()
				_, err = io.Copy(f, file)
					if err != nil{
						log.Println("Copy to filesystem failed:\n")
						log.Println(err)
					}
				extension := filepath.Ext(handler.Filename)
				extension = "*" + extension
				stmt, err := formdb.Prepare("UPDATE datasets SET filename=?, filetype=? WHERE id=?")
				checkErr(err)
				res, err := stmt.Exec(handler.Filename, extension, ID)
				fmt.Println(res)
				checkErr(err)
				affect, rowerr := res.RowsAffected()
				log.Printf("ID = %d, affected = %d\n", ID, affect)
				fmt.Println("closing file")
                                file.Close()
				if _, err := os.Stat(UploadDIR + "/" + username + "/" + handler.Filename); os.IsNotExist(err) {
					w.WriteHeader(http.StatusInternalServerError)
					return

				} else if rowerr != nil {
					log.Fatal(rowerr)
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					w.WriteHeader(http.StatusCreated)
					return
				}
			} else if overwrite == "true" {
				secs := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
				oldpath := UploadDIR + "/" + username + "/" + handler.Filename
				newpath := oldpath + "." + secs
				os.Rename(oldpath, newpath)

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
				extension := filepath.Ext(handler.Filename)
				extension = "*" + extension
				stmt, err := formdb.Prepare("UPDATE datasets SET filename=?, filetype=? WHERE id=?")
				checkErr(err)
				res, err := stmt.Exec(handler.Filename, extension, ID)
				fmt.Println(res)
				checkErr(err)
				affect, rowerr := res.RowsAffected()
				log.Printf("ID = %d, affected = %d\n", ID, affect)
                               	fmt.Println("closing file")
                                file.Close()

				if _, err := os.Stat(UploadDIR + "/" + username + "/" + handler.Filename); os.IsNotExist(err) {
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else if rowerr != nil {
					log.Fatal(rowerr)
					w.WriteHeader(http.StatusInternalServerError)
					return
				} else {
					w.WriteHeader(http.StatusCreated)
					return
				}

			} else {

				w.WriteHeader(http.StatusFound)
				return
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	fmt.Println("file should close now")
	} else if username == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}
	return

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
  <form enctype="multipart/form-data" action="https://` + r.Host + `/simpleup/?overwrite=true" method="post">
   <input type="file" name="file" />
   <input type="submit" value="upload" />
   ID: <input type="text" name="dsetid"><br>
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

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func FileExists(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	if filename != "" {
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

			if _, err := os.Stat(UploadDIR + "/" + username + "/" + filename); os.IsNotExist(err) {
				exists := Exists{false}
				js, err := json.Marshal(exists)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.Write(js)
			} else {
				exists := Exists{true}
				js, err := json.Marshal(exists)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.Write(js)
			}
		} else {
			w.Write([]byte("You must be logged in to use this service"))
		}
	} else {
		w.Write([]byte("must supply a filename example: simpleup/exists?filename=abc.txt"))
	}

}
