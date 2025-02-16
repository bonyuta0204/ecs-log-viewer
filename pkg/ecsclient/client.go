package ecsclient

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// EcsClient provides methods to interact with AWS ECS service
type EcsClient struct {
	ctx    context.Context
	client *ecs.Client
}

// NewEcsClient creates a new EcsClient.
func NewEcsClient(ctx context.Context, config *aws.Config) *EcsClient {
	return &EcsClient{
		ctx:    ctx,
		client: ecs.NewFromConfig(*config),
	}
}

// DescribeTaskDefinition retrieves details for a task definition.
func (e *EcsClient) DescribeTaskDefinition(taskDefArn string) (*ecsTypes.TaskDefinition, error) {
	input := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefArn),
	}
	resp, err := e.client.DescribeTaskDefinition(e.ctx, input)
	if err != nil {
		return nil, err
	}
	return resp.TaskDefinition, nil
}

// ListTaskDefinitions retrieves a list of task definitions.
func (e *EcsClient) ListTaskDefinitions() ([]TaskDef, error) {
	var taskDefs []TaskDef
	input := &ecs.ListTaskDefinitionsInput{}
	for {
		resp, err := e.client.ListTaskDefinitions(e.ctx, input)
		if err != nil {
			return nil, err
		}

		for _, taskDefArn := range resp.TaskDefinitionArns {
			taskDefs = append(taskDefs, TaskDef{Arn: taskDefArn})
		}

		if resp.NextToken == nil {
			break
		}
		input.NextToken = resp.NextToken
	}
	return taskDefs, nil
}

// ListTaskDefinitionFamilies retrieves all task definition families from ECS
func (e *EcsClient) ListTaskDefinitionFamilies() ([]TaskDefFamily, error) {
	var families []TaskDefFamily
	input := &ecs.ListTaskDefinitionFamiliesInput{}
	for {
		resp, err := e.client.ListTaskDefinitionFamilies(e.ctx, input)
		if err != nil {
			return nil, err
		}

		for _, family := range resp.Families {
			families = append(families, TaskDefFamily{Name: family})
		}

		if resp.NextToken == nil {
			break
		}
		input.NextToken = resp.NextToken
	}
	return families, nil
}

// DescribeLatestTaskDefinition retrieves the latest task definition for a given family.
func (e *EcsClient) DescribeLatestTaskDefinition(family TaskDefFamily) (*ecsTypes.TaskDefinition, error) {
	input := &ecs.ListTaskDefinitionsInput{
		FamilyPrefix: aws.String(family.Name),
		Sort:         ecsTypes.SortOrderDesc,
		MaxResults:   aws.Int32(1),
	}

	resp, err := e.client.ListTaskDefinitions(e.ctx, input)
	if err != nil {
		return nil, err
	}

	if len(resp.TaskDefinitionArns) == 0 {
		return nil, nil
	}

	// Get the latest task definition
	return e.DescribeTaskDefinition(resp.TaskDefinitionArns[0])
}

// ContainerDefinition represents an ECS container definition with essential information
type ContainerDefinition struct {
	ecsTypes.ContainerDefinition
}

// Label returns the display label for the container definition
func (c ContainerDefinition) Label() string {
	return *c.Name
}

// TaskDef represents an ECS task definition with selector capabilities
type TaskDef struct {
	Arn string
}

// Label returns the display label for the task definition
func (t TaskDef) Label() string {
	return t.Arn
}

// TaskDefFamily represents an ECS task definition family
type TaskDefFamily struct {
	Name string
}

// Label returns the display label for the task definition family
func (t TaskDefFamily) Label() string {
	return t.Name
}
