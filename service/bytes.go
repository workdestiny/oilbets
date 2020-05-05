package service

import "io/ioutil"

// MustLoadBytesFromFile loads bytes from given file or panic if error
func MustLoadBytesFromFile(filename string) []byte {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return b
}
