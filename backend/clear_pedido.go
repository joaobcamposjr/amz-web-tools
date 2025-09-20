package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"amz-web-tools/backend/internal/config"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	// Load environment variables
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Erro ao carregar .env:", err)
	}

	// Load config
	cfg := config.Load()

	fmt.Printf("🔍 Conectando ao PostgreSQL: %s:%s/%s\n", cfg.PGHost, cfg.PGPort, cfg.PGDatabase)

	// Conectar ao PostgreSQL
	connectionString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.PGHost, cfg.PGPort, cfg.PGUser, cfg.PGPassword, cfg.PGDatabase, cfg.PGSSLMode)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("Erro ao conectar:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Erro no ping:", err)
	}

	fmt.Println("✅ Conexão PostgreSQL estabelecida!")

	// Verificar argumentos
	if len(os.Args) < 2 {
		log.Fatal("Uso: go run clear_pedido.go <numero_pedido>")
	}

	// Limpar registro do pedido
	pedido := os.Args[1]
	fmt.Printf("🗑️ Removendo registro do pedido %s...\n", pedido)

	_, err = db.Exec("DELETE FROM integrator.fato_StatusVenda WHERE num_pedido = $1", pedido)
	if err != nil {
		log.Fatal("Erro ao remover pedido:", err)
	}

	fmt.Printf("✅ Pedido %s removido com sucesso!\n", pedido)
}
