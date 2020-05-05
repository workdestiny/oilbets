package repository

// func convMe(user *entity.UserModel) *entity.Me {

// 	a := []entity.AboutMe{}
// 	isVerify := false

// 	for i := 0; i < len(user.AboutMe); i++ {

// 		a = append(a, entity.AboutMe{
// 			Type:             user.AboutMe[i].Type,
// 			Value:            user.AboutMe[i].Value,
// 			StatusUpdateTime: user.AboutMe[i].StatusUpdateTime,
// 		})
// 	}
// 	date := user.BirthDate.Format("2006-01-02")

// 	var day, month, year string
// 	if len(date) == 10 {
// 		year = date[0:4]
// 		month = date[5:7]
// 		day = date[8:10]
// 	}

// 	if user.Verify.UserVerifyLevel > 0 {
// 		isVerify = true
// 	}

// 	return &entity.Me{
// 		ID:        user.ID(),
// 		Email:     user.Email,
// 		FirstName: user.FirstName,
// 		LastName:  user.LastName,
// 		DisplayImage: entity.DisplayImage{
// 			Normal: user.DisplayImage,
// 			Mini:   user.DisplayImageMini,
// 		},
// 		BirthDate: entity.BirthDate{
// 			Day:   day,
// 			Month: month,
// 			Year:  year,
// 		},
// 		Gender: user.Gender,
// 		Contact: entity.Contact{
// 			Address: user.Contact.Address,
// 			City:    user.Contact.City,
// 			Country: user.Contact.Country,
// 		},
// 		Role:    user.Role.String(),
// 		AboutMe: a,
// 		Count: entity.CountMe{
// 			Topic: convbox.ShortNumber(user.Count.Topic),
// 			Page:  convbox.ShortNumber(user.Count.Page),
// 		},
// 		EmailVerify: entity.EmailVerifyMe{
// 			Status: user.EmailVerify.Status,
// 			At:     user.EmailVerify.At,
// 		},
// 		ToolTip: entity.ToolTip{
// 			Category: user.ToolTip.Category,
// 			Follow:   user.ToolTip.Follow,
// 			Post:     user.ToolTip.Post,
// 		},
// 		HasPage:  user.HasPage,
// 		IsVerify: isVerify,
// 		IsSignin: true,
// 	}

// }
