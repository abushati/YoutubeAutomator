package main

import (
	"./PlayListSync"
	//"./Notifier"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func markVideoWatched(videoID string)  {
	PlayListSync.MarkVideoWatched(videoID)
}


func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("here")
	v := strings.Split(r.URL.Path,"/")[2]
	fmt.Println(v)
	go markVideoWatched(v)
	//todo: return error is v is out of range
	const youtubeBase = "https://www.youtube.com/watch?v="
	//fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
	http.Redirect(w, r, youtubeBase+v, http.StatusSeeOther)
}

func main() {
	//Notifier.NN()
	http.HandleFunc("/video/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
