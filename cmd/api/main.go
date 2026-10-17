package main

import (
	"context"
	"jungle-gaming-backend-go/internal/auth"
	"jungle-gaming-backend-go/internal/database"
	"jungle-gaming-backend-go/internal/httpapi"
	"jungle-gaming-backend-go/internal/repository"
	"jungle-gaming-backend-go/internal/service"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	const issuer = "http://localhost:8081/realms/jungle-gaming"
	const audience = "jungle-backend-api"

	validator, err := auth.NewKeycloakValidator(
		issuer,
		audience,
	)
	if err != nil {
		log.Fatal("Erro ao configurar autenticacao: ", err)
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL nao configurada")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	db, err := database.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatal("Erro ao conectar PostgreSQL: ", err)
	}
	defer db.Close()

	log.Println("PostgreSQL conectado com sucesso")

	walletRepository := repository.NewWalletRepository(db)
	betService := service.NewBetService(db)

	router := httpapi.NewRouter(
		validator,
		walletRepository,
		betService,
	)

	log.Println("Servidor iniciado na porta 8080")

	err = http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal(err)
	}
}
