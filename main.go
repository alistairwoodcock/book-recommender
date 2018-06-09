package main;

func checkError(err error){
	if(err != nil){
		print("Error!");
		panic(err);	
	}
}

var databasePath = "./data/data2.db";


func main(){
	server();
}