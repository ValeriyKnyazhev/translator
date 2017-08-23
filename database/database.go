package database

import (
	"database/sql"
	"fmt"
	"github.com/ValeriyKnyazhev/translator/configuration"
	_ "github.com/lib/pq"
	"log"
	"time"
)

type DBManager interface {
	CreateTable() error
	GetData(uuid string) (*Data, error)
	SetData(d *Data) error
	UpdateData(d *Data) error
}

type dbmanager struct {
	db *sql.DB
}

var Manager DBManager

func init() {
	dbconfig, err := configuration.ReadDBConfig()
	if err != nil {
		log.Fatal("Error: Can't read database config")
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbconfig.Host, dbconfig.Port, dbconfig.User, dbconfig.Password, dbconfig.DBname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal("Error: The data source arguments are not valid")
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("Error: Could not establish a connection with the database")
	}
	Manager = &dbmanager{db: db}
}

var createTableStatment string = `
	CREATE TABLE IF NOT EXISTS tasks (
		id             UUID        CONSTRAINT uuid PRIMARY KEY,
		timestamp      TIMESTAMP   DEFAULT (NOW()),
		userId         integer     NOT NULL,
		pictureUrl     text        NOT NULL,
		recognizedText text,
		recognizedLang varchar(3),
		checkedText    text,
		translatedText text,
		translatedLang varchar(3),
		error          text
	);
	CREATE INDEX ON tasks (userId);`

func (mgr *dbmanager) CreateTable() (err error) {
	_, err = mgr.db.Exec(createTableStatment)
	if err != nil {
		log.Println("can't create table in database")
		return err
	}
	return nil
}

type Data struct {
	Id             string
	Timestamp      time.Time
	UserId         int
	PictureUrl     string
	RecognizedText string
	RecognizedLang string
	CheckedText    string
	TranslatedText string
	TranslatedLang string
	Error          string
}

var getStatment string = `SELECT * FROM tasks WHERE id=$1`

func (mgr *dbmanager) GetData(uuid string) (*Data, error) {
	row := mgr.db.QueryRow(getStatment, uuid)
	d := &Data{}
	switch err := row.Scan(&d.Id, &d.Timestamp, &d.UserId, &d.PictureUrl,
		&d.RecognizedText, &d.RecognizedLang, &d.CheckedText,
		&d.TranslatedText, &d.TranslatedLang, &d.Error); err {
	case sql.ErrNoRows:
		log.Println("No rows were returned!")
		return nil, nil
	case nil:
		return d, nil
	default:
		return nil, err
	}
}

var setStatment string = `INSERT INTO tasks (
	id,
	userId,
	pictureUrl,
	recognizedText,
	recognizedLang,
	checkedText,
	translatedText,
	translatedLang,
	error) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

func (mgr *dbmanager) SetData(d *Data) error {
	_, err := mgr.db.Exec(setStatment, d.Id, d.UserId, d.PictureUrl,
		d.RecognizedText, d.RecognizedLang, d.CheckedText,
		d.TranslatedText, d.TranslatedLang, d.Error)
	if err != nil {
		log.Println("can't set data: ", err)
		return err
	}
	return err
}

var updateStatment string = `UPDATE tasks SET (
	userId,
	pictureUrl,
	recognizedText,
	recognizedLang,
	checkedText,
	translatedText,
	translatedLang,
	error) = ($1, $2, $3, $4, $5, $6, $7, $8)`

func (mgr *dbmanager) UpdateData(d *Data) error {
	_, err := mgr.db.Exec(updateStatment, d.UserId, d.PictureUrl,
		d.RecognizedText, d.RecognizedLang, d.CheckedText,
		d.TranslatedText, d.TranslatedLang, d.Error)
	if err != nil {
		log.Println("can't update data: ", err)
		return err
	}
	return err
}
