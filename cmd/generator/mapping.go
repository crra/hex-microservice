package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"hex-microservice/adder"
	"hex-microservice/deleter"
	"hex-microservice/lookup"
	"hex-microservice/meta/value"
	"hex-microservice/typeconverter/parser"
	"html/template"
	"io"
	"log"
	"os"
	"os/signal"
	"path"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/spf13/afero"
	"golang.org/x/tools/imports"
)

const (
	templatePath = "cmd/generator"
	// TODO: think about https://pkg.go.dev/embed, but on the other hand, the generator always runs
	//       in the source filetree
	templateFilename = "mapping.tmpl"
)

type fromTo struct {
	From string
	To   string
}

type conversion struct {
	MethodName   string
	FromTypeName string
	ToTypeName   string
	Fields       []fromTo
}

type converter struct {
	Package     string
	Path        string
	Imports     []string
	TypeSource  string
	Conversions []conversion
}

func typeToTitleString(s string) string {
	s = strings.Title(s)
	parts := strings.Split(s, ".")
	if len(parts) == 2 {
		packageName := parts[0]
		typeName := strings.Title(parts[1])
		s = packageName + typeName
	}

	return s
}

func methodNameFromTypeNames(fromType, toType string) string {
	return fmt.Sprintf("from%sTo%s", typeToTitleString(fromType), typeToTitleString(toType))
}

func fields(fromFields, toFields []string) []fromTo {
	var fields []fromTo
	// O(m+n), but the assumption is that m and n are smaller
	// then creating index data structures (e.g. map)
	for _, fromField := range fromFields {
		for _, toField := range toFields {
			if fromField == toField {
				fields = append(fields, fromTo{
					From: fromField,
					To:   toField,
				})
				continue
			}
		}
	}

	return fields
}

var errTypeNotFound = errors.New("type not found found")

func fieldNamesFromParser(f afero.Fs, file, typeName string) ([]string, error) {
	var fields []string
	typeSource, err := f.Open(file)
	if err != nil {
		return fields, fmt.Errorf("error opening type source: %w", err)
	}
	defer typeSource.Close()

	typeSourceContent, err := afero.ReadAll(typeSource)
	if err != nil {
		return fields, fmt.Errorf("error reading template: %w", err)
	}

	p, err := parser.New(string(typeSourceContent))
	if err != nil {
		return fields, fmt.Errorf("error creating parser: %w", err)
	}

	res, err := p.Parse()
	if err != nil {
		return fields, fmt.Errorf("error parsing type source file: %w", err)
	}

	for _, t := range res.Structs {
		if t.Name == typeName {
			return t.Fields, nil
		}
	}

	return fields, errTypeNotFound
}

func fieldNameFromType(t reflect.Type) ([]string, error) {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var fields []string
	for i, maxFields := 0, t.NumField(); i < maxFields; i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		fields = append(fields, f.Name)
	}

	return fields, nil
}

// run encloses the program in a function that can take dependencies (parameters) and can return an error.
func run(parent context.Context, log logr.Logger, f afero.Fs) error {
	templateFile, err := f.Open(path.Join(templatePath, templateFilename))
	if err != nil {
		return fmt.Errorf("error opening template file: %w", err)
	}
	defer templateFile.Close()

	templateContent, err := io.ReadAll(templateFile)
	if err != nil {
		return fmt.Errorf("error reading template: %w", err)
	}

	t, err := template.New("file").Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	converters := []converter{
		{
			Path:    "repository/memory/converter_gen.go",
			Package: "memory",
			Imports: []string{
				"hex-microservice/adder",
				"hex-microservice/deleter",
				"hex-microservice/lookup",
			},
			Conversions: []conversion{
				func(fromTypeName, toTypeName string) conversion {
					return conversion{
						FromTypeName: fromTypeName,
						ToTypeName:   toTypeName,
						MethodName:   methodNameFromTypeNames(fromTypeName, toTypeName),
						Fields: fields(
							// TODO: get parse only once
							value.Must(fieldNamesFromParser(f, "repository/memory/redirect.go", fromTypeName)),
							// TODO: find a way to infer the type from string
							value.Must(fieldNameFromType(reflect.TypeOf(&lookup.RedirectStorage{}))),
						),
					}
				}("redirect", "lookup.RedirectStorage"),
				func(fromTypeName, toTypeName string) conversion {
					return conversion{
						FromTypeName: fromTypeName,
						ToTypeName:   toTypeName,
						MethodName:   methodNameFromTypeNames(fromTypeName, toTypeName),
						Fields: fields(
							// TODO: find a way to infer the type from string
							value.Must(fieldNameFromType(reflect.TypeOf(&adder.RedirectStorage{}))),
							// TODO: get parse only once
							value.Must(fieldNamesFromParser(f, "repository/memory/redirect.go", toTypeName)),
						),
					}
				}("adder.RedirectStorage", "redirect"),
				func(fromTypeName, toTypeName string) conversion {
					return conversion{
						FromTypeName: fromTypeName,
						ToTypeName:   toTypeName,
						MethodName:   methodNameFromTypeNames(fromTypeName, toTypeName),
						Fields: fields(
							// TODO: get parse only once
							value.Must(fieldNamesFromParser(f, "repository/memory/redirect.go", fromTypeName)),
							// TODO: find a way to infer the type from string
							value.Must(fieldNameFromType(reflect.TypeOf(&deleter.RedirectStorage{}))),
						),
					}
				}("redirect", "deleter.RedirectStorage"),
			},
		},
	}

	for _, c := range converters {
		fmt.Printf("%+v\n", c)

		buf := bytes.NewBuffer([]byte{})
		err = t.Execute(buf, c)

		generatedFile, err := f.Create(c.Path)
		if err != nil {
			return fmt.Errorf("error opening destination file: %w", err)
		}
		defer generatedFile.Close()

		final, err := imports.Process("", buf.Bytes(), nil)
		if err != nil {
			return fmt.Errorf("error optimizing imports: %w", err)
		}

		_, err = generatedFile.Write(final)
		if err != nil {
			return fmt.Errorf("error writing to destination file: %w", err)
		}
	}

	return nil
}

// main is the entrypoint of the program.
// main is the only place where external dependencies (e.g. output stream, logger, filesystem)
// are resolved and where final errors are handled (e.g. writing to the console).
func main() {
	// use the built in logger
	log := stdr.New(log.New(os.Stdout, "", log.Lshortfile))

	// create a parent context that listens on os signals (e.g. CTRL-C)
	context, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	// cancel the parent context and all children if an os signal arrives
	go func() {
		<-context.Done()
		cancel()
	}()

	cwd, err := os.Getwd()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	sourceFilesystem := afero.NewBasePathFs(afero.NewOsFs(), cwd)

	// run the program and clean up
	if err := run(context, log, sourceFilesystem); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}
}
