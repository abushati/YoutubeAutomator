package PlayListSync

import (
	"../Youtube"
	"encoding/json"
	"fmt"
	"google.golang.org/api/youtube/v3"
	"io/ioutil"
	"time"
)

var config_file string = "PlayListSync/saved_playlist"
var client *youtube.Service = Youtube.YoutubeClient()

type Playlists struct {
	Playlists []Playlist `json:"PlaylistsInfo"`
}

type Playlist struct {
	Title      string  `json:"PlaylistsTitle"`
	PlaylistID string  `json:"PlaylistsId"`
	Videos     []Video `json:"All_Videos"`
}
type Video struct {
	Id               string
	Title            string
	Url              string
	NotificationData NotificationData
}
type NotificationData struct {
	NotificationSentforDay bool
	NoMoreNotification     bool
}

type playListCall struct {
	part       string
	playlistId string
	mine       bool
}
type playListItemCall struct {
	part      string
	id        string
	nextToken string
}

func (cfg Playlists) ContainsVideo(videoID string, playlistTitle string) bool {
	for _, i2 := range cfg.Playlists {
		if i2.Title == playlistTitle{
			for _, i4 := range i2.Videos {
				if i4.Id == videoID{
					return true
				}
			}
		}
	}
	return false
}

func callPlayListItemAPI(callParm playListItemCall) *youtube.PlaylistItemListResponse {
	call := client.PlaylistItems.List(callParm.part)
	call = call.PlaylistId(callParm.id)
	if callParm.nextToken != "" {
		call = call.PageToken(callParm.nextToken)
	}
	call = call.MaxResults(25)
	res, err := call.Do()
	if err != nil {
		fmt.Println("Error call")
	}
	return res
}

func callPlayListAPI(callParam playListCall) (*youtube.PlaylistListResponse, error) {
	call := client.Playlists.List(callParam.part)
	if callParam.playlistId != "" {
		call = call.Id(callParam.playlistId)
	}
	if callParam.mine == true {
		call = call.Mine(true)
	}
	call = call.MaxResults(25)
	resp, err := call.Do()
	if err == nil {
		return resp, err
	}
	return resp, err
}

func isIdSaved(idQuery string) bool {
	s := GetConfig()
	for _, pl := range s.Playlists {
		if pl.PlaylistID == idQuery {
			return true
		}
	}
	return false
}
func videoAlreadySaved(videoId string,i int, cfg Playlists) bool {
	for _, video := range  cfg.Playlists[i].Videos{
		if video.Id == videoId{
			return true
		}
	}
	return false
}

func MarkVideoWatched(videoID string)  {
	cfg := GetConfig()
	for pindex, playlist := range cfg.Playlists{
		for vindex,video:= range playlist.Videos{
			if video.Id == videoID{
				cfg.Playlists[pindex].Videos[vindex].NotificationData = NotificationData{
					NoMoreNotification:     true,
				}
			}
		}
	}
	SaveConfig(cfg)
}

func resetNotification()  {
	cfg := GetConfig()
	for pindex, playlist := range cfg.Playlists{
		for vindex, _ := range playlist.Videos{
			updatedNotificationData := NotificationData{
				NotificationSentforDay: false,
				NoMoreNotification: cfg.Playlists[pindex].Videos[vindex].NotificationData.NoMoreNotification,
			}
			cfg.Playlists[pindex].Videos[vindex].NotificationData = updatedNotificationData
		}
	}
	SaveConfig(cfg)
}

func savedVideos(video []Video, playListId string) {
	cfg := GetConfig()

	for i, playlist := range cfg.Playlists {
		if playlist.PlaylistID == playListId {
			videoList := playlist.Videos
			for _, video := range video {
				if cfg.ContainsVideo(video.Id,playlist.Title) {
					fmt.Println("Video is already saved")
					continue
					//videoList = append(videoList, video)
				} else {
					videoList = append(videoList, video)
				}
			}
			cfg.Playlists[i].Videos = videoList
			break
		}
	}
	SaveConfig(cfg)
}

func GetConfig() Playlists {
	var Playlists Playlists
	file, err := ioutil.ReadFile(config_file)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(file, &Playlists)
	return Playlists
}

func SaveConfig(cfg Playlists) {
	res, _ := json.MarshalIndent(cfg, "", "    ")
	ioutil.WriteFile(config_file, res, 0644)
}

//todo: use SaveConfig() to save changes done
func saveToConfig(newPlayList Playlist) {
	savedPlaylist := GetConfig()
	newPlaylists := append(savedPlaylist.Playlists, newPlayList)
	newPlaylist := Playlists{newPlaylists}
	res, _ := json.MarshalIndent(newPlaylist, "", "    ")
	ioutil.WriteFile(config_file, res, 0644)
}

func getVideosInPlayList(playListId string) []Video {
	nextPageToken := ""
	defaultNotification := NotificationData{
		NotificationSentforDay: false,
		NoMoreNotification:     false,
	}
	var videoList []Video
	for {
		callParm := playListItemCall{id: playListId, part: "contentDetails,snippet", nextToken: nextPageToken}
		res := callPlayListItemAPI(callParm)
		for _, video := range res.Items {
			newVideo := Video{Id: video.ContentDetails.VideoId, Url: video.Snippet.Thumbnails.High.Url,NotificationData: defaultNotification}
			videoList = append(videoList, newVideo)
		}
		if res.NextPageToken == "" {
			break
		}
		nextPageToken = res.NextPageToken
	}
	return videoList
}

func getAllMinePlayList() []Playlist {
	fields := playListCall{part: "id", mine: true}
	resp, err := callPlayListAPI(fields)
	if err != nil {
	}
	for _, r := range resp.Items {
		isSaved := isIdSaved(r.Id)
		if isSaved == true {
			fmt.Println(r.Id + " is saved locally")
			continue
		} else if !isSaved {
			fmt.Println(r.Id + " is being fetched via API")
			fields := playListCall{playlistId: r.Id, part: "snippet"}
			resp2, err := callPlayListAPI(fields)
			if err != nil {
				fmt.Println("bad")
			}
			items := resp2.Items[0]
			newPlayList := Playlist{
				Title:      items.Snippet.Title,
				PlaylistID: items.Id,
			}
			saveToConfig(newPlayList)
		}
	}
	allPlayList := GetConfig().Playlists
	return allPlayList
}

func RunSync() {
	allPlayList := getAllMinePlayList()
	fmt.Println(allPlayList)
	for _ , playlist := range allPlayList {
		if playlist.PlaylistID == "PLhbEyJUgbdml2wBZMPqWz75LEoQfS9qoK" {
			playListVideos := getVideosInPlayList(playlist.PlaylistID)
			savedVideos(playListVideos, playlist.PlaylistID)
			break
		}

	}
	hour,_,_ :=time.Now().Clock()
	fmt.Println(hour)
	if hour >= 22 && hour <=  24{
		resetNotification()
	}
}
