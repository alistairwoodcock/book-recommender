package main

import (
	"log"
	"mime"
	"fmt"
	
	"database/sql"
	_ "github.com/mattn/go-sqlite3"

	"net/http"
	"io/ioutil"

	"encoding/json"
	"path/filepath"

	"github.com/gorilla/mux"
)

type Book struct{
	Id int;
	Title string;
	Image string;
};

var db *sql.DB;

func loadPage(name string) (string, error) {
	filename := "server/pages/" + name + ".html"
	body, err := ioutil.ReadFile(filename)

	if err != nil {
		body := "no file found"
		return body, err
	}

	return string(body), nil
}

func show404(w http.ResponseWriter, r *http.Request){
	page, err := loadPage("404")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	fmt.Fprintf(w, page)
}

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	page, err := loadPage("index")

	if err != nil {
		show404(w, r);
		return;
	}

	fmt.Fprintf(w, page);
}

func ServeFileInPath(filePath string, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
    file := vars["filename"]

    print(file+"\n");

    // Super simple. Doesn't set any cache headers, check existence, avoid race conditions, etc.
    w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(file)))
    http.ServeFile(w, r, filePath+file)
}

func CSSFileServer(w http.ResponseWriter, r *http.Request) {
    ServeFileInPath("./server/css/", w, r);
}

func JSFileServer(w http.ResponseWriter, r *http.Request) {
    ServeFileInPath("./server/js/", w, r);
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w);
	defer r.Body.Close()

	var err error;

	ret := new(struct {Books []Book });
	ret.Books = make([]Book,0);

	defer encoder.Encode(ret);

	post := struct {
		Search string
	}{
		"",
	};

	err = decoder.Decode(&post)
	if err != nil {
        return;
    }

    rows, err := db.Query(`
		SELECT books.book_id, title, image_url 
		FROM books
		WHERE LOWER(title) LIKE '%' || ? || '%'
		OR LOWER(authors) LIKE '%' || ? || '%'
		ORDER BY average_rating
	`, post.Search, post.Search);

    if err != nil {
		log.Fatal(err)
		return;
	}

	defer rows.Close()

	if(post.Search != "") {
		for rows.Next() {
			book := Book{};
			
			err = rows.Scan(&book.Id, &book.Title, &book.Image);
			checkError(err);
			if err != nil {
				log.Fatal(err);
			}
			ret.Books = append(ret.Books, book);
		}	
	}
}


func likesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	var err error;

	ret := new(struct {Books []Book});
	ret.Books = make([]Book, 0);

	defer json.NewEncoder(w).Encode(ret);

	rows, err := db.Query(`
		SELECT books.book_id, title, image_url 
		FROM books
		INNER JOIN user_likes ON books.book_id = user_likes.book_id
	`);

    if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {
		book := Book{};
		
		err = rows.Scan(&book.Id, &book.Title, &book.Image);
		if err != nil {
			log.Fatal(err);
		}

		ret.Books = append(ret.Books, book);
	}	
}

func likeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json");

	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w);


	req := struct {
		Id int
	}{
		-1,
	}

	ret := new(struct {Success bool});
	ret.Success = false;

	defer encoder.Encode(ret);
	
	
	err := decoder.Decode(&req);

	if(err != nil){
		return;
	}

	stmt, err := db.Prepare(`INSERT INTO user_likes (book_id) VALUES (?); `);
    
	if(err != nil){
		return;
	}

	if(req.Id < 0){
		return;
	}

	bookRow, err := db.Query(`SELECT book_id FROM books WHERE book_id = ?`, req.Id);

	if(err != nil){
		return;
	}

	if(!bookRow.Next()){
		return;
	}

	_, err = stmt.Exec(req.Id)


    if(err != nil){
    	print(err);
    	return;
    } else {
    	ret.Success = true;
    }
}

func unlikeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json");

	req := struct {
		Id int
	}{
		-1,
	}

	ret := new(struct {Success bool});
	ret.Success = false;
	
	defer json.NewEncoder(w).Encode(ret);

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req);

	if(err != nil){
		return;
	}

	stmt, err := db.Prepare(`DELETE FROM user_likes WHERE book_id = ?`)

	stmt.Exec(req.Id);

	ret.Success = true;
}


func recommendHandler(w http.ResponseWriter, r *http.Request) {

	ret := new(struct {
		Books []Book
		Updating bool
	});
	ret.Books = make([]Book, 0);
	ret.Updating = false;

	defer json.NewEncoder(w).Encode(ret);

	bookIDs := generateRecommendation(16);

	for _, bookID := range bookIDs {

		row, err := db.Query(`
			SELECT book_id, title, image_url 
			FROM books
			WHERE book_id = ?
		`, bookID);

		if(row.Next()){
			book := Book{};
		
			err = row.Scan(&book.Id, &book.Title, &book.Image);

			if err != nil {
				log.Fatal(err);
			}

			ret.Books = append(ret.Books, book);
		}

		row.Close();
	}
}

func server(){
	var err error;
	db, err = sql.Open("sqlite3", databasePath);
	if err != nil {
		log.Fatal(err);
	}

	setupUserRecommendData();

	r := mux.NewRouter()

	r.HandleFunc("/css/{filename}", CSSFileServer)
	r.HandleFunc("/js/{filename}", JSFileServer)


	r.HandleFunc("/", homePageHandler);
	r.HandleFunc("/api/search", searchHandler);
	r.HandleFunc("/api/likes", likesHandler);
	r.HandleFunc("/api/like", likeHandler);
	r.HandleFunc("/api/unlike", unlikeHandler);
	r.HandleFunc("/api/recommend", recommendHandler);

	http.Handle("/", r)

	print("running on port 3003\n");

	http.ListenAndServe(":3003", nil)
}



