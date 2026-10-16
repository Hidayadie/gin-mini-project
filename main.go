package main

import (
  "fmt"
  "latihanGin/database"
  "latihanGin/router"
  "log"
)

func main() {
  db, err := database.Connect()
  if err != nil {
    log.Fatal("Gagal terhubung ke database: ", err)
  }
  defer db.Close()
  fmt.Println("✅ Terhubung ke database PostgreSQL")

  r := router.SetupRouter(db)

  fmt.Println("🚀 Server Gin berjalan di http://localhost:8080")
  if err := r.Run(":8080"); err != nil {
    log.Fatal("Gagal menjalankan server: ", err)
  }
}
