package ai_grpc_client

import (
	"context"
	"math"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/necutya-diploma/ai-service/gen/go"
)

type AiGrpc struct {
	addr string
}

type CheckMessageResponse struct {
	Message          string
	IsGenerated      bool
	GeneratedPercent float32
}

func New(addr string) *AiGrpc {
	return &AiGrpc{
		addr: addr,
	}
}

func (ai *AiGrpc) CheckMessage(msg string) (string, bool, float64, error) {
	conn, err := grpc.Dial(ai.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", false, 0.0, err
	}

	defer conn.Close()

	c := pb.NewAIClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	r, err := c.CheckMessage(ctx, &pb.Message{Message: msg})
	if err != nil {
		return "", false, 0.0, err
	}
	return r.Message, r.IsGenerated, roundValueToTwoDecimalPlaces(float64(r.GeneratedPercent)), nil
}

func convertGrpcMessageResponseToCheckMessageResponse(response *pb.MessageResponse) *CheckMessageResponse {
	return &CheckMessageResponse{
		Message:          response.Message,
		IsGenerated:      response.IsGenerated,
		GeneratedPercent: response.GeneratedPercent,
	}
}

func roundValueToTwoDecimalPlaces(value float64) float64 {
	return math.Round(value*100) / 100
}
