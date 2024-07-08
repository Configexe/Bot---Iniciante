package main

import (
	"fmt"
	"log"
	"os"
	"projeto-go/app"
)

func main() {
	fmt.Println("Ponto de inicio")

	aplicacao := app.Gerar()
	if erro := aplicacao.Run(os.Args); erro != nil {
		log.Fatal(erro)
	}
}
