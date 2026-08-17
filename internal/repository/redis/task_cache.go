package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"mmktestbasisByDGanichev/internal/task"

	redisclient "github.com/redis/go-redis/v9"
)

type TaskCache struct {
	client *redisclient.Client
	ttl    time.Duration
}

func NewTaskCache(
	client *redisclient.Client,
	ttl time.Duration,
) *TaskCache {
	return &TaskCache{
		client: client,
		ttl:    ttl,
	}
}

func (c *TaskCache) Get(
	ctx context.Context,
	params task.TaskParams,
	version int64,
) ([]task.Response, error) {
	data, err := c.client.Get(ctx, c.listKey(params, version)).Bytes()
	if err == redisclient.Nil {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("get task list cache: %w", err)
	}

	var response []task.Response
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("unmarshal task list cache: %w", err)
	}

	if response == nil {
		response = make([]task.Response, 0)
	}

	return response, nil
}

func (c *TaskCache) Set(
	ctx context.Context,
	params task.TaskParams,
	version int64,
	response []task.Response,
) error {
	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("marshal task list cache: %w", err)
	}

	if err := c.client.Set(
		ctx,
		c.listKey(params, version),
		data,
		c.ttl,
	).Err(); err != nil {
		return fmt.Errorf("set task list cache: %w", err)
	}

	return nil
}

func (c *TaskCache) Invalidate(
	ctx context.Context,
	teamID int64,
) error {
	if err := c.client.Incr(ctx, c.versionKey(teamID)).Err(); err != nil {
		return fmt.Errorf("increment task list cache version: %w", err)
	}

	return nil
}

func (c *TaskCache) Version(
	ctx context.Context,
	teamID int64,
) (int64, error) {
	version, err := c.client.Get(ctx, c.versionKey(teamID)).Int64()
	if err == redisclient.Nil {
		return 0, nil
	}

	if err != nil {
		return 0, fmt.Errorf("get task list cache version: %w", err)
	}

	return version, nil
}

func (c *TaskCache) versionKey(teamID int64) string {
	return fmt.Sprintf("task-lists:team:%d:version", teamID)
}

func (c *TaskCache) listKey(
	params task.TaskParams,
	version int64,
) string {
	status := "all"
	if params.Status != nil {
		status = string(*params.Status)
	}

	assigneeID := "all"
	if params.AssigneeID != nil {
		assigneeID = strconv.FormatInt(*params.AssigneeID, 10)
	}

	return fmt.Sprintf(
		"task-lists:team:%d:v%d:status:%s:assignee:%s:limit:%d:offset:%d",
		params.TeamID,
		version,
		status,
		assigneeID,
		params.Limit,
		params.Offset,
	)
}
