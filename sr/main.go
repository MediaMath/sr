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
		cli.BoolFlag{
			Name:  "pretty",
			Usage: "pretty print output",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "stupid",
			Usage:  "sr ls foo 12 | sr stupid",
			Action: stupid,
		},
		{
			Name:   "unstupid",
			Usage:  "sr ls foo 12 | sr unstupid",
			Action: unstupid,
		},
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
	}

	app.Run(os.Args)
}

func config(ctx *cli.Context) {
	address := getAddress(ctx)
	argCount := len(ctx.Args())
	switch argCount {
	case 0:
		out(sr.GetDefaultCompatibility(client(ctx), address))
	case 1:
		out(sr.GetSubjectDerivedCompatibility(client(ctx), address, sr.Subject(ctx.Args()[0])))
	default:
		log.Fatal("usage sr config [subject]")
	}
}

func schema(ctx *cli.Context) {
	if len(ctx.Args()) != 1 {
		log.Fatal("sr schema ID")
	}

	id, err := strconv.Atoi(ctx.Args()[0])
	if err != nil {
		log.Fatal(err)
	}

	address := getAddress(ctx)

	out(sr.GetSchema(client(ctx), address, uint32(id)))
}

func ls(ctx *cli.Context) {
	address := getAddress(ctx)
	argCount := len(ctx.Args())
	switch argCount {
	case 0:
		out(sr.ListSubjects(client(ctx), address))
	case 1:
		out(sr.ListVersions(client(ctx), address, sr.Subject(ctx.Args()[0])))
	case 2:
		_, schema, err := sr.GetVersion(client(ctx), address, sr.Subject(ctx.Args()[0]), ctx.Args()[1])
		out(schema, err)
	default:
		log.Fatal("usage sr ls [subject] [version]")
	}
}

func compatible(ctx *cli.Context) {
	address := getAddress(ctx)

	if len(ctx.Args()) < 2 {
		log.Fatal("usage sr compatible [subject] [version] [name of file | stdin]")
	}

	subject := ctx.Args()[0]
	version := ctx.Args()[1]

	inputFile, err := getStdinOrFile(ctx, 2)
	if err != nil {
		log.Fatal(err)
	}

	schemaString, err := ioutil.ReadAll(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	out(sr.IsCompatible(client(ctx), address, sr.Subject(subject), version, sr.Schema(schemaString)))
}

func exists(ctx *cli.Context) {
	address := getAddress(ctx)

	if len(ctx.Args()) < 1 {
		log.Fatal("usage sr exists [subject] [name of file | stdin]")
	}

	subject := ctx.Args()[0]

	inputFile, err := getStdinOrFile(ctx, 1)
	if err != nil {
		log.Fatal(err)
	}

	schemaString, err := ioutil.ReadAll(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	version, id, err := sr.HasSchema(client(ctx), address, sr.Subject(subject), sr.Schema(schemaString))
	out(fmt.Sprintf("%v %v", version, id), err)
}

func add(ctx *cli.Context) {
	address := getAddress(ctx)

	if len(ctx.Args()) < 1 {
		log.Fatal("usage sr add [subject] [name of file | stdin]")
	}

	subject := ctx.Args()[0]

	inputFile, err := getStdinOrFile(ctx, 1)
	if err != nil {
		log.Fatal(err)
	}

	schemaString, err := ioutil.ReadAll(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	id, err := sr.Register(http.DefaultClient, address, sr.Subject(subject), sr.Schema(string(schemaString)))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v", id)
}

func stupid(ctx *cli.Context) {
	inputFile, err := getStdinOrFile(ctx, 0)
	if err != nil {
		log.Fatal(err)
	}

	notStupid, err := ioutil.ReadAll(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	schema := string(notStupid)

	stupid := &sr.SchemaJSON{Schema: sr.Schema(schema)}
	b, err := json.Marshal(stupid)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(b))

}

func unstupid(ctx *cli.Context) {
	inputFile, err := getStdinOrFile(ctx, 0)
	if err != nil {
		log.Fatal(err)
	}

	stupidSchema, err := ioutil.ReadAll(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	schema := &sr.SchemaJSON{}
	err = json.Unmarshal(stupidSchema, schema)
	if err != nil {
		log.Fatal(err)
	}

	jsonObjs := make(map[string]interface{})
	err = json.Unmarshal([]byte(schema.Schema), &jsonObjs)
	output(ctx, jsonObjs, err)
}

func output(ctx *cli.Context, resp interface{}, err error) {
	if err != nil {
		log.Fatal(err)
	}

	var r []byte
	if ctx.GlobalBool("pretty") {
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
	if len(ctx.Args()) > index {
		r, err = os.Open(ctx.Args()[index])
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
	address := ctx.GlobalString("host")
	if address == "" {
		log.Fatal("host or SCHEMA_REGISTRY_URL must be provided")
	}

	return address
}
