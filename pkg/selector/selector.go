package selector

import (
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/manifoldco/promptui"
)

type SelectorItem interface {
	Label() string
}

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
