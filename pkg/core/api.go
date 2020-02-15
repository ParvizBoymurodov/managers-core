package core

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
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
	Address string
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

func Login(login, password string, db *sql.DB) (int64,bool, error) {
	var dbLogin, dbPassword string
    var dbId int64
	err := db.QueryRow(
		LoginForClient,
		login).Scan(&dbId,&dbLogin, &dbPassword)

	if err != nil {
		if err == sql.ErrNoRows {
			return -1, false, nil
		}

		return -1, false, queryError(LoginForClient, err)
	}

	if dbPassword != password {
		return -1, false, ErrInvalidPass
	}

	return dbId, true, nil
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
		err = rows.Scan(&atm.Id, &atm.Name, &atm.Address)
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
		err = rows.Scan(&listService.Id, &listService.Name)
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
			sql.Named("street", atm.Address),
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


func ExportClientsToJSON(db *sql.DB) error {
	return ExportToFile(db, getAllClientsDataSQL, "clients.json",
		mapRowToClient, json.Marshal, mapInterfaceSliceToClients)
}
func ExportAtmsToJSON(db *sql.DB) error {
	return ExportToFile(db, getAllAtmDataSQL, "atms.json",
		mapRowToAtm, json.Marshal,
		mapInterfaceSliceToAtms)
}

//XML

func ExportClientsToXML(db *sql.DB) error {
	return ExportToFile(db, getAllClientsDataSQL, "clients.xml",
		mapRowToClient, xml.Marshal, mapInterfaceSliceToClients)
}
func ExportAtmsToXML(db *sql.DB) error {
	return ExportToFile(db, getAllAtmDataSQL, "atms.xml",
		mapRowToAtm, xml.Marshal,
		mapInterfaceSliceToAtms)
}

func mapRowToClient(rows *sql.Rows) (interface{}, error) {
	client := Client{}
	err := rows.Scan(&client.Id, &client.Login, &client.Password,
		&client.Name, &client.PhoneNumber, &client.Balance, &client.BalanceNumber)
	if err != nil {
		return nil, err
	}
	return client, nil
}
func mapRowToAtm(rows *sql.Rows) (interface{}, error) {
	atm := Atm{}
	err := rows.Scan(&atm.Id,&atm.Name, &atm.Address)
	if err != nil {
		return nil, err
	}
	return atm, nil
}
type ClientsExport struct {
	Clients []Client
}
func mapInterfaceSliceToClients(ifaces []interface{}) interface{} {
	clients := make([]Client, len(ifaces))
	for i := range ifaces {
		clients[i] = ifaces[i].(Client)
	}
	clientsExport := ClientsExport{Clients: clients}
	return clientsExport
}
func mapInterfaceSliceToAtms(ifaces []interface{}) interface{} {
	atms := make([]Atm, len(ifaces))
	for i := range ifaces {
		atms[i] = ifaces[i].(Atm)
	}
	atmsExport := AtmsExport{Atms: atms}
	return atmsExport
}
func ImportClientsFromJSON(db *sql.DB) error {
	return ImportFromFile(
		db,
		"clients.json",
		func(data []byte) ([]interface{}, error) {
			return mapBytesToClients(data, json.Unmarshal)
		},
		insertClientToDB,
	)
}
func ImportAtmsFromJSON(db *sql.DB) error {
	return ImportFromFile(
		db,
		"atms.json",
		func(data []byte) ([]interface{}, error) {
			return mapBytesToAtms(data, json.Unmarshal)
		},
		insertAtmToDB,
	)
}
func ImportClientsFromXML(db *sql.DB) error {
	return ImportFromFile(
		db,
		"clients.xml",
		func(data []byte) ([]interface{}, error) {
			return mapBytesToClients(data, xml.Unmarshal)
		},
		insertClientToDB,
	)
}
func ImportAtmsFromXML(db *sql.DB) error {
	return ImportFromFile(
		db,
		"atms.xml",
		func(data []byte) ([]interface{}, error) {
			return mapBytesToAtms(data, xml.Unmarshal)
		},
		insertAtmToDB,
	)
}
func mapBytesToClients(data []byte,
	unmarshal func([]byte, interface{}) error,
) ([]interface{}, error) {
	clientsExport := ClientsExport{}
	err := unmarshal(data, &clientsExport)
	if err != nil {
		return nil, err
	}
	ifaces := make([]interface{}, len(clientsExport.Clients))
	for index := range ifaces {
		ifaces[index] = clientsExport.Clients[index]
	}
	return ifaces, nil
}
func insertClientToDB(iface interface{}, db *sql.DB) error {
	client := iface.(Client)
	_, err := db.Exec(
		insertClientSQL,
		sql.Named("id", client.Id),
		sql.Named("name", client.Name),
		sql.Named("login", client.Login),
		sql.Named("password", client.Password),
		sql.Named("phone", client.PhoneNumber),
		sql.Named("balance_number", client.BalanceNumber),
		sql.Named("balance", client.Balance),
	)
	if err != nil {
		return err
	}
	return nil
}

type ATM struct {
	id int64
	name string
	address string
}

type AtmsExport struct {
	Atms []Atm
}



func mapBytesToAtms(data []byte,
	unmarshal func([]byte, interface{}) error,
) ([]interface{}, error) {
	atmsExport := AtmsExport{}
	err := unmarshal(data, &atmsExport)
	if err != nil {
		return nil, err
	}
	ifaces := make([]interface{}, len(atmsExport.Atms))
	for index := range ifaces {
		ifaces[index] = atmsExport.Atms[index]
	}
	return ifaces, nil
}
func insertAtmToDB(iface interface{}, db *sql.DB) error {
	atm := iface.(Atm)
	_, err := db.Exec(
		insertAtmSql,
		sql.Named("id", atm.Id),
		sql.Named("name", atm.Name),
		sql.Named("street", atm.Address),
	)
	if err != nil {
		return err
	}
	return nil
}

type MapperRowTo func(rows *sql.Rows) (interface{}, error)
type MapperInterfaceSliceTo func([]interface{}) interface{}
type Marshaller func(interface{}) ([]byte, error)

func ExportToFile(
	db *sql.DB,
	getDataFromDbSQL string,
	filename string,
	mapRow MapperRowTo,
	marshal Marshaller,
	mapDataSlice MapperInterfaceSliceTo) error {

	rows, err := db.Query(getDataFromDbSQL)
	if err != nil {
		return err
	}
	defer func() {
		err = rows.Close()
	}()
	var dataSlice []interface{}
	for rows.Next() {
		dataElement, err := mapRow(rows)
		if err != nil {
			return err
		}
		dataSlice = append(dataSlice, dataElement)
	}
	exportData := mapDataSlice(dataSlice)
	data, err := marshal(exportData)
	err = ioutil.WriteFile(filename, data, 0666)
	if err != nil {
		return err
	}
	return nil
}


type MapperBytesTo func([]byte) ([]interface{}, error)

func ImportFromFile(
	db *sql.DB,
	filename string,
	mapBytes MapperBytesTo,
	insertToDB func(interface{}, *sql.DB) error,
) error {
	itemsData, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	sliceData, err := mapBytes(itemsData)

	for _, datum := range sliceData {
		err = insertToDB(datum, db)
		if err != nil {
			return err
		}
	}

	return nil
}