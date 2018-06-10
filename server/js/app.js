window.onload = app;

function app(){
	document.querySelector("#book-search").addEventListener("keyup", bookSearch);
	updateRecommendedBooks();
}

app.state = {
	search: {
		term: '',
		result: []
	},

	likes: [],

	recommended: [],
};

function likeBook(id){
	var found = app.state.likes.find(book => book.Id == id);

	if(found != undefined) return;
	
	fetchBookDetails(id, (res) => {
		if(res.Success){
			app.state.likes.push(res.Book);
			renderLikedBooks();
			updateRecommendedBooks();
		}
	});
}

function unlikeBook(id){
	var newLikes = app.state.likes.filter(book => book.Id != id);

	app.state.likes = newLikes;

	renderLikedBooks();
	updateRecommendedBooks();
}

function updateRecommendedBooks(){

	likedBookIds = app.state.likes.map(book => book.Id);

	fetchRecommendations(likedBookIds, (res) => {
		app.state.recommended = res.Books;
		renderRecommendedBooks();
		renderInfo();
	})
}

function renderLikedBooks(){
	var likedTitle = document.querySelector(".books-you-like .top");
	var list = document.getElementById("liked-books");

	var html = "";

	for (var i = 0; i < app.state.likes.length; i++) {
		html += renderBook(app.state.likes[i], 'unlikeBook');
	}

	list.innerHTML = html;

	if(app.state.likes.length == 0){
		likedTitle.classList.add('hidden');
	} else {
		likedTitle.classList.remove('hidden');
	}
}

function renderRecommendedBooks(){
	var recommendTitle = document.querySelector(".recommendations .top");
	var list = document.getElementById("recommended-books");

	var html = "";

	for (var i = 0; i < app.state.recommended.length; i++) {
		html += renderBook(app.state.recommended[i], 'likeBook');
	}

	list.innerHTML = html;

	if(app.state.recommended.length == 0){
		recommendTitle.classList.add('hidden');
	} else {
		recommendTitle.classList.remove('hidden');
	}
}



function bookSearch(event){
	app.state.search.term = event.target.value;
	
	fetchSearchResults((res) => {
		app.state.search.result = res.Books;
		renderBookSearch();
		renderInfo();
	});
}

function renderBookSearch(){
	var areaTitle = document.querySelector("#search-area .top");
	var list = document.getElementById("search-result");
	
	var html = "";

	for (var i = 0; i < app.state.search.result.length && i < 100; i++) {
		html += renderBook(app.state.search.result[i], 'likeBook');
	}

	list.innerHTML = html;

	if(app.state.search.result.length == 0){
		areaTitle.classList.add('hidden');
	} else {
		areaTitle.classList.remove('hidden');
	}
}

function renderInfo(){
  var infoEl = document.querySelector("#project-info");
  if(app.state.search.result.length == 0 && 
     app.state.likes.length == 0 &&
     app.state.recommended.length == 0)
  {
  	infoEl.classList.remove('hidden');
  } else {
  	infoEl.classList.add('hidden');
  }
}

function renderBook(book, onClickFuncName){
	return `
		<div class="book" onclick="${onClickFuncName}(${book.Id})">
			<div class="top">
				<img src="${book.Image}" alt="${book.Title}"/>
			</div>
			<div class="bottom">
				<p class="name">${book.Title}</p>
			</div>
		</div>
	`;
}




var fetchRecommendationsXHR = null;
function fetchRecommendations(likedBookIds, func){
	if(fetchRecommendationsXHR != null){
		fetchRecommendationsXHR.abort();
		fetchRecommendationsXHR = null;
	}
	var xmlhttp = new XMLHttpRequest();   // new HttpRequest instance 
	xmlhttp.open("POST", "/api/recommend");
	xmlhttp.setRequestHeader("Content-Type", "application/json");
	xmlhttp.send(JSON.stringify({LikedBookIds: likedBookIds}));
	fetchRecommendationsXHR = xmlhttp;
	
	xmlhttp.onload = function(){
		
		if (xmlhttp.readyState == 4 && xmlhttp.status == "200") {
			var result = JSON.parse(xmlhttp.responseText);
			func(result);
		} else {
			console.log("error: %s", xmlhttp.responseText);
		}
		fetchRecommendationsXHR = null;
	}	
}


var fetchXHR = null;
function fetchSearchResults(func){
	if(fetchXHR != null){
		fetchXHR.abort();
		fetchXHR = null;
	}
	var xmlhttp = new XMLHttpRequest();   // new HttpRequest instance 
	xmlhttp.open("POST", "/api/search");
	xmlhttp.setRequestHeader("Content-Type", "application/json");
	xmlhttp.send(JSON.stringify({search: app.state.search.term}));
	fetchXHR = xmlhttp;
	xmlhttp.onload = function(){
		
		if (xmlhttp.readyState == 4 && xmlhttp.status == "200") {
			var result = JSON.parse(xmlhttp.responseText);
			func(result);
		} else {
			console.log("error: %s", xmlhttp.responseText);
		}
		fetchXHR = null;
	}
}

var bookDetailsXHR = null;
function fetchBookDetails(id, func) {
	if(bookDetailsXHR != null){
		bookDetailsXHR.abort();
		bookDetailsXHR = null;
	}
	var xmlhttp = new XMLHttpRequest();   // new HttpRequest instance 
	xmlhttp.open("POST", "/api/book");
	xmlhttp.setRequestHeader("Content-Type", "application/json");
	xmlhttp.send(JSON.stringify({Id: id}));
	bookDetailsXHR = xmlhttp;
	xmlhttp.onload = function(){
		
		if (xmlhttp.readyState == 4 && xmlhttp.status == "200") {
			var result = JSON.parse(xmlhttp.responseText);
			func(result);
		} else {
			console.log("error: %s", xmlhttp.responseText);
		}
		bookDetailsXHR = null;
	}
}