package mw

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "strings"
    "time"
)

func Logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Printf("%s %s ใช้เวลา %v", r.Method, r.URL.Path,
            time.Since(start).Round(time.Millisecond))
    })
}

func Recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("เชฟทำจานหล่น %v", err)
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusInternalServerError)
                json.NewEncoder(w).Encode(map[string]any{
                    "error": map[string]string{
                        "code":    "INTERNAL_ERROR",
                        "message": "ครัวมีปัญหา ลองใหม่อีกครั้ง",
                    },
                })
            }
        }()
        next.ServeHTTP(w, r)
    })
}

func CORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origin := r.Header.Get("Origin")
        allowedOrigins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
        if len(strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS"))) == 0 {
            allowedOrigins = []string{
                "http://localhost:5173",
                "http://localhost:3000",
                "https://moviehub-670710197.vercel.app",
            }
        }

        for _, allowed := range allowedOrigins {
            if strings.TrimSpace(allowed) == origin {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                w.Header().Set("Vary", "Origin")
                w.Header().Set("Access-Control-Allow-Methods",
                    "GET, POST, PUT, PATCH, DELETE, OPTIONS")
                w.Header().Set("Access-Control-Allow-Headers",
                    "Authorization, Content-Type")
                w.Header().Set("Access-Control-Max-Age", "600")
                if r.Method == http.MethodOptions {
                    w.WriteHeader(http.StatusNoContent)
                    return
                }
                break
            }
        }

        next.ServeHTTP(w, r)
    })
}