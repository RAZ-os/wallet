package wallet

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/RAZ-os/wallet/pkg/types"
	"github.com/google/uuid"
	//"os"
	//"log"
)

type testService struct{
	*Service
}

type testAccount struct {
	phone types.Phone
	balance types.Money
	payments []struct {
		amount 		types.Money
		category 	types.PaymentCategory
	}
}

var defaultTestAccount = testAccount {
	phone: "+992901000876",
	balance: 4_000_00,
	payments: []struct {
		amount types.Money
		category types.PaymentCategory
	}{
		{amount: 4_000_00, category: "auto"},
	},
}

func newTestService() *testService {
	return &testService{Service: &Service{}}
}

func (s *testService) addAccount(data testAccount) (*types.Account, []*types.Payment, []*types.Favorite, error){
	//региструем там пользователя
	account, err := s.RegisterAccount(data.phone)
	if(err != nil) {
		return nil, nil, nil, fmt.Errorf("can't register account, error = %v", err)
	}

	//пополняем его счет
	err = s.Deposit(account.ID, data.balance)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("can't deposit account, error = %v", err)
	}

	payments := make([]*types.Payment, len(data.payments))
	favorites := make([]*types.Favorite, len(data.payments))

	for i, payment := range data.payments {
		//тогда здесь работаем просто через index, а не через append
		payments[i], err = s.Pay(account.ID, payment.amount, payment.category)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("can't make payment, error = %v", err)
		}

		favorites[i], err = s.FavoritePayment(payments[i].ID, "My Favorite Payment #i")
		if err != nil {
			return nil, nil, nil, fmt.Errorf("can't make favorite payment, error = %v", err)
		}
	}

	return account, payments, favorites, nil
}

func TestService_FindAccountByID_found(t *testing.T){
	//создаём сервис
	s := newTestService()

	//регистриуем там пользователя
	account, _, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	got, err := s.FindAccountByID(account.ID);
	if err != nil {
		t.Errorf("FindAccountByID(): error = %v", err)
		return
	}

	//сравниваем аккаунты
	if !reflect.DeepEqual(account, got) {
		t.Errorf("FindAccountByID(): wrong account returned = %v", err)
		return
	}
}


func TestService_FindAccountByID_notfound(t *testing.T){
	//создаём сервис
	s := newTestService()

	//регистриуем там пользователя
	_, _, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	//попробуем найти несуществуйщий аккаунт
	fakeID := s.nextAccountID+1
	_, err = s.FindAccountByID(fakeID)
	if err == nil {
		t.Error("FindAccountByID(): must return error, returned nil")
		return
	}

	//сравниваем ошибки
	if err != ErrAccountNotFound {
		t.Errorf("FindAccountByID(): must return ErrAccountNotFound, returned = %v", err)
		return
	}
}

func TestService_FindPaymentByID_found(t *testing.T){
		//создаём сервис
		s := newTestService()

		//регистриуем там пользователя
		_, payments, _, err := s.addAccount(defaultTestAccount)
		if err != nil {
			t.Error(err)
			return
		}
	
		//попробуем найти платеж
		payment := payments[0]
		got, err := s.FindPaymentByID(payment.ID)
		if err != nil {
			t.Errorf("FindPaymentByID(): error = %v", err)
			return
		}

		//сравниваем платеж
		if !reflect.DeepEqual(payment, got) {
			t.Errorf("FindPaymentByID(): wrong payment returned = %v", err)
			return
		}
}

func TestService_FindPaymentByID_notfound(t *testing.T){
	//создаём сервис
	s := newTestService()

	//регистриуем там пользователя
	_, _, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}


	//попробуем найти несуществуйщий платеж
	_, err = s.FindPaymentByID(uuid.New().String())
	if err == nil {
		t.Error("FindPaymentByID(): must return error, returned nil")
		return
	}

	//сравниваем платеж
	if err != ErrPaymentNotFound {
		t.Errorf("FindPaymentByID(): must return ErrPaymentNotFound, returned = %v", err)
		return
	}
}


func TestService_FindFavoritePaymentByID_found(t *testing.T){
	//создаём сервис
	s := newTestService()

	//регистриуем там пользователя
	_, _, favorites, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	//попробуем найти избранное
	favorite := favorites[0]
	got, err := s.FindFavoritePaymentByID(favorite.ID)
	if err != nil {
		t.Errorf("FindFavoritePaymentByID(): error = %v", err)
		return
	}

	//сравниваем избранное
	if !reflect.DeepEqual(favorite, got) {
		t.Errorf("FindFavoritePaymentByID(): wrong payment returned = %v", err)
		return
	}
}

func TestService_FindFavoritePaymentByID_notfound(t *testing.T){
	//создаём сервис
	s := newTestService()

	//регистриуем там пользователя
	_, _, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}


	//попробуем найти несуществуйщий платеж
	_, err = s.FindFavoritePaymentByID(uuid.New().String())
	if err == nil {
		t.Error("FindFavoritePaymentByID(): must return error, returned nil")
		return
	}

	//сравниваем избранное
	if err != ErrFavoriteNotFound {
		t.Errorf("FindFavoritePaymentByID(): must return ErrPaymentNotFound, returned = %v", err)
		return
	}
}


func TestService_Reject_success(t *testing.T) {
	//создаём сервис
	s := newTestService()

	//регистриуем там пользователя
	_, payments, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	
	//попробуем отменить 
	payment := payments[0]
	err = s.Reject(payment.ID)
	if err != nil {
		t.Errorf("Reject(): error=%v", err)
		return
	}

	savedPayment, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("Reject(): can't find payment by id, error = %v", err)
		return
	}

	if savedPayment.Status != types.PaymentStatusFail {
		t.Errorf("Reject(): status didn't changed, payment=%v", savedPayment)
		return
	}

	savedAccount, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		t.Errorf("Reject(): can't find account by id, error=%v", err)
		return
	}

	if savedAccount.Balance != defaultTestAccount.balance {
		t.Errorf("Reject(): balance didn't changed, account=%v", savedAccount)
		return
	}
}

func TestService_Repeat_success(t *testing.T) {
	//создаём сервис
	s := newTestService()

	//регистриуем там пользователя
	_, payments, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	
	//попробуем отменить 
	oldPayment := payments[0]
	err = s.Reject(oldPayment.ID)
	if err != nil {
		t.Errorf("Repeat(): can't reject payment error=%v", err)
		return
	}

	//попробуем повторить 
	newPayment, err := s.Repeat(oldPayment.ID)
	if err != nil {
		t.Errorf("Repeat(): error=%v", err)
		return
	}

	if newPayment.AccountID != oldPayment.AccountID {
		t.Errorf("Repeat(): account ids of payments is difference. Repeated payment: %v, rejected payment: %v", newPayment, oldPayment)
		return
	}

	if newPayment.Amount != oldPayment.Amount {
		t.Errorf("Repeat(): amount of payments is difference. Repeated payment: %v, rejected payment: %v", newPayment, oldPayment)
		return
	}

	if newPayment.Category != oldPayment.Category {
		t.Errorf("Repeat(): category of payments is difference. Repeated payment: %v, rejected payment: %v", newPayment, oldPayment)
		return
	}
}

func TestService_PayFromFavorite_success(t *testing.T) {
	//создаём сервис
	s := newTestService()

	//регистриуем там пользователя
	_, _, favorites, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	
	//попробуем платить
	favorite := favorites[0] 
	payment, err := s.PayFromFavorite(favorite.ID)
	if err != nil {
		t.Errorf("PayFromFavorite(): error=%v", err)
		return
	}

	if payment.AccountID != favorite.AccountID {
		t.Errorf("PayFromFavorite(): account ids of payments is difference. Current payment: %v, favorite payment: %v", payment, favorite)
		return
	}

	if payment.Category != favorite.Category {
		t.Errorf("PayFromFavorite(): category of payments is difference. Current payment: %v, favorite payment: %v", payment, favorite)
		return
	}
}
////////////////////
func TestService_Export_success(t *testing.T) {
	srv := newTestService()

	//регистриуем там пользователя
	srv.addAccount(defaultTestAccount)

	err := srv.Export("../../files")

	if err != nil {
		t.Errorf("Export(): error=%v", err)
		return
	}
}

func TestService_Import_success(t *testing.T) {
//создаём сервис
srv := newTestService()

//регистриуем там пользователя
srv.addAccount(defaultTestAccount)

err := srv.Import("../../files")

if err != nil {
	t.Errorf("Import(): error=%v", err)
	return
}
}
/*
func TestService_SumPayments_fail(t *testing.T) {
	want := types.Money(5_000_00)
	//создаём сервис
	srv := newTestService()

	_, _, _, err := srv.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}

	got := srv.SumPayments(2)

	if want >= got {
		t.Errorf("SumPayments(): want: %v got: %v", want, got)
		return
	}

	if want >= got {
		t.Errorf("SumPayments(): want: %v got: %v", want, got)
		return
	}
}
*/
//////////////////////
func TestService_SumPayments_success(t *testing.T) {
	want := types.Money(4_000_00)
		//создаём сервис
		s := newTestService()

		//регистриуем там пользователя
		s.addAccount(defaultTestAccount)

		got := s.SumPayments(1)

		if want != got {
			t.Errorf("SumPayments(): want: %v got: %v", want, got)
			return
		}
}
/////////////////////////
func BenchmarkSumPayments(b *testing.B) {

	want := types.Money(4_000_00)
	srv := newTestService()
	gots := 1

	_, _, _, err := srv.addAccount(defaultTestAccount)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		result := srv.SumPayments(gots)
		if result != want {
			b.Fatalf("invalid result got %v, want %v", result, want)
		}
	}
}
/////////////////////////////////////////////////////////////////////////////
func TestService_SumPaymentsWithProgress(t *testing.T) {
	s := newTestService()
	for i := 0; i < 4_000_00; i++ {
		payment := &types.Payment{
			ID:     uuid.New().String(),
			Amount: types.Money(100),
		}
		s.payments = append(s.payments, payment)
	}

	s.SumPaymentsWithProgress()
}