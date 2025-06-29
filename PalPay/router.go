package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

func setupRoutes(db *sql.DB) {
	http.Handle("/styles/", http.StripPrefix("/styles/", http.FileServer(http.Dir("styles"))))
	http.Handle("/scripts/", http.StripPrefix("/scripts/", http.FileServer(http.Dir("scripts"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "login.html"
		} else {
			// Rimuovi lo slash iniziale
			path = path[1:]
		}
		http.ServeFile(w, r, path)
	})

	// ----------------------------- Login Endpoint -----------------------------
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var creds struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		var user struct {
			Username   string  `json:"username"`
			Balance    float64 `json:"balance"`
			FirstLogin bool    `json:"firstlogin"`
			Password   string  `json:"password"`
		}

		var dbPassword string

		err := db.QueryRow("SELECT username, password, balance FROM users WHERE username = ?", creds.Username).
			Scan(&user.Username, &dbPassword, &user.Balance)

		// If the user exists but there was an error return error
		if err != nil && err != sql.ErrNoRows {
			fmt.Println("Error querying user:", err)

			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		} else if err == sql.ErrNoRows {
			// If the user doesn't exist, create a new user with a default balance of 0
			_, err = db.Exec("INSERT INTO users (username, password, balance) VALUES (?, ?, ?)", creds.Username, creds.Password, 0.0)
			if err != nil {
				fmt.Println("Error inserting new user:", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			user.Username = creds.Username
			user.Balance = 0.0
			user.FirstLogin = true
			user.Password = creds.Password
		} else {
			if user.Username != "" && dbPassword != creds.Password {
				fmt.Println("Invalid credentials for user:", user.Username)
				http.Error(w, "Invalid credentials", http.StatusUnauthorized)
				return
			}
			// Controlla se l'utente ha già una carta
			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM cards WHERE username = ?", user.Username).Scan(&count)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if count == 0 {
				user.FirstLogin = true
			} else {

				user.FirstLogin = false
			}
			user.Password = dbPassword
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	})

	// ----------------------------- Add Funds Endpoint -----------------------------
	http.HandleFunc("/add-funds", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Username string  `json:"username"`
			Amount   float64 `json:"amount"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Aggiorna il balance
		_, err := db.Exec("UPDATE users SET balance = ? WHERE username = ?", req.Amount, req.Username)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		var message string
		message = "Funds added"
		_, err = db.Exec("INSERT INTO transactions (sender, receiver, amount, message) VALUES (?, ?, ?, ?)", req.Username, req.Username, req.Amount, message)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]float64{"balance": req.Amount})
	})

	// ----------------------------- Send Money Endpoint -----------------------------
	http.HandleFunc("/send-money", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Sender   string  `json:"sender"`
			Receiver string  `json:"receiver"`
			Amount   float64 `json:"amount"`
			Message  string  `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Aggiorna il balance del sender
		var senderBalance float64
		err := db.QueryRow("SELECT balance FROM users WHERE username = ?", req.Sender).
			Scan(&senderBalance)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Sender not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		senderBalance -= req.Amount
		_, err = db.Exec("UPDATE users SET balance = ? WHERE username = ?", senderBalance, req.Sender)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Aggiorna il balance del receiver
		var receiverBalance float64
		err = db.QueryRow("SELECT balance FROM users WHERE username = ?", req.Receiver).
			Scan(&receiverBalance)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Receiver not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		receiverBalance += req.Amount
		_, err = db.Exec("UPDATE users SET balance = ? WHERE username = ?", receiverBalance, req.Receiver)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Inserisci la transazione nel database
		_, err = db.Exec("INSERT INTO transactions (sender, receiver, amount, message) VALUES (?, ?, ?, ?)", req.Sender, req.Receiver, req.Amount, req.Message)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]float64{"newBalance": senderBalance})
	})

	// ----------------------------- Transactions Endpoint -----------------------------
	http.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "Username is required", http.StatusBadRequest)
			return
		}

		rows, err := db.Query("SELECT sender, receiver, amount, message, timestamp FROM transactions WHERE sender = ? OR receiver = ? ORDER BY timestamp DESC", username, username)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		defer rows.Close()
		var transactions []struct {
			Sender    string  `json:"sender"`
			Receiver  string  `json:"receiver"`
			Amount    float64 `json:"amount"`
			Message   string  `json:"message"`
			Timestamp string  `json:"timestamp"`
		}

		for rows.Next() {
			var t struct {
				Sender    string  `json:"sender"`
				Receiver  string  `json:"receiver"`
				Amount    float64 `json:"amount"`
				Message   string  `json:"message"`
				Timestamp string  `json:"timestamp"`
			}
			if err := rows.Scan(&t.Sender, &t.Receiver, &t.Amount, &t.Message, &t.Timestamp); err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			transactions = append(transactions, t)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(transactions)
	})

	// ----------------------------- Card Management Endpoints -----------------------------
	http.HandleFunc("/set-card", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Username   string `json:"username"`
			CardNumber string `json:"card_number"`
			CVV        string `json:"cvv"`
			Expiry     string `json:"expiry"`
			HolderName string `json:"holder_name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		// Sovrascrivi la carta se già presente per quell'utente
		_, err := db.Exec(`
            INSERT INTO cards (username, card_number, cvv, expiry, holder_name)
            VALUES (?, ?, ?, ?, ?)
            ON CONFLICT(username) DO UPDATE SET
                card_number=excluded.card_number,
                cvv=excluded.cvv,
                expiry=excluded.expiry,
                holder_name=excluded.holder_name
                
        `, req.Username, req.CardNumber, req.CVV, req.Expiry, req.HolderName)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			fmt.Println("Error setting card:", err)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// ------------------------------ Get Card Endpoint -----------------------------
	http.HandleFunc("/get-card", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "Username required", http.StatusBadRequest)
			return
		}
		var card struct {
			CardNumber  string `json:"card_number"`
			Expiry      string `json:"expiry"`
			HolderName  string `json:"holder_name"`
			LastUpdated string `json:"last_updated"`
		}
		err := db.QueryRow("SELECT card_number, expiry, holder_name, created_at FROM cards WHERE username = ?", username).
			Scan(&card.CardNumber, &card.Expiry, &card.HolderName, &card.LastUpdated)
		if err != nil {
			http.Error(w, "Card not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(card)
	})

	// ----------------------------- Search Users Endpoint -----------------------------
	http.HandleFunc("/search-users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		query := r.URL.Query().Get("q")
		if query == "" {
			http.Error(w, "Query parameter is required", http.StatusBadRequest)
			return
		}

		rows, err := db.Query("SELECT username FROM users WHERE LOWER(username) LIKE LOWER(?) LIMIT 10", query+"%")
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var users []string
		for rows.Next() {
			var username string
			if err := rows.Scan(&username); err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
			users = append(users, username)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	})

	// ----------------------------- Get Balance Endpoint -----------------------------
	http.HandleFunc("/get-balance", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		username := r.URL.Query().Get("username")
		if username == "" {
			http.Error(w, "Username is required", http.StatusBadRequest)
			return
		}
		var balance float64
		err := db.QueryRow("SELECT balance FROM users WHERE username = ?", username).Scan(&balance)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]float64{"balance": balance})
	})

}
