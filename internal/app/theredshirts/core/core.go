package core

import "github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/db"

type (

	//Facade
	CoreFacade struct {
		db db.DB
	}

	Core interface {
	}
)

func NewCore() (Core, error) {
	db, err := db.NewConnection()
	if err != nil {
		return nil, err
	}
	return &CoreFacade{db: db}, nil
}
