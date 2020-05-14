package Notifier

import (
	"../PlayListSync"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/smtp"
)

const playListFile = "/PlayListSync/saved_playlist"

type creds struct {
	Username string
	Password string
}

type videoTemplate struct {
	URL string
	LinktoThumb string
}

func getCreds() (string,string)  {
	var cred creds
	file, _ := ioutil.ReadFile("PlayListSync/creds")
	json.Unmarshal(file,&cred)
	fmt.Println(cred.Password)
	return cred.Username, cred.Password
}

func sendEmail(body string)  {
	username, password := getCreds()
	from := username
	pass := password
	to := "arvid.b901@gmail.com"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n";
	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: Hello there\r\n" +
		mime + "\r\n\r\n" +
		body

	err := smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		fmt.Printf("smtp error: %s", err)
		return
	}

	fmt.Print("sent, visit email")
}

func generateEmail(videos []PlayListSync.Video) string {
	var emailBody []videoTemplate
	body := "<html><body><h1>Video Reminder</h1>"
	for _,video := range videos {
		v := videoTemplate{
			URL:        video.Id,
			LinktoThumb: video.Url,
		}
		emailBody = append(emailBody,v)
	}

	for _,videolink := range emailBody{
		videoId := videolink.URL
		url:= "http://localhost:8080/video/"+videoId
		body = body+
			"<img src=\""+videolink.LinktoThumb + "\">\n" +
			"<div>"+url+"</div>"
	}
	body = body+"</body></html>"

	fmt.Println(body)
	return body
}


func getPlayList() []PlayListSync.Playlist{
	cfg := PlayListSync.GetConfig()
	return cfg.Playlists
}

func getVideosToNotify(playlist PlayListSync.Playlist) []PlayListSync.Video{
	var returnVidos []PlayListSync.Video
	for _,video := range playlist.Videos{
		notificationData := video.NotificationData
		if notificationData.NoMoreNotification == false && notificationData.NotificationSentforDay == false{
			returnVidos = append(returnVidos,video)

		}
	}
	return returnVidos
}

func updatedNotificationData(playlist string,videosToNotify []PlayListSync.Video){
	cfg := PlayListSync.GetConfig()
	for _, videoToNotify := range videosToNotify{
		for playlistIndex, playList := range cfg.Playlists{
			if playList.PlaylistID == playlist{
				for videoIndex, video := range playList.Videos{
					if video == videoToNotify{
						newNotification := PlayListSync.NotificationData{
							true,
							video.NotificationData.NoMoreNotification,
						}
						fmt.Println(newNotification)
						cfg.Playlists[playlistIndex].Videos[videoIndex].NotificationData = newNotification
						break
					}
				}
			}
		}
	}
	PlayListSync.SaveConfig(cfg)
}

func NN() {
	PlayListSync.RunSync()
	allPlayList := getPlayList()
	for _,playlist := range allPlayList{
		if playlist.Title == "Watch Later"{
			videosToNotify := getVideosToNotify(playlist)
			if len(videosToNotify) > 0{
				updatedNotificationData(playlist.PlaylistID,videosToNotify)
				body := generateEmail(videosToNotify)
				sendEmail(body)
			}
		}
	}
}