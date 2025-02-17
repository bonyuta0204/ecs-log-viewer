package selector

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// selectorItem represents an item that can be selected through the interactive selector
type selectorItem interface {
	Label() string
}

// SelectItem presents a list of items to the user and returns the selected item
func SelectItem[T selectorItem](items []T, prompt string) (T, error) {
	labels := make([]string, len(items))
	for i, item := range items {
		labels[i] = item.Label()
	}
	answer := ""
	option := &survey.Select{
		Message: prompt,
		Options: labels,
	}
	err := survey.AskOne(option, &answer)

	if err != nil {
		var zero T
		return zero, err
	}

	var result T

	for i, item := range items {
		if labels[i] == answer {
			result = item
			break
		}
	}

	return result, nil
}

// SelectContainerDefinition presents a list of container definitions to the user and returns the selected one
func SelectContainerDefinition(containerDefinitions []types.ContainerDefinition, prompt string) (types.ContainerDefinition, error) {
	labels := make([]string, len(containerDefinitions))
	for i, item := range containerDefinitions {
		labels[i] = *item.Name
	}
	answer := ""
	option := &survey.Select{
		Message: prompt,
		Options: labels,
	}
	err := survey.AskOne(option, &answer)

	if err != nil {
		return types.ContainerDefinition{}, err
	}

	var result types.ContainerDefinition
	for i, item := range containerDefinitions {
		if labels[i] == answer {
			result = item
			break
		}
	}

	return result, nil
}
