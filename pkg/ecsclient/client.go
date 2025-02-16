package ecsclient

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// ListClusters retrieves all ECS cluster ARNs.
func ListClusters(ctx context.Context, client *ecs.Client) ([]string, error) {
	var clusters []string
	input := &ecs.ListClustersInput{}
	for {
		resp, err := client.ListClusters(ctx, input)
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
func ListRunningTasks(ctx context.Context, client *ecs.Client, cluster string) ([]string, error) {
	var tasks []string
	input := &ecs.ListTasksInput{
		Cluster:       aws.String(cluster),
		DesiredStatus: ecsTypes.DesiredStatusRunning,
	}
	for {
		resp, err := client.ListTasks(ctx, input)
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
func DescribeTasks(ctx context.Context, client *ecs.Client, cluster string, taskArns []string) ([]ecsTypes.Task, error) {
	input := &ecs.DescribeTasksInput{
		Cluster: aws.String(cluster),
		Tasks:   taskArns,
	}
	resp, err := client.DescribeTasks(ctx, input)
	if err != nil {
		return nil, err
	}
	return resp.Tasks, nil
}

// DescribeTaskDefinition retrieves details for a task definition.
func DescribeTaskDefinition(ctx context.Context, client *ecs.Client, taskDefArn string) (*ecsTypes.TaskDefinition, error) {
	input := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefArn),
	}
	resp, err := client.DescribeTaskDefinition(ctx, input)
	if err != nil {
		return nil, err
	}
	return resp.TaskDefinition, nil
}
