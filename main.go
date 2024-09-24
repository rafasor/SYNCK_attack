package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v2"
)

type Config struct {
	OutputDir      string   `yaml:"output_dir"`
	Subdirectories []string `yaml:"subdirectories"`
}

type Tool struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

type SubdomainEnumConfig struct {
	Tools []Tool `yaml:"tools"`
}

func main() {
	// Carregar variáveis de ambiente do arquivo .env
	err := godotenv.Load("yaml_config/.env")
	if err != nil {
		log.Fatalf("Erro ao carregar o arquivo .env: %v", err)
	}

	// Obter o token do GitHub da variável de ambiente
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Fatalf("GITHUB_TOKEN não está definido no arquivo .env")
	}

	chaosToken := os.Getenv("CHAOS_TOKEN")
	if chaosToken == "" {
		log.Fatalf("CHAOS TOKEN não está definido")
	}

	// Leia o arquivo de configuração principal
	data, err := ioutil.ReadFile("yaml_config/config.yaml")
	if err != nil {
		log.Fatalf("Erro ao ler o arquivo de configuração: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("Erro ao deserializar o YAML: %v", err)
	}

	// Solicite o domínio ao usuário
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Digite o domínio: ")
	domain, _ := reader.ReadString('\n')
	domain = strings.TrimSpace(domain)

	// Atualize o diretório de saída com o domínio fornecido
	outputDir := fmt.Sprintf("%s%s", config.OutputDir, domain)

	// Crie os diretórios conforme a configuração
	for _, subDir := range config.Subdirectories {
		path := filepath.Join(outputDir, subDir)
		if err := os.MkdirAll(path, 0755); err != nil {
			log.Fatalf("Erro ao criar o diretório %s: %v", path, err)
		}
	}

	fmt.Printf("[+] Iniciando o reconhecimento abrangente e a verificação de vulnerabilidades para %s\n", domain)

	// Leia o arquivo de configuração de enumeração de subdomínios
	data, err = ioutil.ReadFile("yaml_config/subdomain_enum.yaml")
	if err != nil {
		log.Fatalf("Erro ao ler o arquivo de configuração de enumeração de subdomínios: %v", err)
	}

	var subdomainConfig SubdomainEnumConfig
	if err := yaml.Unmarshal(data, &subdomainConfig); err != nil {
		log.Fatalf("Erro ao deserializar o YAML de enumeração de subdomínios: %v", err)
	}

	// Função para executar comandos
	runCommand := func(command string) {
		cmd := exec.Command("sh", "-c", command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("Erro ao executar comando %s: %v", command, err)
		}
	}

	// Executar ferramentas de enumeração de subdomínios e DNS
	fmt.Println("[+] Enumerar subdomínios e executando a enumeração DNS...")
	for _, tool := range subdomainConfig.Tools {
		tmpl, err := template.New("command").Parse(tool.Command)
		if err != nil {
			log.Fatalf("Erro ao analisar o template do comando: %v", err)
		}

		var commandBuilder strings.Builder
		err = tmpl.Execute(&commandBuilder, map[string]string{
			"domain":       domain,
			"output_dir":   outputDir,
			"github_token": githubToken,
			"chaos_token":  chaosToken,
		})
		if err != nil {
			log.Fatalf("Erro ao executar o template do comando: %v", err)
		}

		runCommand(commandBuilder.String())
	}
}
