//elephantsql.com

package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

//global db DB object
var db *sql.DB

type createData struct {
	Link string
}

func logFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func createLink(w http.ResponseWriter, r *http.Request) {
	/*
		-	1.
		-	2.
		-	3.
		-	4.
	*/
	fmt.Println("------------created------------")
	switch r.Method {
	case "GET":
		fmt.Println("CREATED GET")
		str := r.URL.RequestURI()
		fmt.Println(str)
		p, _ := r.URL.Parse(r.URL.String())
		u := p.Query()
		fmt.Println(r.Host)
		fmt.Println(u.Get("url"))
		// serve templated page with url!
		var link *string
		//ns, err := fmt.Sscanf(string(*link), "%/%s", r.Host, u.Get("url"))
		//fmt.Printf("n:err --%d:%s", n, err)
		fmt.Println("full link", link)
		fmt.Println("debug", r.Host, u.Get("url"))
		strLink := "http://" + r.Host + "/" + u.Get("url")
		fmt.Println(strLink)
		data := createData{strLink}
		tmpl := template.Must(template.ParseFiles("create.html"))
		fmt.Println(tmpl.Execute(w, data))
	case "PUT":
		fmt.Println("CREATED PUT")
	}

}

func addURL(w http.ResponseWriter, r *http.Request, url string) int {
	fmt.Println("------------addURL------------")
	queryString := fmt.Sprintf("insert into urls(longURL) values('%s') returning id;", url)
	fmt.Println(queryString)
	result, err := db.Query(queryString)

	//fmt.Println(err)
	var resultStr string
	resultStr = ""
	result.Next()
	err = result.Scan(&resultStr)

	fmt.Println(err, "-", resultStr)
	id, err := strconv.Atoi(resultStr)
	fmt.Print(err)
	return id
}

func updateURL(id int, shortURL, longURL string) {
	fmt.Println(longURL, "------------updateURL------------", shortURL)
	queryString := fmt.Sprintf("UPDATE urls SET shortURL = '%s' WHERE id = '%d';", shortURL, id)
	fmt.Println(queryString)
	result, _ := db.Query(queryString)
	var resultStr string
	result.Next()
	result.Scan(&resultStr)
	fmt.Println("AYY", resultStr)
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Println("------------home------------")
	if r.URL.Path != "/" {
		//check if text after / is id in db -- return redirect w url
		shortURL := r.URL.Path[1:]
		fmt.Println("url path", shortURL)
		id := shortURLtoID(shortURL)
		fmt.Println("id is:", id)
		//redirect page if id is in db, else page not found
		queryString := fmt.Sprintf("SELECT longURL FROM urls where id = %d", id)
		result, _ := db.Query(queryString)
		result.Next()
		var longURL string
		result.Scan(&longURL)
		fmt.Println(longURL)
		if len(longURL) == 0 {
			fmt.Println(longURL, "page not found...")
			//repace with url not found page
			http.Error(w, "404 not found.", http.StatusNotFound)
		} else {
			//fmt.Fprintf(w, longURL)
			//redirect
			http.Redirect(w, r, longURL, http.StatusSeeOther)
		}

		return
	}
	switch r.Method {
	case "GET":
		http.ServeFile(w, r, "home.html")
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		URL := r.FormValue("URL")

		//check url format
		_, err := url.ParseRequestURI(URL)
		if err != nil {
			//error occured
			fmt.Println("URL NOT VALID")
			//fmt.Fprintf(w, "\nURL not valid!")
			//render invalid url main page
			http.ServeFile(w, r, "homeInvalidURL.html")
			return
		}
		//fmt.Fprintf(w, "Post from website! r.PostFrom = %v\n", r.PostForm)
		//fmt.Fprintf(w, "URL : %s\n", URL)
		//add url to db.
		id := addURL(w, r, URL)
		//call url shorten functions
		shortURL := idToShortURL(id)
		//fmt.Fprintf(w, "localhost:8080/%s", shortURL)
		//call update
		updateURL(id, shortURL, URL)
		//redirect to created site url -- first create url to be parsed
		fmt.Println("-----URL formated for post redirect")
		redirURL := fmt.Sprintf("created/?%s=%s", "url", shortURL)
		http.Redirect(w, r, redirURL, http.StatusSeeOther)

	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func main() {
	fmt.Println("hello server")
	var err error
	db, err = sql.Open("postgres", "postgres://twdcnlmu:Hd7RXw1kL22RCi6Qbn0rldKHJMfGcSXp@hansken.db.elephantsql.com:5432/twdcnlmu")
	// postgres://twdcnlmu:Hd7RXw1kL22RCi6Qbn0rldKHJMfGcSXp@hansken.db.elephantsql.com:5432/twdcnlmu
	//Make sure you setup the ELEPHANTSQL_URL to be a uri, e.g. 'postgres://user:pass@host/db?options'
	fmt.Println(err)
	err = db.Ping()
	if err != nil {
		panic(err)
	} else {
		fmt.Println("db connected")
	}
	fs := http.FileServer(http.Dir("./static/"))

	//note USE OF GO FUNCTION. listenandserve blocks execution!
	go http.HandleFunc("/created/", createLink)
	go http.HandleFunc("/", home)
	http.Handle("/static/", http.StripPrefix("/static", fs)) // make this a go function?
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// url code below

func idToShortURL(id int) string {
	strmap := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	shortURL := ""
	for id > 0 {
		shortURL += string(strmap[id%62])
		id /= 62
	}
	//reverse url
	return reverse(shortURL)
}

func shortURLtoID(shortURL string) int {
	id := 0
	strmap := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for i, j := range reverse(shortURL) {
		id += strings.Index(strmap, string(j)) * int(math.Pow(float64(62), float64(i)))
	}
	fmt.Println()
	return id
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
