package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func iniciarAgendadorMeiaNoite() {
	// 1. Forçar fuso horário de Brasília
	loc, _ := time.LoadLocation("America/Sao_Paulo")

	for {
		agora := time.Now().In(loc)

		// Calcula a próxima meia-noite no horário de Brasília
		proximaMeiaNoite := time.Date(agora.Year(), agora.Month(), agora.Day()+1, 0, 1, 0, 0, loc)
		tempoRestante := time.Until(proximaMeiaNoite)

		log.Printf("Aguardando %v para a próxima virada de dia...", tempoRestante)
		time.Sleep(tempoRestante)

		// Dispara a atualização
		message := &messaging.Message{
			Data:    map[string]string{"update_widget": "true"},
			Topic:   "casal",
			Android: &messaging.AndroidConfig{Priority: "high"},
		}

		// Usamos Background context aqui
		_, err := msgClient.Send(context.Background(), message)
		if err != nil {
			log.Printf("Erro no agendador: %v", err)
		} else {
			log.Println("Sinal de virada de dia enviado com sucesso!")
		}

		// Evita disparos múltiplos no mesmo segundo
		time.Sleep(2 * time.Second)
	}
}

var msgClient *messaging.Client

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Arquivo .env não encontrado, lendo variáveis de ambiente")
	}
	ctx := context.Background()

	// 1. Lê a string do JSON da variável de ambiente
	serviceAccountJSON := os.Getenv("FIREBASE_SERVICE_ACCOUNT_KEY")
	if serviceAccountJSON == "" {
		log.Fatal("ERRO: Variável FIREBASE_SERVICE_ACCOUNT_KEY não configurada!")
	}

	// 2. Inicializa o Firebase usando os bytes do JSON
	opt := option.WithCredentialsJSON([]byte(serviceAccountJSON))
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("Erro ao iniciar Firebase: %v", err)
	}

	// Inicializa o cliente de mensagens
	msgClient, err = app.Messaging(ctx)
	if err != nil {
		log.Fatalf("Erro ao obter cliente de mensagens: %v", err)
	}

	// 2. Inicia o agendador de virada de dia em background
	go iniciarAgendadorMeiaNoite()

	r := gin.Default()

	// Rota que o app vai chamar ao responder
	r.POST("/notificar-atualizacao", func(c *gin.Context) {
		// Mensagem silenciosa para o tópico "casal"
		message := &messaging.Message{
			Data: map[string]string{
				"update_widget": "true",
			},
			Topic: "casal",
			Android: &messaging.AndroidConfig{
				Priority: "high",
			},
		}

		response, err := msgClient.Send(ctx, message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		fmt.Println("Sinal enviado com sucesso:", response)
		c.JSON(http.StatusOK, gin.H{"message": "Widget do parceiro será atualizado!"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Porta padrão se rodar localmente
	}
	r.Run(":" + port)
}
