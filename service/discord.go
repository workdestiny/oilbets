package service

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/acoshift/configfile"
	"github.com/workdestiny/oilbets/entity"
)

//SendErrorToDiscord send err
func SendErrorToDiscord(me *entity.Me, err string) {
	url := configfile.NewYAMLReader("config/config-application.yaml").String("webhook")
	if url == "" {
		return
	}

	dc := entity.Discord{
		AvatarURL: me.DisplayImage.Mini,
		Content:   err,
		Username:  me.FirstName + " " + me.LastName,
	}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(dc)
	http.Post(url, "application/json; charset=utf-8", b)
}
