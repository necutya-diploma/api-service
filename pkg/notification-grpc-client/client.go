package notification_grpc_client

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/necutya-diploma/notification-service/pkg/grpc/gen"
)

type NotificationGrpc struct {
	addr string
	from string
}

func New(addr, from string) *NotificationGrpc {
	return &NotificationGrpc{
		addr: addr,
		from: from,
	}
}

func (n *NotificationGrpc) SendEmail(to []string, subject, body string) error {
	conn, err := grpc.Dial(n.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	defer conn.Close()

	c := pb.NewMailerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, err = c.SendEmail(ctx, &pb.EmailMessage{
		From:    n.from,
		To:      to,
		Subject: subject,
		Body:    body,
	})
	if err != nil {
		return err
	}

	return nil
}
