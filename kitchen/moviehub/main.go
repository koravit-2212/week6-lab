package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "strings"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"

    "moviehub/auth"
    "moviehub/handler"
    "moviehub/mw"
    "moviehub/store"
)

func loadEnv(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return nil
        }
        return err
    }

    for _, rawLine := range strings.Split(string(data), "\n") {
        line := strings.TrimSpace(rawLine)
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }

        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            continue
        }

        key := strings.TrimSpace(parts[0])
        value := strings.TrimSpace(parts[1])
        value = strings.Trim(value, `"'`)
        if key == "" {
            continue
        }
        if _, exists := os.LookupEnv(key); !exists {
            if err := os.Setenv(key, value); err != nil {
                return err
            }
        }
    }

    return nil
}

func main() {
    if err := loadEnv(".env"); err != nil {
        log.Fatal("ไม่สามารถอ่านไฟล์ .env ได้: ", err)
    }

    secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
    if len(secret) < 32 {
        log.Fatal("JWT_SECRET ต้องยาวอย่างน้อย 32 ตัวอักษร")
    }

    databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
    if databaseURL == "" {
        log.Fatal("DATABASE_URL ยังไม่ได้ตั้งค่า")
    }

    pool, err := pgxpool.New(context.Background(), databaseURL)
    if err != nil {
        log.Fatal(err)
    }
    defer pool.Close()
    if err := pool.Ping(context.Background()); err != nil {
        log.Fatal("ต่อฐานข้อมูลไม่ติด ", err)
    }

    tickets := &auth.Issuer{Secret: []byte(secret), TTL: time.Hour}
    h := &handler.Handler{Store: &store.Store{Pool: pool}, Tickets: tickets}

    mux := http.NewServeMux()
    mux.HandleFunc("POST /api/auth/register", h.Register)
    mux.HandleFunc("POST /api/auth/login", h.Login)
    mux.HandleFunc("GET /api/movies", h.SearchMovies)
    mux.HandleFunc("GET /api/movies/{id}", h.GetMovie)
    mux.HandleFunc("GET /api/movies/{id}/reviews", h.ListReviews)

    origins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
    if len(strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS"))) == 0 {
        origins = []string{"http://localhost:3000", "https://moviehub-670710197.vercel.app"}
    }

    log.Println("MovieHub API ฟังอยู่ที่ :8080 ต้อนรับ", origins)
    log.Fatal(http.ListenAndServe(":8080", mw.Logging(mw.CORS(origins, mux))))
}