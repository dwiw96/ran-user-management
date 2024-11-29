package password

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func JwtInit(conn *pgxpool.Pool, ctx context.Context) {
	var err error
	privateKey, err := loadKey(conn, ctx) // temporary-should make as method
	if privateKey == nil {
		jwtSetPrivKey(conn, ctx)
	}
	if err != nil {
		log.Println(err)
	}

	log.Println("private key already created")
}

func loadKey(conn *pgxpool.Pool, ctx context.Context) (key *rsa.PrivateKey, err error) {
	q := "select private_key from sec_m"
	var keyBytes []byte
	rows := conn.QueryRow(ctx, q)
	rows.Scan(&keyBytes)
	if keyBytes == nil {
		return nil, err
	}

	return x509.ParsePKCS1PrivateKey(keyBytes)
}

func jwtSetPrivKey(conn *pgxpool.Pool, ctx context.Context) {
	log.Println("Set private key for jwt")
	privateKeyTest := generatePrivateKey()
	err := insertKeyToDatabase(privateKeyTest, conn, ctx)
	if err != nil {
		log.Println("--- [error](JwtSetPrivKey) Failed to insert private key to database, msg:", err)
	}
}

func generatePrivateKey() *rsa.PrivateKey {
	log.Println("<- generate private key")
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Println("--- [error](GeneratePrivateKey) :", err)
	}
	log.Println("-> generate private key is done")
	return privKey
}

func insertKeyToDatabase(privateKey *rsa.PrivateKey, conn *pgxpool.Pool, ctx context.Context) (err error) {
	query := "INSERT INTO sec_m(private_key) VALUES($1);"

	newKeyInbyte := x509.MarshalPKCS1PrivateKey(privateKey)

	res, err := conn.Exec(ctx, query, newKeyInbyte)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	affected := res.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("no rows affected")
	}

	log.Println("private key saved!")

	return err
}
