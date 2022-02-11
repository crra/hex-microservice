package main

import (
	"fmt"
	"hex-microservice/adder"
	"hex-microservice/deleter"
	"hex-microservice/lookup"
	"hex-microservice/meta/value"
	"hex-microservice/typeconverter/parser"
	"path"
	"reflect"

	"github.com/spf13/afero"
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

// typeConverters is the "configuration" for the code generator.
func typeConverters(f afero.Fs) ([]converter, error) {
	var converters []converter

	configForPackage := func(f conversionTemplateFn, paths ...string) config {
		pathLen := len(paths)
		if pathLen < 1 {
			panic("at least one path is required")
		}

		packageName := paths[pathLen-1]

		return config{
			template:           f,
			packageName:        packageName,
			typeFilePath:       path.Join(append(paths, "redirect.go")...),
			fileToGeneratePath: path.Join(append(paths, "converter_gen.go")...),
		}
	}

	for _, c := range []config{
		configForPackage(repositoryTemplate, "repository", "memory"),
		configForPackage(repositoryTemplate, "repository", "redis"),
		configForPackage(repositoryTemplate, "repository", "mongo"),
		// NOTE: "sqlite" performed the mapping manually
		configForPackage(repositoryTemplate, "repository", "gormsqlite"),

		configForPackage(serviceTemplate, "adder"),
		configForPackage(serviceTemplate, "lookup"),
	} {
		parseResult, err := typesFromFile(f, c.typeFilePath)
		if err != nil {
			return converters, fmt.Errorf("error parsing type file: %w", err)
		}

		converters = append(converters, converter{
			Path:    c.fileToGeneratePath,
			Package: c.packageName,
			Data:    c.template(c, parseResult),
		})
	}

	return converters, nil
}

func serviceTemplate(c config, r parser.ParseResult) any {
	return []conversion{
		func(fromTypeName, toTypeName string) conversion {
			return conversion{
				FromTypeName: fromTypeName,
				ToTypeName:   toTypeName,
				MethodName:   methodNameFromTypeNames(fromTypeName, toTypeName),
				Fields: fields(
					value.Must(fieldNamesFromParseResults(r, fromTypeName)),
					value.Must(fieldNamesFromParseResults(r, toTypeName)),
				),
			}
		}("RedirectStorage", "RedirectResult"),
	}
}

func repositoryTemplate(c config, r parser.ParseResult) any {
	return []conversion{
		func(fromTypeName, toTypeName string) conversion {
			return conversion{
				FromTypeName: fromTypeName,
				ToTypeName:   toTypeName,
				MethodName:   methodNameFromTypeNames(fromTypeName, toTypeName),
				Fields: fields(
					value.Must(fieldNamesFromParseResults(r, fromTypeName)),
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
					value.Must(fieldNamesFromParseResults(r, toTypeName)),
				),
			}
		}("adder.RedirectStorage", "redirect"),
		func(fromTypeName, toTypeName string) conversion {
			return conversion{
				FromTypeName: fromTypeName,
				ToTypeName:   toTypeName,
				MethodName:   methodNameFromTypeNames(fromTypeName, toTypeName),
				Fields: fields(
					value.Must(fieldNamesFromParseResults(r, fromTypeName)),
					// TODO: find a way to infer the type from string
					value.Must(fieldNameFromType(reflect.TypeOf(&deleter.RedirectStorage{}))),
				),
			}
		}("redirect", "deleter.RedirectStorage"),
	}
}
