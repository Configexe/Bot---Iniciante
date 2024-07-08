package app

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
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
	// Configurar Intents
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages
	// Registra o manipulador de interações
	dg.AddHandler(interactionHandler)
	// Abre a conexão com o Discord
	err = dg.Open()
	if err != nil {
		log.Fatalf("Erro ao abrir a conexão com o Discord: %v", err)
	}

	log.Println("Conectando ao Discord...") // Adicione um log aqui para depuração

	// Registrar comandos de barra
	for _, command := range app.Commands {
		_, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", &discordgo.ApplicationCommand{
			Name:        command.Name,
			Description: command.Usage,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "host",
					Description: "Host para buscar os IPs/servidores",
					Required:    true,
				},
			},
		})
		if err != nil {
			log.Fatalf("Erro ao registrar o comando %s: %v", command.Name, err)
		}
	}

	// Adicionando comando /ping
	_, err = dg.ApplicationCommandCreate(dg.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:        "ping",
		Description: "Verifica se o bot está online",
	})
	if err != nil {
		log.Fatalf("Erro ao registrar o comando ping: %v", err)
	}

	// Adicionando comando /ip
	_, err = dg.ApplicationCommandCreate(dg.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:        "ip",
		Description: "Busca IPs na net",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "host",
				Description: "Host para buscar os IPs",
				Required:    true,
			},
		},
	})
	if err != nil {
		log.Fatalf("Erro ao registrar o comando ip: %v", err)
	}

	// Adicionando comando /servidores
	_, err = dg.ApplicationCommandCreate(dg.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:        "servidores",
		Description: "Busca o nome dos servidores na internet",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "host",
				Description: "Host para buscar os servidores",
				Required:    true,
			},
		},
	})
	if err != nil {
		log.Fatalf("Erro ao registrar o comando servidores: %v", err)
	}

	fmt.Println("Bot está rodando. Pressione CTRL-C para sair.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	defer dg.Close() // Fecha a conexão com o discord no final da execução

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

// interactionHandler lida com as interações com os comandos de barra
func interactionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	commandData := i.ApplicationCommandData()
	host := commandData.Options[0].StringValue() // Obtém o argumento "host" do comando de barra

	switch commandData.Name {
	case "ping":
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Pong!",
			},
		})
	case "ip":
		ips, err := net.LookupIP(host)
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Erro ao buscar IPs para %s: %v", host, err),
				},
			})
			return
		}
		response := fmt.Sprintf("IPs de %s:\n", host)
		for _, ip := range ips {
			response += fmt.Sprintf("- %s\n", ip)
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
	case "servidores":
		servidores, err := net.LookupNS(host)
		if err != nil {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Erro ao buscar servidores para %s: %v", host, err),
				},
			})
			return
		}
		response := fmt.Sprintf("Servidores de %s:\n", host)
		for _, servidor := range servidores {
			response += fmt.Sprintf("- %s\n", servidor.Host)
		}
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
	default:
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Comando inválido!",
			},
		})
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
