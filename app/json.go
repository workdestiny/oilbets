package app

import (
	"encoding/json"
	"io"

	"github.com/workdestiny/oilbets/entity"
)

func bindJSON(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

func checkProvinceID(id int) bool {

	for _, p := range entity.ProvinceData {
		if p.ID == id {
			return true
		}
	}

	return false
}

func checkTagtopic(tag []string, t string) bool {

	if len(tag) >= 5 {
		return false
	}

	for _, v := range tag {
		if v == t {
			return false
		}
	}

	return true
}
