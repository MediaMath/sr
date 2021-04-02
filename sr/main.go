package main

//Copyright 2016 MediaMath <http://www.mediamath.com>.  All rights reserved.
//Use of this source code is governed by a BSD-style
//license that can be found in the LICENSE file.

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/MediaMath/sr"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "sr"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "host",
			EnvVars: []string{"SCHEMA_REGISTRY_URL"},
			Usage:   "url to the schema registry",
		},
		&cli.BoolFlag{
			Name:  "verbose",
			Usage: "be more wordy",
		},
		&cli.BoolFlag{
			Name:  "pretty",
			Usage: "pretty print output",
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:   "add",
			Usage:  "sr add foo-value < schema.json",
			Action: add,
		},
		{
			Name:   "exists",
			Usage:  "sr exists foo-value < schema.json",
			Action: exists,
		},
		{
			Name:   "compatible",
			Usage:  "sr compatible foo-value 3 < schema.json",
			Action: compatible,
		},
		{
			Name:   "ls",
			Usage:  "sr ls [subject] [version]",
			Action: ls,
		},
		{
			Name:   "schema",
			Usage:  "sr schema 7878",
			Action: schema,
		},
		{
			Name:   "config",
			Usage:  "sr config [subject]",
			Action: config,
		},
		{
			Name:   "set-config",
			Usage:  "sr set-config foo FULL",
			Action: setConfig,
		},
		{
			Name:   "copy",
			Usage:  "sr copy from-url to-url from-prefix to-prefix",
			Action: copyFunc,
		},
	}

	var err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func setConfig(ctx *cli.Context) error {
	if ctx.Args().Len() != 2 {
		log.Fatal("sr set-config SUBJECT LEVEL")
	}

	address := getAddress(ctx)
	out(sr.SetSubjectCompatibility(client(ctx), address, sr.Subject(ctx.Args().First()), sr.Compatibility(ctx.Args().Get(1))))
	return nil
}

func config(ctx *cli.Context) error {
	address := getAddress(ctx)
	argCount := ctx.Args().Len()
	switch argCount {
	case 0:
		out(sr.GetDefaultCompatibility(client(ctx), address))
	case 1:
		out(sr.GetSubjectDerivedCompatibility(client(ctx), address, sr.Subject(ctx.Args().First())))
	default:
		log.Fatal("usage sr config [subject]")
	}
	return nil
}

func schema(ctx *cli.Context) error {
	if ctx.Args().Len() != 1 {
		log.Fatal("sr schema ID")
	}

	id, err := strconv.Atoi(ctx.Args().First())
	if err != nil {
		return err
	}

	address := getAddress(ctx)

	out(sr.GetSchema(client(ctx), address, uint32(id)))
	return nil
}

func ls(ctx *cli.Context) error {
	address := getAddress(ctx)
	argCount := ctx.Args().Len()
	switch argCount {
	case 0:
		subjects, err := sr.ListSubjects(client(ctx), address)
		if err != nil {
			log.Fatal(err)
		}

		for _, subject := range subjects {
			fmt.Println(string(subject))
		}
	case 1:
		out(sr.ListVersions(client(ctx), address, sr.Subject(ctx.Args().First())))
	case 2:
		_, schema, err := sr.GetVersion(client(ctx), address, sr.Subject(ctx.Args().First()), ctx.Args().Get(1))
		out(schema, err)
	default:
		log.Fatal("usage sr ls [subject] [version]")
	}
	return nil
}

func compatible(ctx *cli.Context) error {
	address := getAddress(ctx)

	if ctx.Args().Len() < 2 {
		log.Fatal("usage sr compatible [subject] [version] [name of file | stdin]")
	}

	subject := ctx.Args().First()
	version := ctx.Args().Get(1)

	inputFile, err := getStdinOrFile(ctx, 2)
	if err != nil {
		return err
	}

	schemaString, err := ioutil.ReadAll(inputFile)
	if err != nil {
		return err
	}

	out(sr.IsCompatible(client(ctx), address, sr.Subject(subject), version, sr.Schema(schemaString)))
	return nil
}

func exists(ctx *cli.Context) error {
	address := getAddress(ctx)

	if ctx.Args().Len() < 1 {
		log.Fatal("usage sr exists [subject] [name of file | stdin]")
	}

	subject := ctx.Args().First()

	inputFile, err := getStdinOrFile(ctx, 1)
	if err != nil {
		return err
	}

	schemaString, err := ioutil.ReadAll(inputFile)
	if err != nil {
		return err
	}

	version, id, err := sr.HasSchema(client(ctx), address, sr.Subject(subject), sr.Schema(schemaString))
	out(fmt.Sprintf("%v %v", version, id), err)
	return err
}

func add(ctx *cli.Context) error {
	address := getAddress(ctx)

	if ctx.Args().Len() < 1 {
		log.Fatal("usage sr add [subject] [name of file | stdin]")
	}

	subject := ctx.Args().First()

	inputFile, err := getStdinOrFile(ctx, 1)
	if err != nil {
		return err
	}

	schemaString, err := ioutil.ReadAll(inputFile)
	if err != nil {
		return err
	}

	id, err := sr.Register(http.DefaultClient, address, sr.Subject(subject), sr.Schema(string(schemaString)))
	if err != nil {
		return err
	}
	fmt.Printf("%v", id)
	return nil
}

func output(ctx *cli.Context, resp interface{}, err error) {
	if err != nil {
		log.Fatal(err)
	}

	var r []byte
	if ctx.Bool("pretty") {
		r, err = json.MarshalIndent(resp, "", "\t")
	} else {
		r, err = json.Marshal(resp)
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", r)
}

func getStdinOrFile(ctx *cli.Context, index int) (r io.Reader, err error) {
	r = os.Stdin
	if ctx.Args().Len() > index {
		r, err = os.Open(ctx.Args().Get(index))
	}

	return
}

func out(r interface{}, err error) {
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(r)
}

func client(ctx *cli.Context) sr.HTTPClient {
	return http.DefaultClient
}

func getAddress(ctx *cli.Context) string {
	address := ctx.String("host")
	if address == "" {
		log.Fatal("host or SCHEMA_REGISTRY_URL must be provided")
	}

	return address
}

func copyFunc(ctx *cli.Context) error {
	if ctx.Args().Len() < 4 {
		log.Fatal("usage sr copy [sr from url] [sr to url] [from prefix] [to prefix]")
	}

	var fromURL = ctx.Args().First()
	var toURL = ctx.Args().Get(1)
	var fromPrefix = ctx.Args().Get(2)
	var toPrefix = ctx.Args().Get(3)

	var total, err = sr.Copy(http.DefaultClient, fromURL, toURL, fromPrefix, toPrefix)
	if err != nil {
		return err
	}

	fmt.Printf("%d copied", total)
	return nil
}
