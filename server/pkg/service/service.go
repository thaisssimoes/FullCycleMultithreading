package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

const (
	_brasilAPIURL = "https://brasilapi.com.br/api/cep/v1/"
	_viaCepAPIURL = "http://viacep.com.br/ws/"

	_requestTimeout = 1 * time.Second
)

func App() {
	s := gin.Default()
	rotas(s)
	log.Fatalln(s.Run(":8080"))
}

func rotas(s *gin.Engine) {
	s.GET("/cep", getCEP)
}

func getCEP(c *gin.Context) {
	var viaCEPAPI viaCEP
	var brasilAPI brasiAPI

	cep := c.GetHeader("cep")

	ctx, cancel := context.WithTimeout(c, _requestTimeout)
	defer cancel()

	brasiAPIURL := _brasilAPIURL + cep
	viaCepAPIURL := _viaCepAPIURL + cep + "/json"

	reqBrasilAPI, err := http.NewRequestWithContext(ctx, http.MethodGet, brasiAPIURL, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": err,
		})
	}

	reqViaCepAPI, err := http.NewRequestWithContext(ctx, http.MethodGet, viaCepAPIURL, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": err,
		})
	}

	respBrasilAPI, err := http.DefaultClient.Do(reqBrasilAPI)
	if err != nil {
		select {
		case <-ctx.Done():
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "tempo de contexto excedido",
				"err":     err,
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"err": err,
			})
		}

	}

	respViaCepAPI, err := http.DefaultClient.Do(reqViaCepAPI)
	if err != nil {
		select {
		case <-ctx.Done():
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "tempo de contexto excedido",
				"err":     err,
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"err": err,
			})
		}
	}

	json.NewDecoder(respBrasilAPI.Body).Decode(&brasilAPI)
	json.NewDecoder(respViaCepAPI.Body).Decode(&viaCEPAPI)

	resultado := fmt.Sprintf("ViaCEP: %v \n\n BrasilAPI: %v", viaCEPAPI, brasilAPI)

	c.JSON(http.StatusOK, resultado)

}
