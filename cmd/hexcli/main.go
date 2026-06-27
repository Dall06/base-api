package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed all:template
var templateFS embed.FS

const placeholderModule = "template-placeholder"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "init":
		runInit()
	case "add":
		runAdd()
	default:
		fmt.Printf("Error: Comando '%s' no reconocido.\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Uso: hexcli <comando> [argumentos]")
	fmt.Println("Comandos:")
	fmt.Println("  init <nombre-proyecto> --module <path-modulo>   Inicializa un nuevo proyecto hexagonal")
	fmt.Println("  add service <nombre-servicio>                   Agrega un nuevo servicio hexagonal en srv/")
}

func runInit() {
	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	moduleName := initCmd.String("module", "", "Nombre del módulo Go (ej: github.com/usuario/mi-app)")
	
	if len(os.Args) < 3 {
		fmt.Println("Uso: hexcli init <nombre-proyecto> --module <nombre-modulo>")
		os.Exit(1)
	}
	
	projectName := os.Args[2]
	initCmd.Parse(os.Args[3:])

	if *moduleName == "" {
		fmt.Println("Error: El argumento --module es requerido.")
		os.Exit(1)
	}

	fmt.Printf("Inicializando proyecto '%s' con módulo '%s'...\n", projectName, *moduleName)

	targetDir := projectName
	err := fs.WalkDir(templateFS, "template", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Determinar ruta de destino
		relPath, err := filepath.Rel("template", path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		
		// Remover la extensión .tmpl si estuviera presente
		cleanRelPath := relPath
		if strings.HasSuffix(cleanRelPath, ".tmpl") {
			cleanRelPath = strings.TrimSuffix(cleanRelPath, ".tmpl")
		}
		
		destPath := filepath.Join(targetDir, cleanRelPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		// Asegurar que exista el directorio padre del archivo
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		// Leer contenido e inyectar el módulo correcto
		content, err := templateFS.ReadFile(path)
		if err != nil {
			return err
		}

		fileStr := string(content)
		fileStr = strings.ReplaceAll(fileStr, placeholderModule, *moduleName)

		return os.WriteFile(destPath, []byte(fileStr), 0644)
	})

	if err != nil {
		fmt.Printf("Error al inicializar el proyecto: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("¡Proyecto '%s' inicializado exitosamente en ./%s!\n", projectName, projectName)
}

func runAdd() {
	if len(os.Args) < 4 || os.Args[2] != "service" {
		fmt.Println("Uso: hexcli add service <nombre-servicio>")
		os.Exit(1)
	}

	serviceName := strings.ToLower(os.Args[3])
	fmt.Printf("Agregando servicio '%s' a la estructura...\n", serviceName)

	// Verificar que estemos en la raíz de un proyecto hexagonal (buscando carpeta srv/)
	if _, err := os.Stat("srv"); os.IsNotExist(err) {
		fmt.Println("Error: No se encuentra la carpeta 'srv/'. Asegúrate de estar en la raíz de tu proyecto hexagonal.")
		os.Exit(1)
	}

	basePath := filepath.Join("srv", serviceName)
	dirs := []string{"domain", "ports", "usecases", "handlers", "repositories"}

	for _, d := range dirs {
		dirPath := filepath.Join(basePath, d)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			fmt.Printf("Error al crear carpeta %s: %v\n", dirPath, err)
			os.Exit(1)
		}
	}

	fmt.Printf("¡Estructura para el servicio '%s' creada exitosamente en srv/%s/!\n", serviceName, serviceName)
}
