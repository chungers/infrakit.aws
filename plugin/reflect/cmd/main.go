package main

import (
	"encoding/json"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/infrakit.aws/plugin/reflect"
	"github.com/docker/infrakit/cli"
	"github.com/spf13/cobra"
	"strings"
)

// go run plugin/reflect/cmd/main.go --stack dchung1 --region us-west-1 will reflect on the stack 'dchung1'
func main() {

	builder := &reflect.Builder{}

	var logLevel int
	var name, stack, templateURL string
	var namespaceTags []string
	var reflector reflect.Plugin

	cmd := &cobra.Command{
		Use:   os.Args[0],
		Short: "AWS instance plugin",
		PersistentPreRunE: func(c *cobra.Command, args []string) error {

			cli.SetLogLevel(logLevel)

			namespace := map[string]string{}
			for _, tagKV := range namespaceTags {
				keyAndValue := strings.Split(tagKV, "=")
				if len(keyAndValue) != 2 {
					log.Error("Namespace tags must be formatted as key=value")
					os.Exit(1)
				}

				namespace[keyAndValue[0]] = keyAndValue[1]
			}

			reflectPlugin, err := builder.BuildReflectPlugin(namespace)
			if err != nil {
				return err
			}

			reflector = reflectPlugin
			return nil
		},
	}
	// TODO(chungers) - the exposed flags here won't be set in plugins, because Docker engine plugin install doesn't allow
	// user to pass in command line args like containers with entrypoint.
	cmd.PersistentFlags().IntVar(&logLevel, "log", cli.DefaultLogLevel, "Logging level. 0 is least verbose. Max is 5")
	cmd.PersistentFlags().StringVar(&name, "name", "reflect-aws", "Plugin name to advertise for discovery")
	cmd.PersistentFlags().StringVar(&stack, "stack", "myCFNStack", "CFN stack name to introspect")
	cmd.PersistentFlags().AddFlagSet(builder.Flags())

	inspectCmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspect stack",
		RunE: func(c *cobra.Command, args []string) error {

			if reflector == nil {
				return fmt.Errorf("no plugin")
			}

			model, err := reflector.Inspect(stack)
			if err != nil {
				return err
			}

			buff, err := json.MarshalIndent(model, "", "  ")
			if err != nil {
				return err
			}

			fmt.Println(string(buff))
			return nil
		},
	}

	renderCmd := &cobra.Command{
		Use:   "render",
		Short: "Render Infrakit config",
		RunE: func(c *cobra.Command, args []string) error {

			if reflector == nil {
				return fmt.Errorf("no plugin")
			}

			model, err := reflector.Inspect(stack)
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}

			buff, err := json.MarshalIndent(model, "", "  ")
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}

			// apply the template
			buff, err = reflector.Render(model, templateURL)
			if err != nil {
				return err
			}
			fmt.Println(string(buff))
			return nil
		},
	}
	renderCmd.Flags().StringVar(&templateURL, "template", "", "URL to the template to render")

	cmd.AddCommand(cli.VersionCommand(), inspectCmd, renderCmd)

	err := cmd.Execute()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
