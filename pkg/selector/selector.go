package selector

import (
	"fmt"
	"log"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// selectorItem represents an item that can be selected through the interactive selector.
type selectorItem interface {
	Label() string
}

// selectByLabels displays a selection prompt with the given labels and returns the selected label.
func selectByLabels(labels []string, prompt string) (string, error) {
	var out terminal.FileWriter
	if tty, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0); err == nil {
		defer func() {
			if err := tty.Close(); err != nil {
				log.Printf("Warning: failed to close tty: %v\n", err)
			}
		}()
		out = tty
	} else {
		out = os.Stdout
	}

	option := &survey.Select{
		Message: prompt,
		Options: labels,
	}

	var answer string
	if err := survey.AskOne(option, &answer, survey.WithStdio(os.Stdin, out, os.Stderr)); err != nil {
		return "", err
	}

	return answer, nil
}

// SelectItem presents a list of items to the user and returns the selected item.
func SelectItem[T selectorItem](items []T, prompt string) (T, error) {
	labels := make([]string, len(items))
	for i, item := range items {
		labels[i] = item.Label()
	}

	answer, err := selectByLabels(labels, prompt)
	if err != nil {
		var zero T
		return zero, err
	}

	for i, item := range items {
		if labels[i] == answer {
			return item, nil
		}
	}

	var zero T
	return zero, fmt.Errorf("selected answer %q not found in items", answer)
}

// SelectContainerDefinition presents a list of container definitions to the user and returns the selected one.
func SelectContainerDefinition(containerDefinitions []types.ContainerDefinition, prompt string) (types.ContainerDefinition, error) {
	labels := make([]string, len(containerDefinitions))
	for i, item := range containerDefinitions {
		// Check for nil to avoid panic.
		if item.Name != nil {
			labels[i] = *item.Name
		} else {
			labels[i] = "<Unnamed>"
		}
	}

	answer, err := selectByLabels(labels, prompt)
	if err != nil {
		return types.ContainerDefinition{}, err
	}

	for _, item := range containerDefinitions {
		if item.Name != nil && *item.Name == answer {
			return item, nil
		}
	}

	return types.ContainerDefinition{}, fmt.Errorf("selected answer %q not found in container definitions", answer)
}
