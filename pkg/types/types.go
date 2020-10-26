package types

//Money - представляет собой денежную сумму в минимальных единицах (центы копейки, дирамы и т.д)
type Money int64

//Currency представляет код валюты
type Currency string

//Коды валют
const (
	TJS Currency = "TJS"
	RUB Currency = "RUB"
	USD Currency = "USD"
)

//PAN представляет номер карты
type PAN string
//Status of card
type Status string

//Card представляет информацию о платежной карты
type Card struct {
	ID 			int
	PAN 	 	PAN
	Balance  	Money
	MinBalance 	Money
	Currency 	Currency
	Color 	 	string
	Name 	 	string
	Active 	 	bool
}

//Payment представляет информацию о платеже 
type Payment struct {
	ID 			string
	AccountID	int64
	Amount 		Money
	Category 	PaymentCategory
	Status 		PaymentStatus
}

//PaymentSource представляет информацию короткую инфо о картах пользователья 
type PaymentSource struct {
	Type string // 'card'
	Number string // номер вида '5058 xxxx xxxx 8888'
	Balance Money // баланс в дирамах
	
}
//Category представляет cобой
type Category string
//Phone
type Phone string

// Accounts
type Account struct {
	ID int64 // 'card'
	Phone Phone // номер вида '5058 xxxx xxxx 8888'
	Balance Money // баланс в дирамах
}

type PaymentCategory string

type PaymentStatus string

//Предопределённые статусы платежей
const (
	PaymentStatusOk PaymentStatus = "OK"
	PaymentStatusFail PaymentStatus = "FAIL"
	PaymentStatusInProgress PaymentStatus = "INPROGRESS"
)

//Favorite представляет инфо о избранном платеже
type Favorite struct {
	ID 			string
	AccountID	int64
	Name		string
	Amount		Money
	Category	PaymentCategory
}

type Progress struct{
	Part  int
	Result Money
}
