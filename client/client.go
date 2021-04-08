package main

import (
	"context"
	"fmt"
	"log"

	. "github.com/gugazimmermann/go-grpc-ecomm-go/ecommpb/ecommpb"
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
	cl := NewEcommServiceClient(cc)
	// categoriesMenu(cl)
	// CategoryBreadcrumb(cl)
	// CategoriesSideMenu(cl)
	// Products(cl)
	// ProductsFromCategory(cl)
	SearchProducts(cl)
}

func categoriesMenu(cl EcommServiceClient) {
	fmt.Println("Reading CategoriesMenu")
	res, err := cl.CategoriesMenu(context.Background(), &emptypb.Empty{})
	if err != nil {
		fmt.Printf("Error while reading the categories menu: %v\n", err)
	}
	fmt.Printf("Categories: %v\n", res)
}

func CategoryBreadcrumb(cl EcommServiceClient) {
	// Use a valid category ID
	id := "606e5968a77732328085941d"
	fmt.Printf("Reading CategoryBreadcrumb with ID: %v\n", id)
	res, err := cl.CategoryBreadcrumb(context.Background(), &CategoryRequest{Id: id})
	if err != nil {
		fmt.Printf("Error while reading the categories breadcrumb: %v\n", err)
	}
	fmt.Printf("Categories: %v\n", res)
}

func CategoriesSideMenu(cl EcommServiceClient) {
	// Use a valid category ID
	id := "606e5968a777323280859418"
	fmt.Printf("Reading CategoriesSideMenu with ID: %v\n", id)
	res, err := cl.CategoriesSideMenu(context.Background(), &CategoryRequest{Id: id})
	if err != nil {
		fmt.Printf("Error while reading the categories sidemenu: %v\n", err)
	}
	fmt.Printf("Categories: %v\n", res)
}

func Products(cl EcommServiceClient) {
	fmt.Println("Reading Products")
	res, err := cl.Products(context.Background(), &ProductRequest{Start: 5, Qty: 10})
	if err != nil {
		fmt.Printf("Error while reading the products: %v\n", err)
	}
	fmt.Printf("Products: %v\n", res)
}

func ProductsFromCategory(cl EcommServiceClient) {
	// Use a valid category ID
	id := "606e5968a77732328085941b"
	fmt.Printf("Reading CategoryBreadcrumb with ID: %v\n", id)
	res, err := cl.ProductsFromCategory(context.Background(), &ProductFromCategoryRequest{
		CategoryId: id,
		Start:      2,
		Qty:        3,
	})
	if err != nil {
		fmt.Printf("Error while reading the products from category: %v\n", err)
	}
	fmt.Printf("Products: %v\n", res)
}

func SearchProducts(cl EcommServiceClient) {
	// Use part of a product name
	name := "drag"
	fmt.Printf("Reading SearchProducts with Name: %v\n", name)
	res, err := cl.SearchProducts(context.Background(), &SearchProductsRequest{
		Name:  name,
		Start: 0,
		Qty:   20,
	})
	if err != nil {
		fmt.Printf("Error while reading the search products: %v\n", err)
	}
	fmt.Printf("Products: %v\n", res)
}
