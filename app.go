package app

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

// Discord session
var dg *discordgo.Session

// Gerar retorna a aplicação de linha de comando e inicia o bot do Discord
func Gerar() *cli.App {
	app := cli.NewApp()
	app.Name = "Aplicação de linha de comando"
	app.Usage = "Busca IPs e Nomes de Host na Web"

	flags := []cli.Flag{
		cli.StringFlag{
			Name:  "host",
			Value: "devbook.com.br",
		},
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar o arquivo .env")
	}

	token := os.Getenv("DISCORD_TOKEN")

	// Inicia o bot do Discord
	dg, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Erro ao criar a sessão do Discord: %v", err)
	}

	// Registra o manipulador de mensagens
	dg.AddHandler(messageHandler)

	// Abre a conexão com o Discord
	err = dg.Open()
	if err != nil {
		log.Fatalf("Erro ao abrir a conexão com o Discord: %v", err)
	}

	fmt.Println("Bot está rodando. Pressione CTRL-C para sair.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	defer dg.Close()

	app.Commands = []cli.Command{
		{
			Name:   "ip",
			Usage:  "Busca IPs na net",
			Flags:  flags,
			Action: buscarIps,
		},
		{
			Name:   "servidores",
			Usage:  "Busca o nome dos servidores na internet",
			Flags:  flags,
			Action: buscarServidores,
		},
	}
	return app
}

// messageHandler lida com as mensagens recebidas no Discord
func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, "!ip ") { // Verifica se a mensagem começa com "!ip "
		host := strings.TrimPrefix(m.Content, "!ip ")
		ips, err := net.LookupIP(host)
		if err != nil {
			log.Println("Erro ao buscar IPs:", err)
			return
		}
		for _, ip := range ips {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("IP de %s: %s", host, ip))
		}
	} else if strings.HasPrefix(m.Content, "!servidores ") { // Verifica se a mensagem começa com "!servidores "
		host := strings.TrimPrefix(m.Content, "!servidores ")
		servidores, err := net.LookupNS(host)
		if err != nil {
			log.Println("Erro ao buscar servidores:", err)
			return
		}
		for _, servidor := range servidores {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Servidor de %s: %s", host, servidor.Host))
		}
	}
}

func buscarIps(c *cli.Context) {
	host := c.String("host")

	ips, erro := net.LookupIP(host)
	if erro != nil {
		log.Fatal(erro)
	}

	for _, ip := range ips {
		fmt.Println(ip)
	}
}

func buscarServidores(c *cli.Context) {
	host := c.String("host")

	servidores, erro := net.LookupNS(host)
	if erro != nil {
		log.Fatal(erro)
	}

	for _, servidor := range servidores {
		fmt.Println(servidor.Host)

	}
}
