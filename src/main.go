package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/subpop/go-log"

	"github.com/redhatinsights/yggdrasil/worker"
	pb "github.com/redhatinsights/yggdrasil_v0/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var yggdDispatchSocketAddr string

func printHelp() {
	fmt.Println("Usage: foreman_ygg_worker [options]")
	fmt.Println("\nThis worker connects to Yggdrasil and handles foreman tasks.")
	fmt.Println("\nOptions:")
	fmt.Println("  -h, --help     Show this help message")
	fmt.Println("\nEnvironment Variables:")
	fmt.Println("  YGG_LOG_LEVEL     Set the logging level (e.g., debug, info, warn, error)")
	fmt.Println("  YGG_SOCKET_ADDR   Address of the Yggdrasil dispatcher socket for gRPC connection")
	fmt.Println("  FOREMAN_YGG_WORKER_WORKDIR   Directory to store worker state, if unset XDG_RUNTIME_DIR is used")
	fmt.Println("  XDG_RUNTIME_DIR   Directory to store worker state, if unset /run is used")
}

func main() {
	// Parse command line flags
	help := flag.Bool("h", false, "Show help message")
	helpLong := flag.Bool("help", false, "Show help message")
	flag.Parse()

	// Show help if requested
	if *help || *helpLong {
		printHelp()
		return
	}

	if len(os.Args) > 1 {
		// Parse the subcommand
		subcommand := os.Args[1]
		switch subcommand {
		default:
			fmt.Printf("Unknown subcommand: %s\n", os.Args[1])
			printHelp()
			os.Exit(1)
		}
	} else {
		runWorker()
	}
}

func runWorker() {
	var ok bool

	yggdLogLevel, ok := os.LookupEnv("YGG_LOG_LEVEL")
	if ok {
		level, ok := log.ParseLevel(yggdLogLevel)
		if ok != nil {
			log.Errorf("Could not parse log level '%v'", yggdLogLevel)
		} else {
			log.SetLevel(level)
		}
	} else {
		// Yggdrasil < 3.0 does not share its configured log level with the
		// workers in any way
		log.SetLevel(log.LevelInfo)
	}

	js := newJobStorage()
	fs := foremanServer{
		serverContext: serverContext{jobStorage: &js, workingDirectory: determineWorkdir()},
	}

	// Get initialization values from the environment.
	yggdDispatchSocketAddr, ok = os.LookupEnv("YGG_SOCKET_ADDR")
	if ok {
		log.Info("YGG_SOCKET_ADDR environment variable found; attempting gRPC connection")

		// Dial the dispatcher on its well-known address.
		conn, err := grpc.Dial(yggdDispatchSocketAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		// Create a dispatcher client
		c := pb.NewDispatcherClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		// Register as a handler of the "foreman" type.
		r, err := c.Register(ctx, &pb.RegistrationRequest{Handler: "foreman", Pid: int64(os.Getpid()), DetachedContent: true})
		if err != nil {
			log.Fatal(err)
		}
		if !r.GetRegistered() {
			log.Fatalf("handler registration failed: %v", err)
		}

		// Listen on the provided socket address.
		l, err := net.Listen("unix", r.GetAddress())
		if err != nil {
			log.Fatal(err)
		}

		// Register as a Worker service with gRPC and start accepting connections.
		s := grpc.NewServer()

		fs.serverContext.externalCommunicator = &YggdrasilGrpc{}

		pb.RegisterWorkerServer(s, &fs)
		if err := s.Serve(l); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Info("YGG_SOCKET_ADDR environment variable not found; attemping D-Bus connection")

		w, err := worker.NewWorker("foreman", true, nil, nil, fs.handleRx, nil)
		if err != nil {
			log.Fatalf("cannot create worker: %v", err)
		}

		// Set up a channel to receive the TERM or INT signal over and clean up
		// before quitting.
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

		fs.serverContext.externalCommunicator = &YggdrasilDBus{w: w}

		if err := w.Connect(quit); err != nil {
			log.Fatalf("cannot connect: %v", err)
		}
	}
}

func determineWorkdir() string {
	workdir, workdirP := os.LookupEnv("FOREMAN_YGG_WORKER_WORKDIR")
	if workdirP {
		return workdir
	}

	workdir, workdirP = os.LookupEnv("XDG_RUNTIME_DIR")
	if workdirP {
		return workdir
	}

	return "/run"
}
