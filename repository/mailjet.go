package repository

import (
	"fmt"
	"strconv"
	"time"

	"github.com/acoshift/configfile"
	mailjet "github.com/mailjet/mailjet-apiv3-go"
	"github.com/workdestiny/oilbets/config"
)

type sendEmailModel struct {
	Name string
	Link string
}

// SendEmailVerify send email to verity
func SendEmailVerify(baseURL, email, code, name string) {

	r := configfile.NewReader("config/secret")

	mj := mailjet.NewMailjetClient(r.String("mailjet_public_key"), r.String("mailjet_secret_key"))
	param := &mailjet.InfoSendMail{
		FromEmail: "noreply@x.com",
		FromName:  "x",
		Recipients: []mailjet.Recipient{
			mailjet.Recipient{
				Email: email,
			},
		},
		MjTemplateID:       "194696",
		MjTemplateLanguage: "true",
		Vars: sendEmailModel{
			Name: name,
			Link: baseURL + config.URLVerifyEmailCallBack + "?email=" + email + "&code=" + code,
		},
		Subject: "Please confirm your email  #" + strconv.FormatInt(time.Now().Unix(), 10),
	}
	mj.SendMail(param)

}

// SendEmailVerifyCreator send email to verify is creator in x.com
func SendEmailVerifyCreator(email string, id string, name string, imageIDCard, imageFace, IDCardContentType, FaceContentType string) bool {
	r := configfile.NewReader("config/secret")

	mailjetClient := mailjet.NewMailjetClient(r.String("mailjet_public_key"), r.String("mailjet_secret_key"))
	messagesInfo := []mailjet.InfoMessagesV31{
		mailjet.InfoMessagesV31{
			From: &mailjet.RecipientV31{
				Email: "noreply@x.com",
				Name:  name,
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: "support@x.com",
				},
			},
			Subject:  "ต้องการยืนยันบัญชี ID " + id,
			TextPart: "ID " + id + "  ( " + name + " ) #" + strconv.FormatInt(time.Now().Unix(), 10),
			HTMLPart: "<h3>รูปบัตรประชาชน x! \n" + " ID(" + id + ")  ชื่อ(" + name + ") Email(" + email + ") <img src=\"cid:id1\"> </h3><br/>",
			Attachments: &mailjet.AttachmentsV31{
				mailjet.AttachmentV31{
					ContentType:   IDCardContentType,
					Filename:      "images.jpg",
					Base64Content: imageIDCard,
				},
				mailjet.AttachmentV31{
					ContentType:   FaceContentType,
					Filename:      "images.jpg",
					Base64Content: imageFace,
				},
			},
			ReplyTo: &mailjet.RecipientV31{
				Email: email,
			},
		},
	}
	messages := mailjet.MessagesV31{Info: messagesInfo}
	_, err := mailjetClient.SendMailV31(&messages)
	if err != nil {
		return false
	}

	return true
}

// SendEmailVerifyBookbank send email to verify bookbank in x.com
func SendEmailVerifyBookbank(email, id, name, image, contentType, bookbankName, bookbankNumber, bankName string) bool {
	r := configfile.NewReader("config/secret")

	mailjetClient := mailjet.NewMailjetClient(r.String("mailjet_public_key"), r.String("mailjet_secret_key"))
	messagesInfo := []mailjet.InfoMessagesV31{
		mailjet.InfoMessagesV31{
			From: &mailjet.RecipientV31{
				Email: "noreply@x.com",
				Name:  name,
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: "support@x.com",
				},
			},
			Subject:  "ต้องการยืนยันบัญชีธนาคาร ID " + id + " #" + strconv.FormatInt(time.Now().Unix(), 10),
			TextPart: "ID " + id + "  ( " + name + " ) #" + strconv.FormatInt(time.Now().Unix(), 10),
			HTMLPart: "<h3>รูปบัญชีธนาคาร x! \n" + " ID(" + id + ")  ชื่อ(" + name + ") Email(" + email + ")\n ชื่อบัญชี(" + bookbankName + ")\n เลขบัญชี(" + bookbankNumber + ")\n ธนาคาร(" + bankName + ")\n <img src=\"cid:id1\"> </h3><br/>",
			Attachments: &mailjet.AttachmentsV31{
				mailjet.AttachmentV31{
					ContentType:   contentType,
					Filename:      "images.jpg",
					Base64Content: image,
				},
			},
			ReplyTo: &mailjet.RecipientV31{
				Email: email,
			},
		},
	}
	messages := mailjet.MessagesV31{Info: messagesInfo}
	_, err := mailjetClient.SendMailV31(&messages)
	if err != nil {
		return false
	}

	return true
}

// SendEmailRevenue send email to verify bookbank in x.com
func SendEmailRevenue(id, name, email, gapID, gapName string) bool {
	r := configfile.NewReader("config/secret")

	mailjetClient := mailjet.NewMailjetClient(r.String("mailjet_public_key"), r.String("mailjet_secret_key"))
	messagesInfo := []mailjet.InfoMessagesV31{
		mailjet.InfoMessagesV31{
			From: &mailjet.RecipientV31{
				Email: "noreply@x.com",
				Name:  name,
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: "support@x.com",
				},
			},
			Subject:  "ส่งใบรับเงินของเพจ " + gapName + "  #" + strconv.FormatInt(time.Now().Unix(), 10),
			TextPart: "ส่งใบรับเงินของเพจ " + gapName + "\nAccount ID :" + id + "  ( " + name + " ) #" + strconv.FormatInt(time.Now().Unix(), 10) + "\n" + "Gap ID :" + gapID + " ( " + gapName + " )",
			ReplyTo: &mailjet.RecipientV31{
				Email: email,
			},
		},
	}
	messages := mailjet.MessagesV31{Info: messagesInfo}
	_, err := mailjetClient.SendMailV31(&messages)
	if err != nil {
		return false
	}

	return true
}

// SendEmailAdminDelete send email to user admin delete post
func SendEmailAdminDelete(email, title string) bool {
	r := configfile.NewReader("config/secret")

	mailjetClient := mailjet.NewMailjetClient(r.String("mailjet_public_key"), r.String("mailjet_secret_key"))
	messagesInfo := []mailjet.InfoMessagesV31{
		mailjet.InfoMessagesV31{
			From: &mailjet.RecipientV31{
				Email: "noreply@x.com",
				Name:  "x",
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: email,
				},
			},
			Subject:  `คอนเทนต์ "` + title + `" ถูกลบ`,
			TextPart: `คอนเทนต์ "` + title + `" ถูกลบ` + "\n\n" + `เนื่องจากคอนเทนต์นี้ไม่ปฏิบัติตามข้อกำหนดและเงื่อนไขของทางเว็บไซต์` + "\n" + `ซึ่งอาจส่งผลกระทบให้เกิดความเสียหาย ทางเราจึงดำเนินการลบคอนเทนต์ดังกล่าวโดยไม่ต้องแจ้งให้ทราบล่วงหน้า` + "\n\n" + `แจ้งเพื่อทราบ ขอบคุณค่ะ`,
			ReplyTo: &mailjet.RecipientV31{
				Email: email,
			},
		},
	}
	messages := mailjet.MessagesV31{Info: messagesInfo}
	_, err := mailjetClient.SendMailV31(&messages)
	if err != nil {
		return false
	}

	return true
}

// SendEmailAdminReject send email to user admin reject post
func SendEmailAdminReject(email, title, note string) bool {
	r := configfile.NewReader("config/secret")

	mailjetClient := mailjet.NewMailjetClient(r.String("mailjet_public_key"), r.String("mailjet_secret_key"))
	messagesInfo := []mailjet.InfoMessagesV31{
		mailjet.InfoMessagesV31{
			From: &mailjet.RecipientV31{
				Email: "noreply@x.com",
				Name:  "x",
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: email,
				},
			},
			Subject:  `คอนเทนต์ "` + title + `" ไม่สามารถรับเงินได้`,
			TextPart: `คอนเทนต์ "` + title + `" ไม่สามารถรับเงินได้` + "\n\n" + `เนื่องจากคอนเทนต์นี้ไม่ปฏิบัติตามข้อกำหนดและเงื่อนไขของทางเว็บไซต์` + "\n" + `*หมายเหตุ ` + note + ` จึงถูกระงับ ไม่สามารถสร้างรายได้ได้อีก` + "\n\n" + `แจ้งเพื่อทราบ ขอบคุณค่ะ`,
			ReplyTo: &mailjet.RecipientV31{
				Email: email,
			},
		},
	}
	messages := mailjet.MessagesV31{Info: messagesInfo}
	_, err := mailjetClient.SendMailV31(&messages)
	if err != nil {
		return false
	}

	return true
}

// SendEmailForgetPassword send email user forget password
func SendEmailForgetPassword(baseURL, email, code string) bool {

	r := configfile.NewReader("config/secret")

	mj := mailjet.NewMailjetClient(r.String("mailjet_public_key"), r.String("mailjet_secret_key"))
	param := &mailjet.InfoSendMail{
		FromEmail: "noreply@x.com",
		FromName:  "x",
		Recipients: []mailjet.Recipient{
			mailjet.Recipient{
				Email: email,
			},
		},
		MjTemplateID:       "194909",
		MjTemplateLanguage: "true",
		Vars: sendEmailModel{
			Link: baseURL + config.URLResetPasswordCallBack + "?email=" + email + "&code=" + code,
		},
		Subject: "ส่งคำขอลืมรหัสผ่าน โปรดยืนยันการเปลี่ยนรหัสผ่านใหม่  #" + strconv.FormatInt(time.Now().Unix(), 10),
		// TextPart: "",
	}
	_, err := mj.SendMail(param)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

// // SendEmailUpdateNewEmail send email to update new email
// func SendEmailUpdateNewEmail(baseURL, email, code string, name string) bool {

// 	r := configfile.NewReader("config/secret")

// 	mj := mailjet.NewMailjetClient(r.String("mailjet_public_key"), r.String("mailjet_secret_key"))
// 	param := &mailjet.InfoSendMail{
// 		FromEmail: "noreply@x.com",
// 		FromName:  "x",
// 		Recipients: []mailjet.Recipient{
// 			mailjet.Recipient{
// 				Email: email,
// 			},
// 		},
// 		MjTemplateID:       "194915",
// 		MjTemplateLanguage: "true",
// 		Vars: sendEmailModel{
// 			Name: name,
// 			Link: baseURL + config.URLUpdateEmailCallBack + "?email=" + email + "&code=" + code,
// 		},
// 		Subject: "ส่งคำขอเปลี่ยนแปลงอีเมล โปรดยืนยันอีเมลใหม่ของคุณ  #" + strconv.FormatInt(time.Now().Unix(), 10),
// 		// TextPart: "",
// 	}
// 	_, err := mj.SendMail(param)
// 	if err != nil {
// 		fmt.Println(err)
// 		return false
// 	}
// 	return true
// }

// // SendEmailUpdateOldEmail send to old email user can recovery to this email
// func SendEmailUpdateOldEmail(baseURL, email, code string, name string) bool {

// 	r := configfile.NewReader("config/secret")

// 	mj := mailjet.NewMailjetClient(r.String("mailjet_public_key"), r.String("mailjet_secret_key"))
// 	param := &mailjet.InfoSendMail{
// 		FromEmail: "noreply@x.com",
// 		FromName:  "x",
// 		Recipients: []mailjet.Recipient{
// 			mailjet.Recipient{
// 				Email: email,
// 			},
// 		},
// 		MjTemplateID:       "194874",
// 		MjTemplateLanguage: "true",
// 		Vars: sendEmailModel{
// 			Name: name,
// 			Link: baseURL, config.URLRestoreAccountCallBack + "?email=" + email + "&code=" + code,
// 		},
// 		Subject: "แจ้งเตือน: อีเมลของคุณมีการเปลี่ยนแปลง กรุณาตรวจสอบ!  #" + strconv.FormatInt(time.Now().Unix(), 10),
// 		// TextPart: "",
// 	}
// 	_, err := mj.SendMail(param)
// 	if err != nil {
// 		fmt.Println(err)
// 		return false
// 	}
// 	return true
// }
