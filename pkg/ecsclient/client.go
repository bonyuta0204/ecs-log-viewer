package ecsclient

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

type EcsClient struct {
	ctx    context.Context
	client *ecs.Client
}

// NewEcsClient creates a new EcsClient.
func NewEcsClient(ctx context.Context, client *ecs.Client) *EcsClient {
	return &EcsClient{
		ctx:    ctx,
		client: client,
	}
}

// ListClusters retrieves all ECS cluster ARNs.
func (e *EcsClient) ListClusters() ([]string, error) {
	var clusters []string
	input := &ecs.ListClustersInput{}
	for {
		resp, err := e.client.ListClusters(e.ctx, input)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, resp.ClusterArns...)
		if resp.NextToken == nil {
			break
		}
		input.NextToken = resp.NextToken
	}
	return clusters, nil
}

// ListRunningTasks lists running tasks in the given cluster.
func (e *EcsClient) ListRunningTasks(cluster string) ([]string, error) {
	var tasks []string
	input := &ecs.ListTasksInput{
		Cluster:       aws.String(cluster),
		DesiredStatus: ecsTypes.DesiredStatusRunning,
	}
	for {
		resp, err := e.client.ListTasks(e.ctx, input)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, resp.TaskArns...)
		if resp.NextToken == nil {
			break
		}
		input.NextToken = resp.NextToken
	}
	return tasks, nil
}

// DescribeTasks calls ECS to describe a list of tasks.
func (e *EcsClient) DescribeTasks(cluster string, taskArns []string) ([]ecsTypes.Task, error) {
	input := &ecs.DescribeTasksInput{
		Cluster: aws.String(cluster),
		Tasks:   taskArns,
	}
	resp, err := e.client.DescribeTasks(e.ctx, input)
	if err != nil {
		return nil, err
	}
	return resp.Tasks, nil
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
