package order 

import (
	"context"
	"time"

	userpb "github.com/genin6382/go-grpc-microservices-benchmark/pb/user"
	productpb "github.com/genin6382/go-grpc-microservices-benchmark/pb/product"

    log "github.com/sirupsen/logrus"
)

type UserServiceClient struct {
	Client userpb.UserServiceClient
}

func NewUserServiceClient(client userpb.UserServiceClient) *UserServiceClient {
    return &UserServiceClient{Client: client}
}

func (u *UserServiceClient) CheckUserExists(ctx context.Context, userID string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    resp, err := u.Client.CheckUserExists(ctx, &userpb.UserRequest{UserId: userID})
    if err != nil {
		log.Errorf("Error calling CheckUserExists: %v", err)
		return false, err
	}
	return resp.Exists, nil

}

type ProductServiceClient struct {
	Client productpb.ProductServiceClient
}

func NewProductServiceClient(client productpb.ProductServiceClient) *ProductServiceClient{
	return &ProductServiceClient{Client: client}
}

func (p *ProductServiceClient) GetProductInfo(ctx context.Context, productID string) (*productpb.ProductResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

    resp , err := p.Client.GetProductInfo(ctx,&productpb.ProductRequest{ProductId: productID})
	if err != nil {
		log.Errorf("Error calling GetProductInfo: %v", err)
		return nil, err
	}
	return resp, nil
}
func (p *ProductServiceClient) UpdateStock(ctx context.Context, productID string, quantity int32) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := p.Client.UpdateStock(ctx, &productpb.UpdateStockRequest{
		ProductId: productID,
		Quantity:  quantity,
	})
	if err != nil {
		log.Errorf("Error calling UpdateStock: %v", err)
		return false, err
	}
	return resp.Success, nil
}