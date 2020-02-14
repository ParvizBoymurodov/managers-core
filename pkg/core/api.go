package core

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrInvalidPass = errors.New("invalid password")

type QueryError struct { // alt + enter
	Query string
	Err   error
}

type DbError struct {
	Err error
}

type DbTxError struct {
	Err         error
	RollbackErr error
}

type Atm struct {
	Id int64
	Name string
	Street string
}

type Client struct {
	Id int64
	Name string
	Login string
	Password string
	Balance uint64
	BalanceNumber uint64
	PhoneNumber int64
}

type Services struct {
	Id int64
	Name string
	Balance uint64
}


func (receiver *QueryError) Unwrap() error {
	return receiver.Err
}

func (receiver *QueryError) Error() string {
	return fmt.Sprintf("can't execute query %s: %s", loginSQL, receiver.Err.Error())
}

func queryError(query string, err error) *QueryError {
	return &QueryError{Query: query, Err: err}
}

func (receiver *DbError) Error() string {
	return fmt.Sprintf("can't handle db operation: %v", receiver.Err.Error())
}

func (receiver *DbError) Unwrap() error {
	return receiver.Err
}

func dbError(err error) *DbError {
	return &DbError{Err: err}
}


func Init(db *sql.DB) (err error) {
	ddls := []string{managersDDL, atmDDL,clientDDL,servicesDDL}
	for _, ddl := range ddls {
		_, err = db.Exec(ddl)
		if err != nil {
			return err
		}
	}

	initialData := []string{managersInitialData}
	for _, datum := range initialData {
		_, err = db.Exec(datum)
		if err != nil {
			return err
		}
	}

	return nil
}

func Login(login, password string, db *sql.DB) (int64 ,bool, error) {
	var dbLogin, dbPassword string
    var dbId int64
	err := db.QueryRow(
		LoginForClient,
		login).Scan(&dbId,&dbLogin, &dbPassword)

	if err != nil {
		if err == sql.ErrNoRows {
			return -1, false, nil
		}

		return -1,false, queryError(LoginForClient, err)
	}

	if dbPassword != password {
		return -1,false, ErrInvalidPass
	}

	return dbId ,true, nil
}


func LoginForManagers(login, password string, db *sql.DB) (bool, error) {
	var dbLogin, dbPassword string

	err := db.QueryRow(
		loginSQL,
		login).Scan(&dbLogin, &dbPassword)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}

		return false, queryError(loginSQL, err)
	}

	if dbPassword != password {
		return false, ErrInvalidPass
	}

	return true, nil
}

func GetAllAtms(db *sql.DB) (atms []Atm, err error) {
	rows, err := db.Query(getAllAtmSql)
	if err != nil {
		return nil, queryError(getAllAtmSql, err)
	}
	defer func() {
		if innerErr := rows.Close(); innerErr != nil {
			atms, err = nil, dbError(innerErr)
		}
	}()

	for rows.Next() {
		atm := Atm{}
		err = rows.Scan(&atm.Id, &atm.Name, &atm.Street)
		if err != nil {
			return nil, dbError(err)
		}
		atms = append(atms, atm)
	}
	if rows.Err() != nil {
		return nil, dbError(rows.Err())
	}

	return atms, nil
}

func GetBalanceList(db *sql.DB,user_id int64) (listBalance []Client, err error) {
	rows, err := db.Query(getListBalanceSql, user_id )
	if err != nil {
		return nil, queryError(getListBalanceSql, err)
	}
	defer func() {
		if innerErr := rows.Close(); innerErr != nil {
			listBalance, err = nil, dbError(innerErr)
		}
	}()

	for rows.Next() {
		listAccount := Client{}
		err = rows.Scan(&listAccount.Id, &listAccount.Name, &listAccount.BalanceNumber, &listAccount.Balance )
		if err != nil {
			return nil, dbError(err)
		}
		listBalance = append(listBalance,listAccount)
	}
	if rows.Err() != nil {
		return nil, dbError(rows.Err())
	}

	return listBalance, nil
}

func GetServices(db *sql.DB)(ServiceList []Services,err error)  {
	rows, err := db.Query(getAllServices)
	if err != nil {
		return nil, queryError(getAllServices, err)
	}
	defer func() {
		if innerErr := rows.Close(); innerErr != nil {
			ServiceList, err = nil, dbError(innerErr)
		}
	}()

	for rows.Next() {
		listService := Services{}
		err = rows.Scan(&listService.Id, &listService.Name, &listService.Balance)
		if err != nil {
			return nil, dbError(err)
		}
		ServiceList = append(ServiceList,listService)
	}
	if rows.Err() != nil {
		return nil, dbError(rows.Err())
	}

	return ServiceList, nil
}

func AddClients(client Client, db *sql.DB) (err error) {

	_, err = db.Exec(
		insertClientSQL,
		sql.Named("name", client.Name),
		sql.Named("login", client.Login),
		sql.Named("password", client.Password),
		sql.Named("balance",client.Balance),
		sql.Named("balance_number",client.BalanceNumber),
		sql.Named("phone_number",client.PhoneNumber),
	)
	if err != nil {
		return err
	}

	return nil
}

func AddAtm (atm Atm, db *sql.DB)(err error){

		_, err = db.Exec(
			insertAtmSql,
			sql.Named("name", atm.Name),
			sql.Named("street", atm.Street),
		)
		if err != nil {
			return err
		}

		return nil
}

func AddServices(services Services,db *sql.DB)(err error)  {

	_, err = db.Exec(
		insertServices,
		sql.Named("name", services.Name),
		sql.Named("balance", services.Balance),
	)
	if err != nil {
		return err
	}

	return nil
}

func UpdateBalance(listBalance Client,  db *sql.DB) (err error) {

	_, err = db.Exec(
		updateCardBalanceSQL,
		sql.Named("login", listBalance.Login),
		sql.Named("balance", listBalance.Balance),
	)
	if err != nil {
		return err
	}

	return nil
}

func transactionByPhoneNumberPlus(transaction Client, tx *sql.Tx) (err error) {
	_, err = tx.Exec(
		updateTransactionWithPhoneNumberPlus,
		sql.Named("phone_number",transaction.PhoneNumber ),
		sql.Named("balance", transaction.Balance ),
	)
	if err != nil {
		return err
	}

	return nil
}

func transactionByPhoneNumberMinus(balanceNumber uint64,balance uint64,tx *sql.Tx) (err error) {
	_, err = tx.Exec(
		updateTransactionWithPhoneNumberMinus,
		sql.Named("balance_number", balanceNumber),
		sql.Named("balance", balance),
	)
	if err != nil {
		return err
	}

	return nil
}

func transactionBalanceNumberPlus(transaction Client, tx *sql.Tx) (err error) {

	_, err = tx.Exec(
		updateTransactionWithBalanceNumberPlus,
		sql.Named("balance_number", transaction.BalanceNumber),
		sql.Named("balance", transaction.Balance),
	)
	if err != nil {
		return err
	}

	return nil
}

func transactionBalanceNumberMinus(myBalanceNumber uint64,balance uint64, tx *sql.Tx) (err error) {

	_, err = tx.Exec(
		updateTransactionWithBalanceNumberMinus,
		sql.Named("balance_number", myBalanceNumber),
		sql.Named("balance", balance),
	)
	if err != nil {
		return err
	}

	return nil
}

func servicePaying(pay Services, tx *sql.Tx) (err error) {
	_, err = tx.Exec(
		updateServices,
		sql.Named("id", pay.Id),
		sql.Named("balance", pay.Balance),
	)
	if err != nil {
		return err
	}

	return nil
}

func repay(balanceNumber uint64,balance uint64,tx *sql.Tx) (err error)  {
	_, err = tx.Exec(
		payServices,
		sql.Named("balance_number", balanceNumber),
		sql.Named("balance", balance),
	)
	if err != nil {
		return err
	}

	return nil
}

func CheckByBalanceNumber(balanceNumber uint64, db *sql.DB)(err error)  {
	var id int
	err = db.QueryRow("select id from client where balance_number=?", balanceNumber).Scan(&id)
	return err
}

func CheckByPhoneNumber(phoneNumber int64,db *sql.DB) (err error) {
	var id int
	err = db.QueryRow("select id from client where phone_number=?", phoneNumber).Scan(&id)
	return err
}

func CheckId(id int64,db *sql.DB) (err error) {
	var name int
	err = db.QueryRow("select id from services where id=?", id).Scan(&name)
	return err
}

func TransferByPhoneNumber(balanceNumber uint64,balance uint64,tranzaction Client, db *sql.DB)(err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	err = transactionByPhoneNumberMinus(balanceNumber,balance,tx)
	if err != nil {
		return err
	}
	err = transactionByPhoneNumberPlus(tranzaction,tx)
	if err != nil {
		return err
	}
  return nil
}

func TransferByBalanceNumber(myBalanceNumber uint64,balance uint64,tranzaction Client, db *sql.DB)(err error)  {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	err = transactionBalanceNumberMinus(myBalanceNumber,balance,tx)
	if err != nil {
		return err
	}
	err = transactionBalanceNumberPlus(tranzaction,tx)
	if err != nil {
		return err
	}
	return nil
}

func PayForServices(balanceNumber uint64,balance uint64,pay Services, db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	err = repay(balanceNumber ,balance ,tx)
	if err != nil {
		return err
	}
	err = servicePaying(pay,tx)
	if err != nil {
		return err
	}
	return nil
}