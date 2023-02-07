package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-mastodon"
	crontab "github.com/robfig/cron/v3"
	"github.com/samber/lo"
	"golang.org/x/text/width"
)

var mstdn *mastodon.Client

func init() {
	rand.Seed(time.Now().UnixMicro())

	clientID := os.Getenv("MASTODON_CLIENT_ID")
	clientSecret := os.Getenv("MASTODON_CLIENT_SECRET")
	accessToken := os.Getenv("MASTODON_ACCESS_TOKEN")
	if clientID == "" || clientSecret == "" || accessToken == "" {
		panic("MASTODON_CLIENT_ID, MASTODON_CLIENT_SECRET, and MASTODON_ACCESS_TOKEN must be set")
	}

	mstdn = mastodon.NewClient(&mastodon.Config{
		Server:       "https://social.vivaldi.net",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AccessToken:  accessToken,
	})
}

func main() {
	cron := crontab.New()
	lo.Must(cron.AddFunc("0 * * * *", updateIcon))
	cron.Start()
	go streamUser()
	log.Println("Wasa wasa")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("wasawasa"))
	})
	http.ListenAndServe("0.0.0.0:80", nil)
}

func streamUser() {
	delay := time.Second
	for {
		ch, err := mstdn.StreamingUser(context.Background())
		if err != nil {
			log.Println(err)
			log.Println("Reconnect:", delay)
			time.Sleep(delay)
			delay *= 2
			continue
		}
		log.Println("Started streaming!!")
		delay = time.Second
		for e := range ch {
			switch e := e.(type) {
			case *mastodon.UpdateEvent:
				onUpdate(e.Status)
			case *mastodon.NotificationEvent:
				onNotification(e.Notification)
			}
		}
		log.Println("Closed streaming. Restarting...")
		time.Sleep(delay)
	}
}

func onUpdate(status *mastodon.Status) {
	if status.Account.StatusesCount%686 == 0 {
		count := status.Account.StatusesCount
		killme := count / 686
		flavor := ""
		if killme == 686 {
			flavor = "あなたがカヅホ神ですか！？"
		} else if killme%10 > 0 {
			flavor = ohome[rand.Intn(len(ohome))]
		} else {
			flavor = majires[rand.Intn(len(majires))]
		}
		text := fmt.Sprintf(
			"@%v さんが、投稿数: %v\n(%vキルミー) に到達したよ！！\n%v",
			status.Account.Username,
			count,
			killme,
			flavor,
		)
		toot := &mastodon.Toot{
			Status:     text,
			Visibility: "unlisted",
			Language:   "ja",
		}
		if status.Reblog != nil {
			toot.InReplyToID = status.ID
		}
		content, err := mstdn.PostStatus(context.Background(), toot)
		if err != nil {
			log.Println(err)
		}
		log.Println("Tooted:", content.ID)
	}
	{ // Favorite
		content := strings.ToLower(width.Fold.String(status.Content))
		for _, fav := range favorites {
			if strings.Contains(content, fav) {
				log.Println("Favorited:", status.ID)
				go mstdn.Favourite(context.Background(), status.ID)
				break
			}
		}
	}
	onUpdatePopular(status)
}

func onNotification(notif *mastodon.Notification) {
	switch notif.Type {
	case "follow":
		log.Println("Followed by:", notif.Account.Acct)
		go mstdn.AccountFollow(context.Background(), notif.Account.ID)
	}
}

func updateIcon() {
	log.Println("Updating icon...")
	name := i686[rand.Intn(len(i686))]
	url := fmt.Sprintf("http://aka.saintpillia.com/killme/icon/%v.png", name)
	log.Println("URL:", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	buf.ReadFrom(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Downloaded.")

	body := new(bytes.Buffer)
	mp := multipart.NewWriter(body)
	w, err := mp.CreateFormFile("avatar", "avatar.png")
	if err != nil {
		log.Println(err)
		return
	}
	buf.WriteTo(w)
	mp.Close()

	req, _ := http.NewRequest("PATCH", "https://social.vivaldi.net/api/v1/accounts/update_credentials", body)
	req.Header.Set("Content-Type", mp.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+os.Getenv("MASTODON_ACCESS_TOKEN"))

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	if 200 <= resp.StatusCode && resp.StatusCode < 300 {
		log.Println("Update icon done.")
		return
	}
	log.Println("Update icon failed.")
	log.Println("Status:", resp.Status)
	bd, _ := io.ReadAll(resp.Body)
	log.Println("Body:", string(bd))
}
