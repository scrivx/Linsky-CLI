package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"linsky-backend/internal/config"
	"linsky-backend/internal/db"
	"linsky-backend/internal/repository"
	"linsky-backend/internal/service"
	"linsky-backend/internal/web"

	"github.com/fatih/color"
)

func main() {
    ctx := context.Background()

    // Cargar config
    cfg, err := config.Load()
    if err != nil {
        color.New(color.FgRed).Printf("Error cargando config: %v\n", err)
        os.Exit(1)
    }

    // Conectar BD
    pool, err := db.NewPool(cfg)
    if err != nil {
        color.New(color.FgRed).Printf("%v\n", err)
        os.Exit(1)
    }
    defer pool.Close()

    // Inicializar capas
    repo := repository.NewURLRepository(pool)
    svc := service.NewURLService(repo, cfg.BaseURL)

    // Migrar tabla
    if err := repo.Migrate(ctx); err != nil {
        color.New(color.FgRed).Printf("Error migrando: %v\n", err)
        os.Exit(1)
    }

    // Iniciar servidor HTTP (para integrar con frontend)
    addr := ":" + cfg.HTTPPort
    srv := web.Start(ctx, svc, addr)
    color.New(color.FgGreen).Printf("🚀 HTTP: http://localhost:%s (API en /api)\n", cfg.HTTPPort)

    // CLI interactivo (mantener compatibilidad)
    scanner := bufio.NewScanner(os.Stdin)
    printHelp()

    for {
        fmt.Print("\n> ")
        if !scanner.Scan() {
            break
        }

        parts := strings.Fields(scanner.Text())
        if len(parts) == 0 {
            continue
        }

        switch parts[0] {

        case "create":
            if len(parts) < 3 {
                color.New(color.FgYellow).Println("Uso: create <alias> <url>")
                continue
            }
            u, err := svc.Shorten(ctx, parts[1], parts[2])
            if err != nil {
                color.New(color.FgRed).Printf("❌ Error: %v\n", err)
                continue
            }
            color.New(color.FgGreen).Printf("✅ Creado!\n   Alias : %s\n   Corta : %s\n   Larga : %s\n",
                u.Alias, svc.ShortLink(u.Alias), u.OriginalURL)

        case "resolve":
            if len(parts) < 2 {
                color.New(color.FgYellow).Println("Uso: resolve <alias>")
                continue
            }
            u, err := svc.Resolve(ctx, parts[1])
            if err != nil {
                color.New(color.FgRed).Printf("❌ %v\n", err)
                continue
            }
            color.New(color.FgCyan).Printf("🔗 %s  →  %s\n   Clicks: %d\n", svc.ShortLink(u.Alias), u.OriginalURL, u.Clicks)

        case "list":
            urls, err := svc.List(ctx)
            if err != nil {
                color.New(color.FgRed).Printf("❌ Error: %v\n", err)
                continue
            }
            if len(urls) == 0 {
                color.New(color.FgYellow).Println("No hay URLs registradas.")
                continue
            }
            color.New(color.FgMagenta).Println("\n┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓")
            color.New(color.FgMagenta).Println("┃                   URLs registradas                      ┃")
            color.New(color.FgMagenta).Println("┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛")
            fmt.Printf("  %-20s %-8s %s\n", "ALIAS", "CLICKS", "URL ORIGINAL")
            fmt.Println("  " + strings.Repeat("-", 70))
            for _, u := range urls {
                fmt.Printf("  %-20s %-8d %s\n", u.Alias, u.Clicks, u.OriginalURL)
            }

        case "delete":
            if len(parts) < 2 {
                color.New(color.FgYellow).Println("Uso: delete <alias>")
                continue
            }
            if err := svc.Delete(ctx, parts[1]); err != nil {
                color.New(color.FgRed).Printf("❌ %v\n", err)
                continue
            }
            color.New(color.FgGreen).Printf("🗑️  Alias '%s' eliminado.\n", parts[1])

        case "help":
            printHelp()

        case "exit", "quit":
            color.New(color.FgBlue).Println("👋 Hasta luego!")
            // Intentar apagar servidor HTTP
            ctxShut, cancel := context.WithTimeout(context.Background(), 5*time.Second)
            defer cancel()
            if srv != nil {
                _ = srv.Shutdown(ctxShut)
            }
            return

        default:
            color.New(color.FgYellow).Println("Comando desconocido. Escribe 'help' para ver los comandos.")
        }
    }
}

func printHelp() {
    hiRed   := color.New(color.FgHiRed)
    hiCyan  := color.New(color.FgHiCyan)
    hiBlue  := color.New(color.FgHiBlue)
    yellow  := color.New(color.FgYellow)
    gray    := color.New(color.FgHiBlack)
    green   := color.New(color.FgHiGreen)

    hiRed.Println(`
        ██╗     ██╗███╗   ██╗███████╗██╗  ██╗██╗   ██╗
        ██║     ██║████╗  ██║██╔════╝██║ ██╔╝╚██╗ ██╔╝
        ██║     ██║██╔██╗ ██║███████╗█████╔╝  ╚████╔╝ 
        ██║     ██║██║╚██╗██║╚════██║██╔═██╗   ╚██╔╝  
        ███████╗██║██║ ╚████║███████║██║  ██╗   ██║   
        ╚══════╝╚═╝╚═╝  ╚═══╝╚══════╝╚═╝  ╚═╝   ╚═╝   
    `)

    hiCyan.Println("  Linsky CLI — v1.0.0")

    gray.Println("  ─────────────────────────────────────────────────────")
    hiCyan.Println("  COMANDOS DISPONIBLES")
    gray.Println("  ─────────────────────────────────────────────────────")

    printCmd(yellow, hiBlue, gray,
        "create", "<alias> <url>",
        "Crear una URL corta con alias personalizado")

    printCmd(yellow, hiBlue, gray,
        "resolve", "<alias>    ",
        "Obtener la URL original")
    green.Println("                                               [+1 click]")

    printCmd(yellow, hiBlue, gray,
        "list", "           ",
        "Listar todas las URLs registradas")

    printCmd(yellow, hiBlue, gray,
        "delete", "<alias>    ",
        "Eliminar un alias existente")

    printCmd(yellow, hiBlue, gray,
        "help", "           ",
        "Mostrar esta ayuda")

    printCmd(yellow, hiBlue, gray,
        "exit", "           ",
        "Salir de la aplicación")

    gray.Println("  ─────────────────────────────────────────────────────")
    gray.Print("  Tip: usa ")
    hiCyan.Print("linsky help <comando>")
    gray.Println(" para más detalles.")
    gray.Println("  ─────────────────────────────────────────────────────")
}

func printCmd(
    cmdColor, argColor, descColor *color.Color,
    cmd, args, desc string,
) {
    fmt.Print("  ")
    cmdColor.Printf("%-10s", cmd)
    argColor.Printf("%-18s", args)
    descColor.Println(desc)
}
