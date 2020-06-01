package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

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

//SendWithdrawToDiscord send withdraw
func SendWithdrawToDiscord(me *entity.Me, number int64) {
	url := configfile.NewYAMLReader("config/config-application.yaml").String("webhook_withdraw")
	if url == "" {
		return
	}

	dc := entity.Discord{
		Content:  "ถอนจำนวน : " + strconv.FormatInt(number, 10) + " บาท :( " + me.Email + " )",
		Username: me.FirstName + " " + me.LastName,
	}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(dc)
	http.Post(url, "application/json; charset=utf-8", b)
}

//SendWelletToDiscord send Wellet
func SendWelletToDiscord(me *entity.Me, number, mainWallet int64) {
	url := configfile.NewYAMLReader("config/config-application.yaml").String("webhook_wallet")
	if url == "" {
		return
	}

	dc := entity.Discord{
		Content:  "เงินเข้า : " + strconv.FormatInt(number, 10) + " บาท:( " + me.Email + " ) จำนวนเงินทั้งหมด : " + strconv.FormatInt(mainWallet, 10) + " บาท",
		Username: me.FirstName + " " + me.LastName,
	}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(dc)
	http.Post(url, "application/json; charset=utf-8", b)
}
