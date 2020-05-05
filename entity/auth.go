package entity

//FacebookOauth2 model callback
type FacebookOauth2 struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Email   string  `json:"email"`
	Picture Picture `json:"picture"`
}

//Picture FacebookOauth2 images
type Picture struct {
	Data Data `json:"data"`
}

//Data  FacebookOauth2.Picture
type Data struct {
	URL string `json:"url"`
}
