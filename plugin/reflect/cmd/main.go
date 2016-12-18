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
	var name, stack string
	var namespaceTags []string
	cmd := &cobra.Command{
		Use:   os.Args[0],
		Short: "AWS instance plugin",
		Run: func(c *cobra.Command, args []string) {

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
				log.Error(err)
				os.Exit(1)
			}

			cli.SetLogLevel(logLevel)

			spec, err := reflectPlugin.Inspect(stack)
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}

			buff, err := json.MarshalIndent(spec, "", "  ")
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}

			fmt.Println(string(buff))
			//			cli.RunPlugin(name, reflect_plugin.PluginServer(reflectPlugin))
		},
	}

	cmd.Flags().IntVar(&logLevel, "log", cli.DefaultLogLevel, "Logging level. 0 is least verbose. Max is 5")
	cmd.Flags().StringVar(&name, "name", "reflect-aws", "Plugin name to advertise for discovery")
	cmd.Flags().StringVar(&stack, "stack", "myCFNStack", "CFN stack name to introspect")
	cmd.Flags().StringSliceVar(
		&namespaceTags,
		"namespace-tags",
		[]string{},
		"A list of key=value resource tags to namespace all resources created")

	// TODO(chungers) - the exposed flags here won't be set in plugins, because plugin install doesn't allow
	// user to pass in command line args like containers with entrypoint.
	cmd.Flags().AddFlagSet(builder.Flags())

	cmd.AddCommand(cli.VersionCommand())

	err := cmd.Execute()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
