package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jason9075/openfga_webserver/middleware"
	"github.com/jason9075/openfga_webserver/pkg/handler"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/openfga/go-sdk/client"
	"os"
)

// 預設三組權限的資料
var defaultPermissions = []struct {
	object   string
	relation string
	user     string
}{
	{"user:alice", "manager", "user:jason"},
	{"user:ethan", "manager", "user:alice"},
	{"page:jason-page", "owner", "user:jason"},
	{"page:alice-page", "owner", "user:alice"},
	{"page:ethan-page", "owner", "user:ethan"},
}

func main() {

	godotenv.Load()

	e := echo.New()

	fgaClient, err := client.NewSdkClient(&client.ClientConfiguration{
		ApiUrl: os.Getenv("OPENFGA_API_URL"),
	})
	if err != nil {
		e.Logger.Fatal(err)
	}

	// 建立或取得 Store ID
	storeID, err := getOrCreateStore(fgaClient, "FGA Demo")
	if err != nil {
		e.Logger.Fatal(err)
	}
	fmt.Printf("Store ID: %v\n", storeID)

	// 設定 Store ID
	err = fgaClient.SetStoreId(storeID)
	if err != nil {
		e.Logger.Fatal(err)
	}

	// 設定 Model
	if err := setupModel(fgaClient, storeID, "./init/openfga/model.json"); err != nil {
		e.Logger.Fatal(err)
	}

	// Setup default permissions
	if err := setupDefaultPermissions(fgaClient); err != nil {
		e.Logger.Fatal(err)
	}

	authConfig := middleware.OpenFGAConfig{
		Client: fgaClient,
	}

	e.GET("/public", handler.PublicHandler)
	e.GET("/health", handler.HealthCheckHandler)

	fgaGroup := e.Group("/page", middleware.Authorization(authConfig))
	fgaGroup.GET("/:page-uri", handler.PageHandler)

	e.Start(":" + os.Getenv("WEB_APP_PORT"))
}

func getOrCreateStore(fgaClient *client.OpenFgaClient, storeName string) (string, error) {
	ctx := context.Background()

	// 先讀取所有 stores
	listResp, err := fgaClient.ListStores(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to read stores: %v", err)
	}

	// 檢查是否已有同名 store
	for _, s := range listResp.Stores {
		if s.GetName() == storeName {
			// 找到同名 store
			fmt.Printf("Name: %v, ID: %v\n", s.GetName(), s.GetId())
			return s.GetId(), nil
		}
	}

	// 沒找到 -> 新建一個
	createResp, err := fgaClient.CreateStore(ctx).
		Body(client.ClientCreateStoreRequest{
			Name: storeName,
		}).
		Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create store: %v", err)
	}

	return createResp.GetId(), nil
}

// 寫入 Model 的函式 (以 JSON 檔案為來源)
func setupModel(fgaClient *client.OpenFgaClient, storeId string, modelPath string) error {
	ctx := context.Background()

	// 讀取 model.json
	content, err := os.ReadFile(modelPath)
	if err != nil {
		return fmt.Errorf("failed to read model file: %w", err)
	}

	// 解析 JSON => WriteAuthorizationModelRequest
	var writeModelReq client.ClientWriteAuthorizationModelRequest
	if err := json.Unmarshal(content, &writeModelReq); err != nil {
		return fmt.Errorf("failed to unmarshal model JSON: %w", err)
	}

	// 呼叫 WriteAuthorizationModel
	resp, httpResp, err := fgaClient.OpenFgaApi.WriteAuthorizationModel(ctx, storeId).
		Body(writeModelReq).
		Execute()
	if err != nil {
		return fmt.Errorf("failed to write authorization model: %v (HTTP %s)", err, httpResp.Status)
	}

	fmt.Printf("Authorization model created with id: %s\n", resp.GetAuthorizationModelId())
	return nil
}

// setupDefaultPermissions 檢查預設的 (object#relation@user) 是否已存在，
// 若不存在則自動寫入
func setupDefaultPermissions(fgaClient *client.OpenFgaClient) error {
	ctx := context.Background()

	for _, perm := range defaultPermissions {
		// 1. 先 check
		checkReq := client.ClientCheckRequest{
			Object:   perm.object,
			Relation: perm.relation,
			User:     perm.user,
		}
		resp, err := fgaClient.Check(ctx).Body(checkReq).Execute()
		if err != nil {
			return fmt.Errorf("failed to check permission for [%s#%s@%s]: %v",
				perm.object, perm.relation, perm.user, err)
		}

		// 2. 如果不允許 (代表尚未寫入該 tuple)，則 write
		if !resp.GetAllowed() {
			writeReq := client.ClientWriteRequest{
				Writes: []client.ClientTupleKey{
					{
						Object:   perm.object,
						Relation: perm.relation,
						User:     perm.user,
					},
				},
			}
			_, writeErr := fgaClient.Write(ctx).Body(writeReq).Execute()
			if writeErr != nil {
				return fmt.Errorf("failed to write permission for [%s#%s@%s]: %v",
					perm.object, perm.relation, perm.user, writeErr)
			}
			fmt.Printf("Added default permission: [%s#%s@%s]\n",
				perm.object, perm.relation, perm.user)
		}
	}

	return nil
}
