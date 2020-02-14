package core

const managersDDL = `
CREATE TABLE IF NOT EXISTS managers
(
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    name    TEXT    NOT NULL,
    login   TEXT    NOT NULL UNIQUE,
    password TEXT NOT NULL
);`

const managersInitialData = `INSERT INTO managers
VALUES (1, 'Vasya', 'vasya', 'secret'),
       (2, 'Petya', 'petya', '1212'),
       (3, 'Vanya', 'vanya', '1313'),
       (4, 'Masha', 'masha', '1414'),
       (5, 'Dasha', 'dasha', '1515'),
       (6, 'Sasha', 'sasha', '1616')
       ON CONFLICT DO NOTHING;`

const clientDDL = `
create table if not exists client
(
id integer primary key autoincrement,
name text not null,
login text not null unique,
password text not null,
balance integer not null check(balance>=0),
balance_number integer not null unique,
phone_number integer not null unique
);`

const atmDDL = `
create table if not exists atm (
id  integer primary key autoincrement,
name text not null,
street text not null
);`

const servicesDDL = `
create table if not exists services(
id integer primary key autoincrement,
name text not null,
price integer not null check (price>0)
);`

const getAllAtmSql = `select id,name,street from atm;`
const loginSQL = `SELECT login, password FROM managers WHERE login = ?`
const insertClientSQL = `INSERT INTO client(name, login, password, balance, balance_number, phone_number) values (:name, :login, :password, :balance, :balance_number, :phone_number);`
const LoginForClient = `select id, login,password from client where login = ?`
const insertAtmSql = `insert into atm (name,street) values (:name, :street);`
const insertServices = `insert into services(name, price) values(:name, :price);`
const getAllServices = `select id,name,price from services;`
const getListBalanceSql = `select id, name, balance_number, balance from client where id = ?;`
const updateCardBalanceSQL = ` UPDATE client SET balance = balance + :balance WHERE login = :login;`
const 	updateTransactionWithPhoneNumberMinus = `UPDATE client SET balance = balance - :balance WHERE phone_number = :phone_number;`
const updateTransactionWithPhoneNumberPlus = `UPDATE client SET balance = balance + :balance where phone_number = :phone_number;`
const updateTransactionWithBalanceNumberMinus = `UPDATE client SET balance = balance - :balance WHERE balance_number = :balance_number;`
const updateTransactionWithBalanceNumberPlus = `UPDATE client SET balance = balance + :balance where balance_number = :balance_number;`