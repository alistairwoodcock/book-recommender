window.onload = app;

function app(){
	document.querySelector("#book-search").addEventListener("keyup", bookSearch);
	updateLikedBooks();
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
	sendBooklike(id, (res) => {
		if(res.Success){
			updateLikedBooks();	
		}
	})
}

function unlikeBook(id){
	sendBookUnlike(id, (res) => {
		if(res.Success){
			updateLikedBooks();	
		}
	})	
}

function updateLikedBooks(){
	fetchLikedBooks((res) => {
		app.state.likes = res.Books;
		renderLikedBooks();
		updateRecommendedBooks();
	});
}

function updateRecommendedBooks(){
	fetchRecommendations((res) => {
		app.state.recommended = res.Books;
		renderRecommendedBooks();
	})
}

function renderLikedBooks(){
	var list = document.getElementById("liked-books");

	var html = "";

	for (var i = 0; i < app.state.likes.length; i++) {
		html += renderBook(app.state.likes[i], 'unlikeBook');
	}

	list.innerHTML = html;
}

function renderRecommendedBooks(){
	var list = document.getElementById("recommended-books");

	var html = "";

	for (var i = 0; i < app.state.recommended.length; i++) {
		html += renderBook(app.state.recommended[i], 'likeBook');
	}

	list.innerHTML = html;
}



function bookSearch(event){
	app.state.search.term = event.target.value;
	
	fetchSearchResults((res) => {
		app.state.search.result = res.Books;
		renderBookSearch();
	});
}

function renderBookSearch(){
	var list = document.getElementById("search-result");
	
	var html = "";

	for (var i = 0; i < app.state.search.result.length && i < 100; i++) {
		html += renderBook(app.state.search.result[i], 'likeBook');
	}

	list.innerHTML = html;
}

function renderBook(book, onClickFuncName){
	return `
		<div class="book" onclick="${onClickFuncName}(${book.Id})">
			<div class="top">
				<img src="${book.Image}"/>
			</div>
			<div class="bottom">
				<p class="name">${book.Title}</p>
			</div>
		</div>
	`;
}


var fetchRecommendationsXHR = null;
function fetchRecommendations(func){
	if(fetchRecommendationsXHR != null){
		fetchRecommendationsXHR.abort();
		fetchRecommendationsXHR = null;
	}
	var xmlhttp = new XMLHttpRequest();   // new HttpRequest instance 
	xmlhttp.open("POST", "/api/recommend");
	xmlhttp.setRequestHeader("Content-Type", "application/json");
	xmlhttp.send(JSON.stringify({search: app.state.search.term}));
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

var fetchLikesXHR = null;
function fetchLikedBooks(func){
	if(fetchLikesXHR != null){
		fetchLikesXHR.abort();
		fetchLikesXHR = null;
	}
	var xmlhttp = new XMLHttpRequest();   // new HttpRequest instance 
	xmlhttp.open("GET", "/api/likes");
	xmlhttp.setRequestHeader("Content-Type", "application/json");
	xmlhttp.send(JSON.stringify({search: app.state.search.term}));
	fetchLikesXHR = xmlhttp;
	xmlhttp.onload = function(){
		
		if (xmlhttp.readyState == 4 && xmlhttp.status == "200") {
			var result = JSON.parse(xmlhttp.responseText);
			func(result);
		} else {
			console.log("error: %s", xmlhttp.responseText);
		}
		fetchLikesXHR = null;
	}
}

function sendBooklike(id, func){
	
	var xmlhttp = new XMLHttpRequest();   // new HttpRequest instance 
	xmlhttp.open("POST", "/api/like");
	xmlhttp.setRequestHeader("Content-Type", "application/json");
	xmlhttp.send(JSON.stringify({id: id}));
	fetchXHR = xmlhttp;
	xmlhttp.onload = function(){
		
		if (xmlhttp.readyState == 4 && xmlhttp.status == "200") {
			var result = JSON.parse(xmlhttp.responseText);
			func(result);
		} else {
			console.log("error: %s", xmlhttp.responseText);
		}
	}	
}

function sendBookUnlike(id, func){
	
	var xmlhttp = new XMLHttpRequest();   // new HttpRequest instance 
	xmlhttp.open("POST", "/api/unlike");
	xmlhttp.setRequestHeader("Content-Type", "application/json");
	xmlhttp.send(JSON.stringify({id: id}));
	fetchXHR = xmlhttp;
	xmlhttp.onload = function(){
		
		if (xmlhttp.readyState == 4 && xmlhttp.status == "200") {
			var result = JSON.parse(xmlhttp.responseText);
			func(result);
		} else {
			console.log("error: %s", xmlhttp.responseText);
		}
	}	
}