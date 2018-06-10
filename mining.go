package main

import (
	"fmt"
	"math"
	_ "github.com/mattn/go-sqlite3"
	"github.com/emirpasic/gods/maps/hashmap"
)


type User struct {
	Id int;
	Books *hashmap.Map;
	WantToRead *hashmap.Map;
	kNNUser []User;
	kNNDist []float64;
};

var users []User;
var averageRatingByBookId [10001]float64;

func setupUserRecommendData() {

	toRead := make([][]int, 54000);

	toReads, err := db.Query(`select book_id, user_id from to_read where 1`);
	checkError(err);

	for toReads.Next() {
		var BookId int;
		var UserId int;

		err = toReads.Scan(&BookId, &UserId);
		checkError(err);

		toRead[UserId] = append(toRead[UserId], BookId);
	}

	toReads.Close()


	userRows, err := db.Query(`SELECT DISTINCT user_id FROM ratings WHERE 1`);
	checkError(err);

	for userRows.Next() {
		var user User;
		user.Books = hashmap.New();
		user.WantToRead = hashmap.New();

		err = userRows.Scan(&user.Id);
		checkError(err);

		booksToRead := toRead[user.Id];

		for _, BookId := range booksToRead {
			user.WantToRead.Put(BookId, true);
			user.Books.Put(BookId, true);
		}

		users = append(users, user);
	}

	fmt.Println("user count: ", len(users));

	userRows.Close();

	avgRows, err := db.Query(`SELECT book_id, average_rating FROM books WHERE 1`);
	checkError(err);

	for avgRows.Next() {
		var BookId int;
		var Rating float64;

		err = avgRows.Scan(&BookId, &Rating);
		checkError(err);

		averageRatingByBookId[BookId] = Rating;
	}

	avgRows.Close();


}

func generateRecommendation(numberOfBooks int) []int{
	var recommendations []int; //return array
	
	//Active User
	user1 := User{-1, nil, nil, nil, nil};
	user1.Id = -1;
	user1.WantToRead = hashmap.New();
	defer user1.WantToRead.Clear();

	likedBooksRow, err := db.Query(`select book_id from user_likes where 1`);
	checkError(err);

	for likedBooksRow.Next() {
		var BookId int;
		err := likedBooksRow.Scan(&BookId);
		checkError(err);

		user1.WantToRead.Put(BookId, true); // user1.WantToRead = append(user1.WantToRead, BookId);
	}

	likedBooksRow.Close();

	fmt.Println("user likes: ", user1.WantToRead.Size());

	N := 20;

	for i := 0; i < N; i++ {
		emptyUser := User{-1, nil, nil, nil, nil};
		user1.kNNUser = append(user1.kNNUser, emptyUser);
		user1.kNNDist = append(user1.kNNDist, -1);
	}
	
	user1BookCount := user1.WantToRead.Size();
	if(user1BookCount == 0){
		return recommendations;
	}

	for _, user2 := range users {
		user2BookCount := user2.Books.Size();
		intersectingBookCount := 0;
		
		for _, bookId := range user1.WantToRead.Keys() {
			_, wantsToRead := user2.Books.Get(bookId);
			if(wantsToRead){
				intersectingBookCount++;
			}
		}

		user2ToReadCount := user2.WantToRead.Size();
		user1ToReadCount := user1.WantToRead.Size();
		intersectingToReadCount := 0;

		for _, bookId := range user1.WantToRead.Keys() {
			_, wantsToRead := user2.WantToRead.Get(bookId);
			if(wantsToRead){
				intersectingToReadCount++;
			}
		}

		BooksJaccardIndex := float64(intersectingBookCount) / float64((user1BookCount + user2BookCount) - intersectingBookCount);
		ToReadJaccardIndex := float64(intersectingToReadCount) / float64((user1ToReadCount + user2ToReadCount) - intersectingToReadCount);

		similarity := ToReadJaccardIndex + BooksJaccardIndex;

		if(math.IsNaN(similarity)){
			fmt.Println("NaN");
			continue;
		}
		
		for i := 0; i < N; i++ {
			if(user1.kNNUser[i].Id == -1 || similarity > user1.kNNDist[i]){
				user1.kNNUser[i] = user2;
				user1.kNNDist[i] = similarity;
				break;
			}
		}
	}

	var bestRatings []float64;
	var bestRatingBooks []int;
	var maxBestRatings = numberOfBooks;

	for i:=0;i<maxBestRatings;i++{
		bestRatings = append(bestRatings, -1);
		bestRatingBooks = append(bestRatingBooks, -1);
	}


	var recBookLen int = 0;
	var recBooks []int;
	var recBookRatingSum []float64;
	var recBookRatingCount []float64;


	ratings := hashmap.New(); //use this for ratings as we pull them from each closest user

	for i, user := range user1.kNNUser {

		ratingsRow, err := db.Query(`select rating, book_id from ratings where user_id = ?`, user.Id);
		checkError(err);

		for ratingsRow.Next() {
			var bookId int;
			var rating float64;

			err = ratingsRow.Scan(&rating, &bookId);
			checkError(err);

			ratings.Put(bookId, rating);
		}

		ratingsRow.Close();

		for _, id := range user.Books.Keys() {
			BookId := id.(int);
			
			skipBook := false;

			for _, userbook := range user1.WantToRead.Keys() {
				if(BookId == userbook){
					skipBook = true;
					break;
				}
			}

			if(skipBook){
				break;
			}

			recIndex := -1;

			for i:=0; i < recBookLen; i++ {
				if(recBooks[i] == BookId){
					recIndex = i;
					break;
				}
			}

			if(recIndex == -1){

				//check it doesn't exist in users liked list
				recBooks = append(recBooks, BookId);
				recBookRatingSum = append(recBookRatingSum, 0);
				recBookRatingCount = append(recBookRatingCount, 0);
				recIndex = len(recBooks)-1;	
			}

			dist := user1.kNNDist[i];
			r, hasRating := ratings.Get(BookId);

			if(!hasRating){
				continue;
			}

			rating := r.(float64);

			recBookRatingSum[recIndex] += rating * dist;
			recBookRatingCount[recIndex] += dist;
		}

		ratings.Clear();
	}

	for a:=0; a < len(recBooks); a++ {

		BookId := recBooks[a];
		prediction := (recBookRatingSum[a] / recBookRatingCount[a]);
		
		for i:=0; i < maxBestRatings; i++ {

			if(bestRatings[i] == -1 || prediction > bestRatings[i]){
				alreadyRecommended := false;

				for j:=0;j<maxBestRatings;j++{
					if(bestRatingBooks[j] == BookId){
						alreadyRecommended = true;
						break;
					}
				}

				if(!alreadyRecommended){
					bestRatings[i] = prediction;
					bestRatingBooks[i] = BookId;
				}
				break;
			}
		}
	}

	for i, bookId := range bestRatingBooks {
		if(bookId == -1 || bestRatings[i] == -1){
			continue;
		}

		recommendations = append(recommendations, bookId);
	}

	

	return recommendations;
}
	