package main

import (
	"fmt"
	//"wallet/pkg/types"
	//"strconv"
	"github.com/RAZ-os/wallet/pkg/wallet"
) 

func main(){

	 svc := &wallet.Service{}
	account, err := svc.RegisterAccount("+992901000876")
	payment, err := svc.Pay(account.ID, 20, "auto")
	favorite, err := svc.FavoritePayment(payment.ID, "My Favorite Payment")
	//path := "files/accounts.txt"
	dir := "files"
	err = svc.Export(dir)
	err = svc.Import(dir)
	//err = svc.ExportToFile(path)
	//err = svc.ImportFromFile(path)
	
	if err != nil{
		fmt.Println(err)
		return
	}

/*	if err_a != nil{
		fmt.Println(err_a)
		return
	}

	if err_p != nil {
		fmt.Println(err_p)
		return
	}

	if err_f != nil {
		fmt.Println(err_f)
		return
	}
	*/
	fmt.Println(*account)
	fmt.Println(payment)
	fmt.Println(favorite)
	fmt.Println(err)
}