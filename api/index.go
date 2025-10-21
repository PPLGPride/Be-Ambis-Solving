package handler

import (
	"fmt"
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	// PENTING: Menangani CORS di sini jika Rewrites tidak cukup
	w.Header().Set("Access-Control-Allow-Origin", "https://ambis-task.vercel.app/")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.URL.Path == "/api/register" {
		// Logika register Anda
		fmt.Fprintf(w, "Endpoint Register Berhasil Dihubungi!")
		return
	}
}

// Jika menggunakan router seperti chi atau gorilla/mux, Anda tinggal menjalankan router.
// router.ServeHTTP(w, r)
