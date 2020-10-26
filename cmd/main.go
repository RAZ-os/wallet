package main

import (
	"fmt"
	//"wallet/pkg/types"
	//"strconv"
	"github.com/RAZ-os/wallet/pkg/wallet"
	//"github.com/RAZ-os/wallet/pkg/types"
) 

func main(){

	/*payments := []types.Payment{
		{ID: 1, Category: "auto", Amount: 2_000_000},
		{ID: 2, Category: "food", Amount: 2_000_000},
		{ID: 3, Category: "auto", Amount: 3_000_000},
		{ID: 4, Category: "auto", Amount: 4_000_000},
		{ID: 5, Category: "fun",  Amount: 5_000_000},
	}*/

	svc := &wallet.Service{}
	account, err := svc.RegisterAccount("+992901000876")
	payment, err := svc.Pay(1, 5_000_000, "fun")
	payment, err = svc.Pay(1, 10_000_000, "auto")
	//favorite, err := svc.FavoritePayment(payment.ID, "My Favorite Payment")
	//sum := svc.SumPayments(1)
	pro := svc.SumPaymentsWithProgress()
	//path := "files/accounts.txt"
	//dir := "files"
	//err = svc.Export(dir)
	//err = svc.Import(dir)
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
	//fmt.Println(favorite)
	fmt.Println(pro)
	fmt.Println(err)
}