package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gugazimmermann/go-grpc-ecomm-go/ecommpb/ecommpb"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	fmt.Println("Starting Client...")
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer cc.Close()
	cl := ecommpb.NewEcommServiceClient(cc)
	categoriesMenu(cl)
}

func categoriesMenu(cl ecommpb.EcommServiceClient) {
	fmt.Println("Reading CategoriesMenu")
	res, err := cl.CategoriesMenu(context.Background(), &emptypb.Empty{})
	if err != nil {
		fmt.Printf("Error while reading the categories menu: %v\n", err)
	}
	fmt.Printf("Categories: %v\n", res)
}
