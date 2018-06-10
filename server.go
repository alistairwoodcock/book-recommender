package main

import (
	"log"
	"mime"
	"fmt"
	"time"
	
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
		LIMIT 14
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


func bookDetailsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json");

	fmt.Println(r.Body);

	decoder := json.NewDecoder(r.Body)
	encoder := json.NewEncoder(w);

	ret := new(struct {
		Success bool
		Book Book
	});
	ret.Success = false;

	defer encoder.Encode(ret);
	
	req := struct {
		Id int
	}{
		-1,
	}
	
	err := decoder.Decode(&req);

	if(err != nil){
		fmt.Println("decode failed");
		return;
	}

	fmt.Println("looking for book with id: ", req.Id);

	if(req.Id < 0){
		print("Negative id given");
		return;
	}

	bookRow, err := db.Query(`SELECT book_id, title, image_url FROM books WHERE book_id = ?`, req.Id);
	if(err != nil){
		print("couldn't find book with given book_id");
		return;
	}

	defer bookRow.Close();

	if(bookRow.Next()){
		book := Book{};
		err := bookRow.Scan(&book.Id, &book.Title, &book.Image);

		if(err != nil){
			print("Scanning book failed");
			return;
		}

		ret.Book = book;
		ret.Success = true;

	} else {
		print("no book found");
		return;
	}
}

func recommendHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json");

	decoder := json.NewDecoder(r.Body)
	
	req := struct {
		LikedBookIds []int
	}{
		nil,
	}

	err := decoder.Decode(&req);

	if(err != nil){
		print(err);
		return;
	}

	ret := new(struct {
		Books []Book
		Updating bool
	});
	ret.Books = make([]Book, 0);
	ret.Updating = false;

	defer json.NewEncoder(w).Encode(ret);

	start := time.Now()
	fmt.Println("generating recommendation for ", 16, " books");
	bookIDs := generateRecommendation(req.LikedBookIds, 16);
	fmt.Println("finished recommendation");
	fmt.Println("took ", time.Since(start));

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
	
	setupUserRecommendData();

	r := mux.NewRouter()

	r.HandleFunc("/css/{filename}", CSSFileServer)
	r.HandleFunc("/js/{filename}", JSFileServer)


	r.HandleFunc("/", homePageHandler);
	r.HandleFunc("/api/search", searchHandler);
	r.HandleFunc("/api/book", bookDetailsHandler);
	r.HandleFunc("/api/recommend", recommendHandler);

	http.Handle("/", r)

	print("running on port 3003\n");

	http.ListenAndServe(":3003", nil)
}



