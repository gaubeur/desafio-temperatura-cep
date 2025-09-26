package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetCepWeather(t *testing.T) {
	// Cria um handler de teste para a WeatherAPI
	weatherApiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simula uma resposta de sucesso da WeatherAPI
		response := WeatherResponse{}
		response.Current.TempC = 28.5
		json.NewEncoder(w).Encode(response)
	})

	// Cria um handler de teste para o ViaCEP
	viaCepHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simula uma resposta de sucesso do ViaCEP
		response := CepResponse{}
		response.Localidade = "Santos"
		response.Erro = false
		json.NewEncoder(w).Encode(response)
	})

	// Cria o servidor de teste para a WeatherAPI
	weatherAPIServer := httptest.NewServer(weatherApiHandler)
	defer weatherAPIServer.Close()

	// Cria o servidor de teste para o ViaCEP
	viaCepServer := httptest.NewServer(viaCepHandler)
	defer viaCepServer.Close()

	// Cria uma requisição HTTP para o nosso manipulador
	req, err := http.NewRequest("GET", "/cep/09812480", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Cria um ResponseRecorder para capturar a resposta
	rr := httptest.NewRecorder()

	// Chama o nosso manipulador
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Substitui as URLs das APIs externas pelos nossos servidores de teste
		originalViaCepURL := viacepURL
		viacepURL = viaCepServer.URL + "/ws/%s/json/"
		originalWeatherAPIURL := weatherAPIURL
		weatherAPIURL = weatherAPIServer.URL + "/v1/current.json?key=%s&q=%s"

		getCepWeather(w, r, "dummy-key")

		// Restaura as URLs originais
		viacepURL = originalViaCepURL
		weatherAPIURL = originalWeatherAPIURL
	})

	handler.ServeHTTP(rr, req)

	// Verifica o código de status
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Manipulador retornou o código de status errado: obtido %v, esperado %v", status, http.StatusOK)
	}

	// Verifica o corpo da resposta
	expected := `{"temp_C":28.5,"temp_F":83.3,"temp_K":301.65}`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("Manipulador retornou corpo inesperado: obtido %v, esperado %v", rr.Body.String(), expected)
	}
}
