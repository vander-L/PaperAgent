package utility

import (
	"context"
	"fmt"

	cli "github.com/milvus-io/milvus-sdk-go/v2/client"
)

func NewMilvusClient(ctx context.Context) (cli.Client, error) {
	// 1. 先连接default数据库
	defaultClient, err := cli.NewClient(ctx, cli.Config{
		Address: "localhost:19530",
		DBName:  "default",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to default database: %w", err)
	}

	// 2. 检查agent数据库是否存在，不存在则创建
	databases, err := defaultClient.ListDatabases(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}
	agentDBExists := false
	for _, db := range databases {
		if db.Name == MilvusDBName {
			agentDBExists = true
			break
		}
	}
	if !agentDBExists {
		err = defaultClient.CreateDatabase(ctx, MilvusDBName)
		if err != nil {
			return nil, fmt.Errorf("failed to create agent database: %w", err)
		}
	}

	// 3. 创建连接到agent数据库的客户端
	agentClient, err := cli.NewClient(ctx, cli.Config{
		Address: "localhost:19530",
		DBName:  MilvusDBName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to agent database: %w", err)
	}
	// 4. 检查biz collection是否存在，不存在则创建
	collections, err := agentClient.ListCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	paperCollectionExists := false
	for _, collection := range collections {
		if collection.Name == MilvusCollectionName {
			paperCollectionExists = true
			break
		}
	}

	if !paperCollectionExists {
		panic("paper collection not found")
	}
	defaultClient.Close()
	return agentClient, nil
}
