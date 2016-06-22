package main

import (
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
		{Name: "add",
			Usage:  "sr add foo-value < schema.json",
			Action: addSchema,
		},
	}

	app.Run(os.Args)
}

func getHost(ctx *cli.Context) *sr.Host {
	address := ctx.GlobalString("host")
	if address == "" {
		log.Fatal("host or SCHEMA_REGISTRY_URL must be provided")
	}

	host, err := sr.NewHost(address)
	if err != nil {
		log.Fatal(err)
	}

	return host
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
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	read, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if ctx.GlobalBool("verbose") {
		fmt.Printf("Status: %v\n", resp.Status)
		fmt.Printf("Headers: %v\n", resp.Header)
	}

	fmt.Printf("%s\n", read)

}

func getStdinOrFile(ctx *cli.Context) (r io.Reader, err error) {
	r = os.Stdin
	if len(ctx.Args()) > 1 {
		r, err = os.Open(ctx.Args()[1])
	}

	return
}
