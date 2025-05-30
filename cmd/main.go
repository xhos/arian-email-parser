package main

import (
	"arian-parser/internal/email"
	"context"
	"fmt"
	"log"
)

func main() {
	ctx := context.Background()
	client, err := email.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	ids, _ := client.GetUnreadEmails(ctx)
	for _, id := range ids {
		data, _ := client.GetEmail(ctx, id)
		fmt.Println(string(data))
	}
}
