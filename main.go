package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"code.sajari.com/docconv"
)

var m = make(map[string]int)

func uploadFile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	if !strings.HasSuffix(handler.Filename, ".docx") {
		fmt.Fprintln(w, "Formato file non .docx")
	}
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	tempFile, err := ioutil.TempFile("temp-images", "upload-*.png")
	if err != nil {
		fmt.Println(err)
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	err = ioutil.WriteFile("uploaded.docx", fileBytes, 0666)
	if err != nil {
		log.Println(err)
	}
	// write this byte array to our temporary file
	tempFile.Write(fileBytes)
	tempFile.Close()
	f, err := os.Open("uploaded.docx")
	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	text, _, err := docconv.ConvertDocx(f)
	if err != nil {
		log.Println(err)
	}
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "’", " ")

	// fmt.Println(text)

	parole := strings.Split(text, " ")

	for _, parola := range parole {
		if parola == " " || parola == "" || parola == "-" {
			continue
		}
		parola = strings.TrimSpace(parola)
		parola = strings.Trim(parola, ",;.!?\n")
		parola = strings.ToLower(parola)
		m[parola]++
	}

	type kv struct {
		k string
		v int
	}

	var records [][]kv
	var element []kv
	for k, v := range m {
		element = append(element, kv{k, v})
	}
	records = append(records, element)
	sort.SliceStable(element, func(i, j int) bool {
		return element[i].v > element[j].v
	})
	wr := csv.NewWriter(w)

	wr.Comma = ';'

	// Scrive heaerds
	wr.Write([]string{"#Keyword", "Occorrenze"})
	for _, record := range element {
		if err := wr.Write([]string{record.k, strconv.Itoa(record.v)}); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}

	// Write any buffered data to the underlying writer (standard output).
	wr.Flush()

	if err := wr.Error(); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(w, nil)
	fmt.Fprintln(w, nil)

	// for k, v := range m {
	// 	fmt.Fprintln(w, k, v)
	// }
	// return that we have successfully uploaded our file!
	//fmt.Fprintf(w, "Successfully Uploaded File\n")
	//fmt.Fprint(w, text)
}

func index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./index.html")
	return
}

func setupRoutes() {
	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/", index)
	http.ListenAndServe(":8080", nil)
}

func main() {
	fmt.Println("Hello World")
	setupRoutes()
}
