package selector

import (
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/manifoldco/promptui"
)

// SelectorItem represents an item that can be selected through the interactive selector
type SelectorItem interface {
	Label() string
}

// SelectItem presents a list of items to the user and returns the selected item
func SelectItem[T SelectorItem](items []T, prompt string) (T, error) {
	labels := make([]string, len(items))
	for i, item := range items {
		labels[i] = item.Label()
	}
	selector := promptui.Select{
		Label: prompt,
		Items: labels,
	}

	i, _, err := selector.Run()
	if err != nil {
		var zero T
		return zero, err
	}

	return items[i], nil
}

// SelectContainerDefinition presents a list of container definitions to the user and returns the selected one
func SelectContainerDefinition(containerDefinitions []types.ContainerDefinition, prompt string) (types.ContainerDefinition, error) {
	labels := make([]string, len(containerDefinitions))
	for i, item := range containerDefinitions {
		labels[i] = *item.Name
	}
	selector := promptui.Select{
		Label: prompt,
		Items: labels,
	}

	i, _, err := selector.Run()
	if err != nil {
		var zero types.ContainerDefinition
		return zero, err
	}

	return containerDefinitions[i], nil
}
