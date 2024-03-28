package service

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	_brasilAPIURL  = "https://brasilapi.com.br/api/cep/v1/"
	_viaCepAPIURL  = "http://viacep.com.br/ws/"
	_openCepAPIURL = "http://opencep.com/v1/"

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
	wg := sync.WaitGroup{}
	wg.Add(3)

	channel := make(chan string)

	var openCEPAPI openCEPAPI
	var viaCEPAPI viaCEP
	var brasilAPI brasiAPI
	var respBrasilAPI *http.Response
	var respOpenCEPAPI *http.Response
	var respViaCepAPI *http.Response

	cep := c.GetHeader("cep")

	ctx, cancel := context.WithTimeout(c, _requestTimeout)
	defer cancel()

	brasiAPIURL := _brasilAPIURL + cep
	openCEPAPIURL := _openCepAPIURL + cep
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

	reqOpenCEPAPI, err := http.NewRequestWithContext(ctx, http.MethodGet, openCEPAPIURL, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"err": err,
		})
	}

	go func() {
		start := time.Now()
		respBrasilAPI, err = http.DefaultClient.Do(reqBrasilAPI)
		if err != nil {
			select {
			case <-ctx.Done():
				log.Print("Brasil API - tempo de contexto excedido")
				channel <- "Error"
				wg.Done()
			default:
				log.Printf("Brasil API - erro na solicitação. err -> %v", err)
				channel <- "Error"
				wg.Done()
			}
		}
		wg.Done()
		channel <- "Brasil API"
		elapsed := time.Since(start)
		log.Printf("Brasil API - %s", elapsed)
	}()

	go func() {
		start := time.Now()
		respOpenCEPAPI, err = http.DefaultClient.Do(reqOpenCEPAPI)
		if err != nil {
			select {
			case <-ctx.Done():
				log.Print("Open CEP API - tempo de contexto excedido")
				wg.Done()
				channel <- "Error"
			default:
				log.Printf("Open CEP API - erro na solicitação. err -> %v", err)
				wg.Done()
				channel <- "Error"
			}
		}
		wg.Done()
		channel <- "Open CEP API"
		elapsed := time.Since(start)
		log.Printf("Open CEP API - %s", elapsed)
	}()

	go func() {
		start := time.Now()
		respViaCepAPI, err = http.DefaultClient.Do(reqViaCepAPI)
		if err != nil {
			select {
			case <-ctx.Done():
				log.Print("Via CEP - tempo de contexto excedido")
				wg.Done()
				channel <- "Error"
			default:
				log.Printf("Via CEP API - erro na solicitação. err -> %v", err)
				wg.Done()
				channel <- "Error"
			}
		}
		wg.Done()
		channel <- "Via CEP API"
		elapsed := time.Since(start)
		log.Printf("Via CEP - %s", elapsed)
	}()

	wg.Wait()
	API := <-channel

	switch API {
	case "Via CEP API":
		json.NewDecoder(respViaCepAPI.Body).Decode(&viaCEPAPI)
		viaCEPAPI.API = "Via CEP API"
		c.JSON(http.StatusOK, viaCEPAPI)
	case "Brasil API":
		json.NewDecoder(respBrasilAPI.Body).Decode(&brasilAPI)
		brasilAPI.API = "Brasil API"
		c.JSON(http.StatusOK, brasilAPI)
	case "Open CEP API":
		json.NewDecoder(respOpenCEPAPI.Body).Decode(&openCEPAPI)
		openCEPAPI.API = "Open CEP API"
		c.JSON(http.StatusOK, openCEPAPI)
	default:
		c.JSON(http.StatusInternalServerError, "erro ao realizar a chamada API")
	}
}
