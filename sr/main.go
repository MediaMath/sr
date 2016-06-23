package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/MediaMath/sr"
	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "sr"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "host",
			EnvVar: "SCHEMA_REGISTRY_URL",
			Usage:  "url to the schema registry",
		},
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "be more wordy",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "add",
			Usage:  "sr add foo-value < schema.json",
			Action: addSchema,
		},
		{
			Name:   "ls",
			Usage:  "sr ls [subject] [version]",
			Action: ls,
		},
	}

	app.Run(os.Args)
}

func getHost(ctx *cli.Context) *sr.Host {
	address := ctx.GlobalString("host")
	if address == "" {
		log.Fatal("host or SCHEMA_REGISTRY_URL must be provided")
	}

	host, err := sr.NewHost(address, ctx.GlobalBool("verbose"))
	if err != nil {
		log.Fatal(err)
	}

	return host
}

func ls(ctx *cli.Context) {

	host := getHost(ctx)

	var resp interface{}
	var err error

	argCount := len(ctx.Args())
	switch argCount {
	case 0:
		resp, err = host.ListSubjects()
	case 1:
		resp, err = host.ListVersions(ctx.Args()[0])
	case 2:
		resp, err = host.GetVersion(ctx.Args()[0], ctx.Args()[1])
	default:
		log.Fatal("usage sr ls [subject] [version]")
	}

	output(ctx, resp, err)
}

func addSchema(ctx *cli.Context) {

	host := getHost(ctx)

	if len(ctx.Args()) < 1 {
		log.Fatal("usage sr add [subject] [name of file | stdin]")
	}

	subject := ctx.Args()[0]

	inputFile, err := getStdinOrFile(ctx)
	if err != nil {
		log.Fatal(err)
	}

	schemaString, err := ioutil.ReadAll(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	schema := &sr.Schema{
		Schema: string(schemaString),
	}

	resp, err := host.AddSchema(subject, schema)
	output(ctx, resp, err)
}

func output(ctx *cli.Context, resp interface{}, err error) {
	if err != nil {
		log.Fatal(err)
	}

	r, err := json.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", r)
}

func getStdinOrFile(ctx *cli.Context) (r io.Reader, err error) {
	r = os.Stdin
	if len(ctx.Args()) > 1 {
		r, err = os.Open(ctx.Args()[1])
	}

	return
}