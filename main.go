package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	for _, cep := range os.Args[1:] {
		endereco := getEndereco(cep)
		fmt.Println(endereco)
	}

	http.HandleFunc("/", handlerCotacao)
	http.ListenAndServe(":8080", nil)
}

func handlerCotacao(w http.ResponseWriter, r *http.Request) {
	log.Println("Request iniciada")
	defer log.Println("Request finalizada")

	cep := r.URL.Query().Get("cep")
	if cep == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	endereco := getEndereco(cep)
	fmt.Println(endereco)
	w.Write([]byte(endereco))
}

func getEndereco(cep string) string {
	ctx := context.Background()
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second))
	defer cancel()

	ch1 := make(chan *string)
	go getViaCep(ch1, cep, ctx)
	ch2 := make(chan *string)
	go getBrasilApi(ch2, cep, ctx)

	select {
	case result := <-ch1:
		// fmt.Println(*result)
		cancel()
		return *result
	case result := <-ch2:
		// fmt.Println(*result)
		cancel()
		return *result
	case <-ctx.Done():
		// fmt.Println("timeout")
		return "timeout"
	}
}

func getViaCep(ch chan<- *string, cep string, ctx context.Context) {
	getURL(ch, "https://viacep.com.br/ws/"+cep+"/json/", ctx)
}

func getBrasilApi(ch chan<- *string, cep string, ctx context.Context) {
	getURL(ch, "https://brasilapi.com.br/api/cep/v1/"+cep, ctx)
}

func getURL(ch chan<- *string, url string, ctx context.Context) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		// println("Erro ao criar request: " + url + " -- \n")
		// println(err.Error())
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// println("Erro ao fazer request " + url + " -- \n")
		// println(err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// println("Erro ao ler body")
		// println(err.Error())
		return
	}
	result := url + " : \n" + string(body)
	ch <- &result
}
