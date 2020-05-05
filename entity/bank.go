package entity

// BankList is type
type BankList struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetBankNameByID given BankID return BankName
func GetBankNameByID(ID string) string {
	for _, v := range BankListData {
		if ID == v.ID {
			return v.Name
		}
	}
	return ""
}

// GetBankIDByName given BankName return BankID
func GetBankIDByName(n string) string {
	for _, v := range BankListData {
		if n == v.Name {
			return v.ID
		}
	}
	return ""
}

// BankListData is data
var BankListData = []BankList{
	{"BKKBTHBK", "กรุงเทพ"},
	{"KASITHBK", "กสิกรไทย"},
	{"KRTHTHBK", "กรุงไทย"},
	{"TMBKTHB", "ทหารไทย"},
	{"SICOTHBK", "ไทยพาณิชย์"},
	{"AYUDTHBK", "กรุงศรีอยุธยา"},
	{"KIFITHB1", "เกียรตินาคิน"},
	{"UBOBTHBK", "ซีไอเอ็มบีไทย"},
	{"TFPCTHB1", "ทิสโก้"},
	{"THBKTHBK", "ธนชาต"},
	{"UOVBTHBK", "ยูโอบี"},
	{"LAHRTHB1", "แลนด์ แอนด์ เฮาส์"},
	{"GSBATHBK", "ออมสิน"},
}
