package wallet

import (
	"bufio"
	"errors"
	"io"
	//"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"github.com/RAZ-os/wallet/pkg/types"
	//"wallet/pkg/types"
	"github.com/google/uuid"
)

var ErrPhoneRegistered = errors.New("phone already registered")
var ErrAmountMustBePositive = errors.New("amount must be greater than zero")
var ErrAccountNotFound = errors.New("account not found")
var ErrPaymentNotFound = errors.New("payment not found")
var ErrFavoriteNotFound = errors.New("favorite not found")

type Service struct {
	nextAccountID int64
	accounts      []*types.Account
	payments      []*types.Payment
	favorites     []*types.Favorite
}

func (s *Service) RegisterAccount(phone types.Phone) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.Phone == phone {
			return nil, ErrPhoneRegistered
		}
	}

	s.nextAccountID++
	account := &types.Account{
		ID:      s.nextAccountID,
		Phone:   phone,
		Balance: 0,
	}
	s.accounts = append(s.accounts, account)

	return account, nil
}

func (s *Service) Deposit(accountID int64, amount types.Money) error {
	if amount <= 0 {
		return ErrAmountMustBePositive
	}

	var account *types.Account

	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}

	if account == nil {
		return ErrAccountNotFound
	}

	account.Balance += amount
	return nil
}

func (s *Service) Pay(accountID int64, amount types.Money, category types.PaymentCategory) (*types.Payment, error) {
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}

	account, accountErr := s.FindAccountByID(accountID)

	if account == nil {
		return nil, accountErr
	}

	account.Balance -= amount
	paymentID := uuid.New().String()
	payment := &types.Payment{
		ID:        paymentID,
		AccountID: accountID,
		Amount:    amount,
		Category:  category,
		Status:    types.PaymentStatusInProgress,
	}

	s.payments = append(s.payments, payment)
	return payment, nil
}

func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {
	var account *types.Account

	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}

	if account == nil {
		return nil, ErrAccountNotFound
	}

	return account, nil
}

func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {
	for _, payment := range s.payments {
		if payment.ID == paymentID {
			return payment, nil
			//break
		}
	}

	return nil, ErrPaymentNotFound
}

func (s *Service) Reject(paymentID string) error {

	payment, err := s.FindPaymentByID(paymentID)

	if err != nil {
		return err
	}

	account, err := s.FindAccountByID(payment.AccountID)

	if err != nil {
		return err
	}

	account.Balance += payment.Amount
	payment.Status = types.PaymentStatusFail

	return nil
}

func (s *Service) Repeat(paymentID string) (*types.Payment, error) {
	oldPayment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}

	account, err := s.FindAccountByID(oldPayment.AccountID)
	if err != nil {
		return nil, err
	}

	newPayment, err := s.Pay(account.ID, oldPayment.Amount, oldPayment.Category)
	if err != nil {
		return nil, err
	}

	return newPayment, nil
}

func (s *Service) FavoritePayment(paymentID string, name string) (*types.Favorite, error) {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}

	account, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		return nil, err
	}

	favoriteID := uuid.New().String()
	favorite := &types.Favorite{
		ID:        favoriteID,
		AccountID: account.ID,
		Name:      name,
		Amount:    payment.Amount,
		Category:  payment.Category,
	}

	s.favorites = append(s.favorites, favorite)
	return favorite, nil
}

func (s *Service) FindFavoritePaymentByID(favorityID string) (*types.Favorite, error) {
	for _, favorite := range s.favorites {
		if favorite.ID == favorityID {
			return favorite, nil
			//break
		}
	}

	return nil, ErrFavoriteNotFound
}

func (s *Service) PayFromFavorite(favoriteID string) (*types.Payment, error) {
	favorite, err := s.FindFavoritePaymentByID(favoriteID)
	if err != nil {
		return nil, err
	}

	account, err := s.FindAccountByID(favorite.AccountID)
	if err != nil {
		return nil, err
	}

	payment, err := s.Pay(account.ID, favorite.Amount, favorite.Category)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

///////////////////////////////////////////////////////////
func (s *Service) ExportToFile(path string) error {

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	content := ""

	for index, account := range s.accounts {
		content += strconv.FormatInt(int64(account.ID), 10) + ";" + string(account.Phone) + ";" + strconv.FormatInt(int64(account.Balance), 10)
		if index != len(s.accounts)-1 {
			content += "|"
		}
	}

	_, err = file.Write([]byte(content))
	if err != nil {
		return err
	}

	//defer func(){
	err = file.Close()
	if err != nil {
		return err
	}
	//}()

	return nil
}

//////////////////////////////////////////////
func (s *Service) ImportFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	buf := make([]byte, 4096)
	read, err := file.Read(buf)
	if err != nil {
		return err
	}

	content := string(buf[:read])

	accounts := strings.Split(content, "|")

	for _, account := range accounts {
		accountConvArr := strings.Split(account, ";")

		importedAccount, err := s.RegisterAccount(types.Phone(accountConvArr[1])) //account phone
		if err != nil {
			return err
		}

		balance, err := strconv.ParseInt(accountConvArr[2], 10, 64) //account balance
		if err != nil {
			return err
		}

		if balance > 0 {
			err = s.Deposit(importedAccount.ID, types.Money(balance))
			if err != nil {
				return err
			}
		}
	}

	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}

////////////////
func (s *Service) Export(dir string) error {
	FavLen := len(s.favorites)
	PayLen := len(s.payments)
	AccLen := len(s.accounts)

	if AccLen > 0 {

		DumpDir := dir + "/accounts.dump"
		
		file, err := os.Create(DumpDir)
    if err != nil {
        return err
    }
    defer func() {
	  if cerr:=	file.Close(); cerr != nil {
     log.Print(cerr)
	  }
	}()

	for _, account := range s.accounts {
		txtitemue := []byte(strconv.FormatInt(int64(account.ID), 10) + string(";") + string(account.Phone) + string(";") + strconv.FormatInt(int64(account.Balance), 10) + string(";") + string('\n'))
		_, err = file.Write(txtitemue)
		if err != nil {
			return err
		}
	}
}

	if FavLen > 0 { //// Данные есть

		DumpDir := dir + "/favorites.dump"

			file, err := os.Create(DumpDir)
			if err != nil {
				return err
			}
			defer func() {
			  if cerr:=	file.Close(); cerr != nil {
			 log.Print(cerr)
			  }
			}()

			for _, fav := range s.favorites {
				text := []byte(fav.ID + ";" + strconv.FormatInt(int64(fav.AccountID), 10) + ";" + fav.Name + ";" + strconv.FormatInt(int64(fav.Amount), 10) + ";" + string(fav.Category) + string('\n'))
				_, err := file.Write(text)
				if err != nil {
					log.Print(err)
					return err
				}
			}
		}

    if PayLen > 0 {

		DumpDir := dir + "/payments.dump"
	   
			file, err := os.Create(DumpDir)
			if err != nil {
				return err
			}
			defer func() {
			  if cerr:=	file.Close(); cerr != nil {
			 log.Print(cerr)
			  }
			}()
		
			for _, payment := range s.payments {
				txtitemue := []byte(string(payment.ID) + string(";") + strconv.FormatInt(int64(payment.AccountID), 10) + string(";") + strconv.FormatInt(int64(payment.Amount), 10) + string(";") + string(payment.Category) + string(";") + string(payment.Status) + string(";") + string('\n'))
				_, err = file.Write(txtitemue)
				if err != nil {
					return err
				}
			}
	}

   return nil
}

func (s *Service) Import(dir string) error {
	//For accounts
	accountFile := "/accounts.dump"
	src, err := os.Open(dir + accountFile)
	if err != nil {
		log.Print("There is no %w file", accountFile)
	} else {
		defer func() {
			if cerr := src.Close(); cerr != nil {
				log.Print(cerr)
			}
		}()

		reader := bufio.NewReader(src)
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				log.Print(line)
				break
			}
			if err != nil {
				log.Print(err)
				return err
			}

			item := strings.Split(line, ";")

			id, err := strconv.ParseInt(item[0], 10, 64)
			if err != nil {
				log.Print(err)
				return err
			}

			phone := item[1]

			balance, err := strconv.ParseInt(item[2], 10, 64)
			if err != nil {
				log.Print(err)
				return err
			}

			findAccount, _ := s.FindAccountByID(id)
			if findAccount != nil {
				findAccount.Phone = types.Phone(phone)
				findAccount.Balance = types.Money(balance)
			} else {
				s.nextAccountID = id
				newAcc := &types.Account{
					ID:      s.nextAccountID,
					Phone:   types.Phone(phone),
					Balance: types.Money(balance),
				}

				s.accounts = append(s.accounts, newAcc)
			}
		}
		log.Print("Imported")

	}

	// For Payments
	paymentsFile := "/payments.dump"
	paySrc, err := os.Open(dir + paymentsFile)
	if err != nil {
		log.Print("There is no %w file", paymentsFile)
	} else {
		defer func() {
			if cerr := paySrc.Close(); cerr != nil {
				log.Print(cerr)
			}
		}()

		payReader := bufio.NewReader(paySrc)
		for {
			payLine, err := payReader.ReadString('\n')
			if err == io.EOF {
				log.Print(payLine)
				break
			}
			if err != nil {
				log.Print(err)
				return err
			}

			item := strings.Split(payLine, ";")

			id := string(item[0])
			accID, err := strconv.ParseInt(item[1], 10, 64)
			if err != nil {
				log.Print(err)
				return err
			}

			amount, err := strconv.ParseInt(item[2], 10, 64)
			if err != nil {
				log.Print(err)
				return err
			}

			category := item[3]

			status := item[4]

			findPay, _ := s.FindPaymentByID(id)
			if findPay != nil {
				findPay.AccountID = accID
				findPay.Amount = types.Money(amount)
				findPay.Category = types.PaymentCategory(category)
				findPay.Status = types.PaymentStatus(status)
			} else {
				newPay := &types.Payment{
					ID:        id,
					AccountID: accID,
					Amount:    types.Money(amount),
					Category:  types.PaymentCategory(category),
					Status:    types.PaymentStatus(status),
				}

				s.payments = append(s.payments, newPay)
			}
		}
		log.Print("Imported")
	}

	//For favorites
	favoritesFile := "/favorites.dump"
	favFile, err := os.Open(dir + favoritesFile)
	if err != nil {
		log.Print("There is no such %w a file", favoritesFile)
	} else {
		reader := bufio.NewReader(favFile)
		for {
			favLine, err := reader.ReadString('\n')
			if err == io.EOF {
				log.Print(favLine)
				break
			}
			if err != nil {
				log.Print(err)
				return err
			}

			item := strings.Split(favLine, ";")

			id := item[0]
			accID, err := strconv.ParseInt(item[1], 10, 64)
			if err != nil {
				log.Print(err)
				return err
			}
			name := item[2]
			amount, err := strconv.ParseInt(item[3], 10, 64)
			if err != nil {
				log.Print(err)
				return err
			}
			category := item[4]

			findFav, _ := s.FindFavoritePaymentByID(id)
			if findFav != nil {
				findFav.AccountID = accID
				findFav.Amount = types.Money(amount)
				findFav.Name = name
				findFav.Category = types.PaymentCategory(category)
			} else {
				newFav := &types.Favorite{
					ID:        id,
					AccountID: accID,
					Name:      name,
					Amount:    types.Money(amount),
					Category:  types.PaymentCategory(category),
				}
				s.favorites = append(s.favorites, newFav)
			}
		}
		log.Print("Imported")
	}

	return nil
}

/*func(s *Service) Import(dir string) error{

	if _, err := os.Stat(dir+"/accounts.dump"); err == nil { //// Accounts start
		file, err := os.Open(dir+"/accounts.dump")
		if err != nil {
			log.Print(err)
		}
		defer func() {
			if cerr := file.Close(); cerr != nil {
				log.Print(cerr)
			}
		}()
		reader := bufio.NewReader(file)
	for{
	   line, err := reader.ReadString('\n')
	   if err == io.EOF{
		   log.Print(line)
		   break
	   }
		
	if err != nil{
		log.Print(err)
		break
	}
	log.Print(line)
}
 } // //// Accounts end 

if _, err := os.Stat(dir+"/payments.dump"); err == nil {
	file, err := os.Open(dir+"/payments.dump")
	if err != nil {
		log.Print(err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()
	reader := bufio.NewReader(file)
for{
   line, err := reader.ReadString('\n')
   if err == io.EOF{
	   log.Print(line)
	   break
   }
	
if err != nil{
	log.Print(err)
	break
}
log.Print(line)
}

}
if _, err := os.Stat(dir+"/favorites.dump"); err == nil {
	file, err := os.Open(dir+"/favorites.dump")
	if err != nil {
		log.Print(err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Print(cerr)
		}
	}()
	reader := bufio.NewReader(file)
for{
   line, err := reader.ReadString('\n')
   if err == io.EOF{
	   log.Print(line)
	   break
   }
	
if err != nil{
	log.Print(err)
	break
}
log.Print(line)
}

}
return nil
}*/