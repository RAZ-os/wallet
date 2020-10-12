package main

import (
	"fmt"
	//"strconv"
	"github.com/RAZ-os/wallet/pkg/wallet"
) 

func main(){

	svc := &wallet.Service{}
	path := "files/accounts.txt"
	//err := svc.ExportToFile(path)
	err := svc.ImportFromFile(path)

	account, err1 := svc.RegisterAccount("+992901000876")

	
	if err != nil{
		fmt.Println(err)
		return
	}
	if err1 != nil{
		fmt.Println(err1)
		return
	}

/*	payment, err := svc.Pay(account.ID, 20, "auto")
	
	if err != nil {
		fmt.Println(err)
		return
	}

	favorite, err := svc.FavoritePayment(payment.ID, "My Favorite Payment")

	if err != nil {
		fmt.Println(err)
		return
	}
*/	
	fmt.Println(*account)
	//fmt.Println(strconv.FormatInt(66,2))
	//fmt.Println(payment)
	//fmt.Println(favorite)
}