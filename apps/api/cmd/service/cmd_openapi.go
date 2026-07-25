package main

import (
	"fmt"
	"log"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/spf13/cobra"
)

func openapiCmd(api huma.API) *cobra.Command {
	var output string

	openapiCmd := &cobra.Command{
		Use:   "openapi",
		Short: "Generate OpenAPI specification",
		Run: func(cmd *cobra.Command, args []string) {
			log.Println("Generating OpenAPI specification...")
			if api == nil {
				must(fmt.Errorf("API is not initialized"), "failed to generate OpenAPI spec")
			}

			b, err := api.OpenAPI().YAML()
			must(err, "failed to generate OpenAPI spec")

			if output == "" {
				fmt.Println(string(b))
			} else {
				err := os.WriteFile(output, b, 0644)
				must(err, "failed to write OpenAPI spec to file", "file", output)
				log.Printf("OpenAPI specification written to file: %s", output)
			}
		},
	}
	openapiCmd.Flags().StringVarP(&output, "output", "o", "", "Output file for OpenAPI specification (optional)")

	return openapiCmd
}
