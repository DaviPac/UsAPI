package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

var msgClient *messaging.Client

func main() {
	ctx := context.Background()

	// Inicializa o Firebase Admin
	opt := option.WithCredentialsFile("serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("Erro ao iniciar Firebase: %v", err)
	}

	msgClient, _ = app.Messaging(ctx)

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
