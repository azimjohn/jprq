package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"
	"github.com/azimjohn/jprq/cli/i18n"
)

var version = "2.4"

type Flags struct {
	debug     bool
	cname     string
	subdomain string
}

func printVersion() {
	log.Printf("v%s", version)
	os.Exit(0)
}

func printHelp() {
	fmt.Println(i18n.T("usage"))
	fmt.Printf("\n")
	fmt.Println(i18n.T("commands"))
	fmt.Println(i18n.T("command_lang"))
	fmt.Println(i18n.T("command_auth"))
	fmt.Println(i18n.T("command_tcp"))
	fmt.Println(i18n.T("command_http"))
	fmt.Println(i18n.T("command_http-1"))
	fmt.Println(i18n.T("command_http-2"))
	fmt.Println(i18n.T("command_serve"))
	fmt.Println(i18n.T("command_help"))
	fmt.Println(i18n.T("command_version"))
	os.Exit(0)
}

func main() {

	log.SetFlags(0)

	var conf Config
	// 1. Config'ni o'qiymiz. Agar fayl bo'lmasa, log.Fatal qilmaymiz!
	_ = conf.Load()

	// 2. Agar til hali tanlanmagan bo'lsa, default "en" qo'yamiz
	if conf.Local.Language == "" {
		conf.Local.Language = "en"
	}
	i18n.Init(conf.Local.Language)

	if len(os.Args) < 2 {
		log.Println(i18n.T("no_command"))
		printHelp()
	}

	switch os.Args[1] {
	case "help", "--help":
		printHelp()
	case "version", "--version":
		printVersion()
	}

	if len(os.Args) < 3 {
		log.Println(i18n.T("no_args"))
		printHelp()
	}

	protocol, port := "", 0
	command, arg := os.Args[1], os.Args[2]
	flags := parseFlags(os.Args[3:])

	switch command {
		case "lang":
			handleLang(arg, &conf)
		case "auth":
			handleAuth(arg, &conf)
		case "serve":
			protocol, port = handleServe(arg)
		case "tcp", "http":
			protocol = command
			port, _ = strconv.Atoi(arg)
		default:
			log.Fatalf("%s %s, jprq --help", i18n.T("unknown_command"), command)
	}

	if port <= 0 {
		log.Fatalln(i18n.T("port_number"))
	}

	fmt.Printf("jprq %s \t %s\n\n", version, i18n.T("press_quit"))
	defer log.Println(i18n.T("tunnel_closed"))

	client := jprqClient{
		config:    conf,
		protocol:  protocol,
		subdomain: flags.subdomain,
		cname:     flags.cname,
	}

	go client.Start(port, flags.debug)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
}

func parseFlags(args []string) Flags {
	var flags Flags
	for i, arg := range args {
		switch arg {
		case "-d", "-debug", "--debug":
			flags.debug = true
		case "-s", "-subdomain", "--subdomain":
			if i+1 >= len(args) {
				log.Fatal(i18n.T("missing_subdomain"))
			}
			flags.subdomain = args[i+1]
		case "-c", "-cname", "--cname":
			if i+1 >= len(args) {
				log.Fatal(i18n.T("missing_cname"))
			}
			flags.cname = args[i+1]
		}
	}
	return flags
}

func handleLang(lang string, conf *Config) {
	if lang != "uz" && lang != "en" && lang != "ru" {
		log.Fatalln(i18n.T("invalid_language"))
	}
	conf.Local.Language = lang
	if err := conf.Write(); err != nil {
		log.Fatalf("%s %s", i18n.T("error_writing_config"), err)
	}
	i18n.Init(lang)
	log.Printf("%s %s", i18n.T("lang_changed"), lang) // yoki o'zingizning i18n xabaringiz
	os.Exit(0)
}

func handleAuth(token string, conf *Config) {
	conf.Local.AuthToken = token
	if conf.Local.Language == "" {
		conf.Local.Language = "en"
	}
	if err := conf.Write(); err != nil {
		log.Fatalf("%s %s", i18n.T("error_writing_config"), err)
	}
	log.Println(i18n.T("auth_token_set"))
	os.Exit(0)
}

func handleServe(dir string) (string, int) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatalf("%s %s", i18n.T("no_such_dir"), dir)
	}

	handler := http.FileServer(http.Dir(dir))
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("%s %s", i18n.T("failed_start_server"), err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	go func() {
		if err := http.Serve(listener, handler); err != nil {
			log.Fatalf("%s %s: %s", i18n.T("cannot_server_files"), dir, err)
		}
	}()

	time.AfterFunc(600*time.Millisecond, func() {
		log.Println(i18n.T("serving"), " \t",  dir)
	})
	return "http", port
}
