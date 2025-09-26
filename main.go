package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// Variáveis globais para os URLs das APIs
var (
	viacepURL     = "https://viacep.com.br/ws/%s/json/"
	weatherAPIURL = "http://api.weatherapi.com/v1/current.json?key=%s&q=%s"
)

// WeatherResponse struct para a API de clima
type WeatherResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
}

// CepResponse struct para a API de CEP do ViaCEP
type CepResponse struct {
	Localidade string `json:"localidade"`
	Erro       bool   `json:"erro"`
}

// FinalResponse struct para o corpo da resposta final
type FinalResponse struct {
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

// Manipulador da rota que busca o CEP e o clima
func getCepWeather(w http.ResponseWriter, r *http.Request, weatherAPIKey string) {
	// Extrai o CEP da URL
	cep := strings.TrimPrefix(r.URL.Path, "/cep/")
	if len(cep) != 8 {
		http.Error(w, "invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	// Busca a cidade usando a API ViaCEP
	resp, err := http.Get(fmt.Sprintf(viacepURL, cep))
	if err != nil {
		log.Println("Erro ao buscar CEP:", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var cepData CepResponse
	if err := json.Unmarshal(body, &cepData); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Verifica se o CEP é inválido com base na resposta do ViaCEP
	if cepData.Erro {
		http.Error(w, "can not find zipcode", http.StatusNotFound)
		return
	}

	// Codifica o nome da localidade para uso na URL
	encodedLocation := url.QueryEscape(cepData.Localidade)

	// Cria uma nova requisição para a WeatherAPI com o header User-Agent
	req, err := http.NewRequest("GET", fmt.Sprintf(weatherAPIURL, weatherAPIKey, encodedLocation), nil)
	if err != nil {
		log.Println("Erro ao criar a requisição para o clima:", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	// Adiciona o cabeçalho User-Agent para evitar bloqueio da API
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")

	// Executa a requisição com o cliente padrão
	client := &http.Client{}
	weatherResp, err := client.Do(req)
	if err != nil {
		log.Println("Erro ao buscar clima:", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	defer weatherResp.Body.Close()

	// Verifica o status da resposta da WeatherAPI
	if weatherResp.StatusCode != http.StatusOK {
		log.Printf("Erro na API de clima: %s", weatherResp.Status)
		http.Error(w, "weather API error", http.StatusFailedDependency)
		return
	}

	weatherBody, err := io.ReadAll(weatherResp.Body)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var weatherData WeatherResponse
	if err := json.Unmarshal(weatherBody, &weatherData); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Calcula as temperaturas em Fahrenheit e Kelvin
	tempC := weatherData.Current.TempC
	tempF := tempC*1.8 + 32
	tempK := tempC + 273.15

	// Cria a struct final com as temperaturas
	finalResponse := FinalResponse{
		TempC: tempC,
		TempF: tempF,
		TempK: tempK,
	}

	// Converte a struct para JSON
	responseJSON, err := json.Marshal(finalResponse)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Adiciona a quebra de linha ao final do JSON
	responseWithNewline := append(responseJSON, '\n')

	// Retorna a resposta JSON com a quebra de linha para o usuário
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(responseWithNewline)
}

func main() {
	// Lê a chave da API de uma variável de ambiente
	//decisão de projeto deixar a api key no binário
	//weatherAPIKey := os.Getenv("WEATHERAPI_KEY")
	weatherAPIKey := "ebdba81c2d6c44578a534745252509"
	if weatherAPIKey == "" {
		log.Fatal("A variável de ambiente WEATHER_API_KEY não está definida.")
	}

	//log.Println("Chave da API de clima lida:", weatherAPIKey)

	// Define o manipulador para a rota /cep/
	http.HandleFunc("/cep/", func(w http.ResponseWriter, r *http.Request) {
		getCepWeather(w, r, weatherAPIKey)
	})

	log.Println("Iniciando o servidor na porta 8080...")
	// Inicia o servidor HTTP na porta 8080
	log.Fatal(http.ListenAndServe(":8080", nil))
}
