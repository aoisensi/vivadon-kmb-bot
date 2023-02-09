package main

import (
	"context"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/mattn/go-mastodon"
)

// :tony_astonished: :tony_grinning: :tony_happy: :tony_laughing: :tony_neutral: :tony_normal: :tony_santa: :tony_side: :tony_sigh: :tony_smiling: :tony_smirking: :tony_unhappy: :tony_wee:
var popularPower = map[string]time.Time{
	":tony_astonished:": {},
	":tony_grinning:":   {},
	":tony_happy:":      {},
	":tony_laughing:":   {},
	":tony_neutral:":    {},
	":tony_normal:":     {},
	":tony_santa:":      {},
	":tony_side:":       {},
	":tony_sigh:":       {},
	":tony_smiling:":    {},
	":tony_smirking:":   {},
	":tony_unhappy:":    {},
	":tony_wee:":        {},
}

var popularCooldown = time.Minute * 5
var popularBoost = time.Second * 15
var popularThreshold = time.Second * 30

var popularLaunchedAt time.Time

func onUpdatePopular(status *mastodon.Status) {
	if popularLaunchedAt.Add(popularCooldown).After(time.Now()) {
		return
	}
	content := status.Content
	content = strings.TrimLeft(content, "<p>")
	content = strings.TrimRight(content, "</p>")
	if t, ok := popularPower[content]; ok {
		log.Println("Popular: ", content)
		if t.Before(time.Now()) {
			t = time.Now()
		}
		t = t.Add(popularBoost)
		popularPower[content] = t
		if time.Now().Add(popularThreshold).Before(t) {
			popularLaunchedAt = time.Now()
			go func() {
				time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
				status, err := mstdn.PostStatus(context.Background(), &mastodon.Toot{
					Status:   content,
					Language: "ja",
				})
				if err != nil {
					log.Println(err)
				}
				log.Println("Tooted: ", status.ID)
			}()
		}
	}
}
