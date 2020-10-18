package wallet

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/RAZ-os/wallet/pkg/types"
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
			break
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
			break
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
		/* damp, err := os.Create(DumpDir)
		 if err != nil {
			 return err
		 }*/

		content := ""

		for index, account := range s.accounts {
			content += strconv.FormatInt(int64(account.ID), 10) + ";" + string(account.Phone) + ";" + strconv.FormatInt(int64(account.Balance), 10)
			if index != len(s.accounts)-1 {
				content += "|"
			}
		}

		err := ioutil.WriteFile(DumpDir, []byte(content), 0644)
		if err != nil {
			log.Print(err)
			return err
		}

	}

	if FavLen > 0 {
		DumpDir := dir + "/favorites.dump"

		content := ""

		for index, favorite := range s.favorites {
			content += strconv.FormatInt(int64(favorite.AccountID), 10) + ";" + strconv.FormatInt(int64(favorite.Amount), 10) + ";" + string(favorite.Category) + ";" + string(favorite.ID) + ";" + string(favorite.Name)
			if index != len(s.accounts)-1 {
				content += "|"
			}
		}

		err := ioutil.WriteFile(DumpDir, []byte(content), 0644)
		if err != nil {
			log.Print(err)
			return err
		}

	/*	defer func() {
			if cerr := DumpDir.Close(); cerr != nil {
				log.Print(cerr)
			}
		}()*/
	}

	if PayLen > 0 {
		DumpDir := dir + "/payments.dump"

		content := ""

		for index, payment := range s.payments {
			content += strconv.FormatInt(int64(payment.AccountID), 10) + ";" + strconv.FormatInt(int64(payment.Amount), 10) + ";" + string(payment.Category) + ";" + string(payment.ID) + ";" + string(payment.Status)
			if index != len(s.accounts)-1 {
				content += "|"
			}
		}

		err := ioutil.WriteFile(DumpDir, []byte(content), 0644)
		if err != nil {
			log.Print(err)
			return err
		}

	/*	defer func() {
			if cerr := DumpDir.Close(); cerr != nil {
				log.Print(cerr)
			}
		}()*/
	}

   return nil
}
