package order

import (
	"context"
	"fmt"
	"time"

	"github.com/cdxy1/go-courier-service/internal/model"
	"github.com/cdxy1/go-courier-service/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderGateway struct {
	client proto.OrdersServiceClient
	conn   *grpc.ClientConn
}

func NewOrderGateway(address string) (*OrderGateway, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	client := proto.NewOrdersServiceClient(conn)

	return &OrderGateway{
		client: client,
		conn:   conn,
	}, nil
}

func (g *OrderGateway) Close() error {
	if g.conn != nil {
		return g.conn.Close()
	}
	return nil
}

func (g *OrderGateway) GetOrders(ctx context.Context, from time.Time) ([]*model.Order, error) {
	req := &proto.GetOrdersRequest{
		From: timestamppb.New(from),
	}

	resp, err := g.client.GetOrders(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	orders := make([]*model.Order, 0, len(resp.Orders))
	for _, pbOrder := range resp.Orders {
		order := &model.Order{
			ID:                pbOrder.Id,
			UserID:            pbOrder.UserId,
			OrderNumber:       pbOrder.OrderNumber,
			FIO:               pbOrder.Fio,
			RestaurantID:      pbOrder.RestaurantId,
			TotalPrice:        pbOrder.TotalPrice,
			Status:            pbOrder.Status,
			CreatedAt:         pbOrder.CreatedAt.AsTime(),
			UpdatedAt:         pbOrder.UpdatedAt.AsTime(),
			EstimatedDelivery: pbOrder.EstimatedDelivery.AsTime(),
		}

		items := make([]model.Item, 0, len(pbOrder.Items))
		for _, pbItem := range pbOrder.Items {
			items = append(items, model.Item{
				FoodID:   "",
				Name:     pbItem.Name,
				Quantity: int(pbItem.Quantity),
				Price:    int(pbItem.Price),
			})
		}
		order.Items = items

		if pbOrder.Address != nil {
			order.Address = model.DeliveryAddress{
				Street:    pbOrder.Address.Street,
				House:     pbOrder.Address.House,
				Apartment: pbOrder.Address.Apartment,
				Floor:     pbOrder.Address.Floor,
				Comment:   pbOrder.Address.Comment,
			}
		}

		orders = append(orders, order)
	}

	return orders, nil
}
