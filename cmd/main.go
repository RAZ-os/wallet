package main

import (
	"github.com/RAZ-os/wallet/pkg/wallet"
	"fmt"
) 

func main(){

	svc := &wallet.Service{}

	account, err := svc.RegisterAccount("+992000000002")
	
	if err != nil{
		fmt.Println(err)
		return
	}

	payment, err := svc.Pay(account.ID, 20, "auto")
	
	if err != nil {
		fmt.Println(err)
		return
	}

	favorite, err := svc.FavoritePayment(payment.ID, "My Favorite Payment")

	if err != nil {
		fmt.Println(err)
		return
	}
	
	fmt.Println(account.Balance)
	fmt.Println(payment)
	fmt.Println(favorite)
}