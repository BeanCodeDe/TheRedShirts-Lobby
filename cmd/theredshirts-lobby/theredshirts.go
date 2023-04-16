package main

import (
	"github.com/BeanCodeDe/TheRedShirts-Lobby/internal/app/theredshirts/api"
	log "github.com/sirupsen/logrus"
)

func main() {
	_, err := api.NewApi()
	if err != nil {
		log.Fatal("Error while starting api: ", err)
	}
}
