package product

import (
	"context"
	"database/sql"

	pb "github.com/genin6382/go-grpc-microservices-benchmark/pb/product"
	log "github.com/sirupsen/logrus"
)

type Server struct {
	pb.UnimplementedProductServiceServer
	DB *sql.DB
}

func (s *Server) GetProductInfo(ctx context.Context, req *pb.ProductRequest) (*pb.ProductResponse, error) {
	log.Infof("Received GetProductDetails request for product ID: %s", req.ProductId)

	product, err := ListProductByID(s.DB, ctx, req.ProductId)

	if err != nil {
		log.Errorf("Error fetching product details: %v", err)
		return nil, err
	}

	return &pb.ProductResponse{
		ProductId: product.Id,
		Price:     float32(product.Price),
		Stock:     int32(product.Stock),
	}, nil
}

func (s *Server) UpdateStock (ctx context.Context, req *pb.UpdateStockRequest) (*pb.UpdateStockResponse, error) {
	log.Infof("Received UpdateProductStock request for product ID: %s with new stock: %d", req.ProductId, req.Quantity)

	product , err := UpdateProductStock(s.DB, ctx, req.ProductId, int(req.Quantity))

	if err != nil {
		log.Errorf("Error updating product stock: %v", err)
		return nil, err
	}

	return &pb.UpdateStockResponse{
		ProductId: req.ProductId,
		Stock:     int32(product.Stock),
		Success:   true,
	}, nil
}
