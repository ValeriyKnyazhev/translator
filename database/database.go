package database

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"time"
)

type Dbmanager struct {
	db *sql.DB
}

const (
	TaskErrNone        = "none"
	TaskStatusRun      = "Run"
	TaskStatusComplete = "Complete"
	TaskStatusStop     = "Stop"
)

func CreateDB(Host string, Port int, User string, Password string, DBname string) (*Dbmanager, error) {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		Host, Port, User, Password, DBname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Println("Error: The data source arguments are not valid: ", err)
		err = errors.New(fmt.Sprintln("Error: The data source arguments are not valid: ", err))
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		log.Println("Error: Could not establish a connection with the database: ", err)
		err = errors.New(fmt.Sprintln("Error: Could not establish a connection with the database: ", err))
		return nil, err
	}
	return &Dbmanager{db: db}, nil
}

var createTableStatment string = `
	CREATE TABLE IF NOT EXISTS tasks (
		id             UUID        CONSTRAINT uuid PRIMARY KEY,
		timestamp      TIMESTAMP   DEFAULT (NOW()),
		userId         integer     NOT NULL,
		currTaskId     integer     NOT NULL,
		pictureUrl     text        NOT NULL,
		recognizedText text,
		recognizedLang varchar(3),
		checkedText    text,
		translatedText text,
		translatedLang varchar(3),
		status         text        NOT NULL,
		error          text        NOT NULL
	);
	CREATE INDEX ON tasks (userId);`

func (mgr *Dbmanager) CreateTable() (err error) {
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
	CurrTaskId     int
	PictureUrl     string
	RecognizedText string
	RecognizedLang string
	CheckedText    string
	TranslatedText string
	TranslatedLang string
	Status         string
	Error          string
}

var getStatment string = `SELECT * FROM tasks WHERE id=$1`

func (mgr *Dbmanager) GetData(uuid string) (*Data, error) {
	row := mgr.db.QueryRow(getStatment, uuid)
	d := &Data{}
	switch err := row.Scan(&d.Id, &d.Timestamp, &d.UserId, &d.CurrTaskId, &d.PictureUrl,
		&d.RecognizedText, &d.RecognizedLang, &d.CheckedText,
		&d.TranslatedText, &d.TranslatedLang, &d.Status, &d.Error); err {
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
	currTaskId,
	pictureUrl,
	recognizedText,
	recognizedLang,
	checkedText,
	translatedText,
	translatedLang,
	status,
	error) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

func (mgr *Dbmanager) SetData(d *Data) error {
	_, err := mgr.db.Exec(setStatment, d.Id, d.UserId, d.CurrTaskId, d.PictureUrl,
		d.RecognizedText, d.RecognizedLang, d.CheckedText,
		d.TranslatedText, d.TranslatedLang, d.Status, d.Error)
	if err != nil {
		log.Println("can't set data: ", err)
		return err
	}
	return err
}

var updateStatment string = `UPDATE tasks SET (
	userId,
	currTaskId,
	pictureUrl,
	recognizedText,
	recognizedLang,
	checkedText,
	translatedText,
	translatedLang,
	status,
	error) = ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

func (mgr *Dbmanager) UpdateData(d *Data) error {
	_, err := mgr.db.Exec(updateStatment, d.UserId, d.CurrTaskId, d.PictureUrl,
		d.RecognizedText, d.RecognizedLang, d.CheckedText,
		d.TranslatedText, d.TranslatedLang, d.Status, d.Error)
	if err != nil {
		log.Println("can't update data: ", err)
		return err
	}
	return err
}
