# Go gRPC Ecommerce - Pt 2 - Go

![cover](imgs/cover.png)

In this 3 articles series I will show you how to use gRPC to create an ecommerce with frontend in React and Backend in GO.

We will use React, Typescript, Keycloak, gRPC, Protocol Buffer, Docker Compose, MongoDB e Go!

# Part 2 - Backend

If you have seen my previous articles on gRPC, you certainly know how to install GO, gRPC, ProtoBuf, etc... Then I will go through these parts quickly. And since the idea of the articles is not to create a fully functional ecommerce for production, but to show the integration between React and GO, some parts will be left behind.

The files can be found here: https://github.com/gugazimmermann/go-grpc-ecomm-go

## Go Mod

First of all we need to initialize the Go Modules, chanhe the Git to your own.

`go mod init github.com/gugazimmermann/go-grpc-ecomm-go`

## MongoDB

https://www.linkedin.com/pulse/phone-book-go-grpc-protobuf-docker-mongodb-zimmermann-negreiros/

In this article you can see how to install Docker, Docker Compose and MongoDB. We will again use a bach file to setup mongo information and store in a .env file.

## Test Server

Let's start creating a test server as a base to see if everything is working.

`ecommpb/ecomm.proto`

In your proto file we will define the message for Category and Product in a way we can use it with the React page.

```protobuf

syntax = "proto3";

package ecomm;

option go_package = "ecommpb/ecommpb";

import "google/protobuf/timestamp.proto";

message Category {
  string id = 1;
  string name = 2;
  string slug = 3;
  repeated Category ancestors = 4;
  repeated Category childrens = 5;
  google.protobuf.Timestamp last_updated = 6;
}

message Product {
  string id = 1;
  string name = 2;
  string slug = 3;
  string image = 4;
  int32 quantity = 5;
  float value = 6;
  Category category = 7;
  google.protobuf.Timestamp last_updated = 8;
}

service EcommService {}

```

`genpb.sh`

```bash
#!/bin/bash

protoc ecommpb/ecomm.proto --go_out=plugins=grpc:.

```

Run `chmod +x ./genpb.sh`, now when we change the proto file we can just run this bash script.

`main.go`

In this file we will get the env file information and start the Mongo Client, start the listener and the gRPC server.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/gugazimmermann/go-grpc-ecomm-go/ecommpb/ecommpb"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"google.golang.org/grpc"
)

type server struct{}

var products, categories *mongo.Collection

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	mongoUsername := os.Getenv("MONGO_USERNAME")
	mongoPassword := os.Getenv("MONGO_PASSWORD")
	mongoDb := os.Getenv("MONGO_DB")

	mongoCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoUri := fmt.Sprintf("mongodb://%s:%s@localhost:27017", mongoUsername, mongoPassword)
	fmt.Println("Connecting to MongoDB...")
	client, err := mongo.Connect(mongoCtx, options.Client().ApplyURI(mongoUri))
	if err != nil {
		log.Fatalf("Error Starting MongoDB Client: %v", err)
	}

	products = client.Database(mongoDb).Collection("products")
	categories = client.Database(mongoDb).Collection("categories")

	fmt.Println("Starting Listener...")
	l, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	ecommpb.RegisterEcommServiceServer(s, &server{})

	go func() {
		fmt.Println("Ecomm Server Started...")
		if err := s.Serve(l); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch
	fmt.Println("Stopping Ecomm Server...")
	s.Stop()
	fmt.Println("Closing Listener...")
	l.Close()
	fmt.Println("Closing MongoDB...")
	client.Disconnect(mongoCtx)
	fmt.Println("All done!")
}

```

Run `go run main.go` and the server will start.

## Store Sample Data

We have the sample data in the React project: https://github.com/gugazimmermann/go-grpc-ecomm-react/blob/master/src/utils/sample-data.ts

Now we just use it as a JSON file: https://github.com/gugazimmermann/go-grpc-ecomm-go/blob/add-sample-data/utils/sample-data.json

`main.go`

Now we will change the main.go to load all the categories and products into MongoDB

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/gosimple/slug"
	"github.com/gugazimmermann/go-grpc-ecomm-go/ecommpb/ecommpb"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct{}

type SampleData struct {
	Categories []SampleCategory
	Products   []SampleProduct
}

type SampleCategory struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Parent int    `json:"parent"`
}

type SampleProduct struct {
	Id       int     `json:"id"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Value    float64 `json:"value"`
	Category int     `json:"category"`
	Parent   *SampleProduct
}

type FlatCategory struct {
	Name      string
	Slug      string
	Parent    string
	Ancestors []string
	Childrens []string
}

type MongoCategory struct {
	ID          primitive.ObjectID     `bson:"_id,omitempty"`
	Name        string                 `bson:"name,omitempty"`
	Slug        string                 `bson:"slug,omitempty"`
	Ancestors   []primitive.ObjectID   `bson:"ancestors,omitempty"`
	Childrens   []primitive.ObjectID   `bson:"childrens,omitempty"`
	LastUpdated *timestamppb.Timestamp `bson:"last_updated,omitempty"`
}

type FlatProduct struct {
	Name        string
	Slug        string
	Quantity    int
	Value       float64
	Category    primitive.ObjectID
	LastUpdated *timestamppb.Timestamp
}

var products, categories *mongo.Collection
var mongoCtx context.Context

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	mongoUsername := os.Getenv("MONGO_USERNAME")
	mongoPassword := os.Getenv("MONGO_PASSWORD")
	mongoDb := os.Getenv("MONGO_DB")

	mongoCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoUri := fmt.Sprintf("mongodb://%s:%s@localhost:27017", mongoUsername, mongoPassword)
	fmt.Println("Connecting to MongoDB...")
	client, err := mongo.Connect(mongoCtx, options.Client().ApplyURI(mongoUri))
	if err != nil {
		log.Fatalf("Error Starting MongoDB Client: %v", err)
	}

	products = client.Database(mongoDb).Collection("products")
	categories = client.Database(mongoDb).Collection("categories")

	sampleDataHandler()

	fmt.Println("Starting Listener...")
	l, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	ecommpb.RegisterEcommServiceServer(s, &server{})

	go func() {
		fmt.Println("Ecomm Server Started...")
		if err := s.Serve(l); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch
	fmt.Println("Stopping Ecomm Server...")
	s.Stop()
	fmt.Println("Closing Listener...")
	l.Close()
	fmt.Println("Closing MongoDB...")
	client.Disconnect(mongoCtx)
	fmt.Println("All done!")
}

func sampleDataHandler() {
	sd := getSampleData()
	cs := getFlatCategories(sd.Categories)
	mcs := handleCategories(cs)
	ps := getFlatProducts(sd, mcs)
	for _, p := range ps {
		insertProduct(p)
	}
}

func getSampleData() *SampleData {
	f, err := ioutil.ReadFile("./utils/sample-data.json")
	if err != nil {
		fmt.Print(err)
	}
	sd := &SampleData{}
	_ = json.Unmarshal(f, sd)
	return sd
}

func getFlatCategories(s []SampleCategory) []*FlatCategory {
	cs := []*FlatCategory{}
	for _, c := range s {
		fc := &FlatCategory{
			Name: c.Name,
			Slug: slug.Make(c.Name),
		}
		for _, s := range s {
			if s.Id == c.Parent {
				fc.Parent = s.Name
			}
		}
		cs = append(cs, fc)
	}
	return cs
}

func handleCategories(cs []*FlatCategory) []*MongoCategory {
	mcs := []*MongoCategory{}
	for _, c := range cs {
		findChildrens(c, cs)
		findParent(c, cs)
		findAncestors(c, cs)
		id := insertCategory(c)
		mcs = append(mcs, &MongoCategory{
			ID:          id,
			Name:        c.Name,
			Slug:        c.Slug,
			LastUpdated: timestamppb.Now(),
		})
	}
	for _, mc := range mcs {
		for _, c := range cs {
			if c.Name == mc.Name {
				if len(c.Ancestors) != 0 {
					as := []primitive.ObjectID{}
					for _, ca := range c.Ancestors {
						a := findMongoCat(ca, mcs)
						as = append(as, a.ID)
					}
					mc.Ancestors = as
				}
				if len(c.Childrens) != 0 {
					chs := []primitive.ObjectID{}
					for _, cc := range c.Childrens {
						c := findMongoCat(cc, mcs)
						chs = append(chs, c.ID)
					}
					mc.Childrens = chs
				}
			}
		}
		updateCategory(mc)
	}
	return mcs
}

func findChildrens(c *FlatCategory, cs []*FlatCategory) {
	chs := []string{}
	for _, ch := range cs {
		if ch.Parent == c.Name {
			chs = append(chs, ch.Name)
			findChildrens(ch, cs)
		}
	}
	c.Childrens = chs
}

func findParent(c *FlatCategory, cs []*FlatCategory) {
	for _, ch := range c.Childrens {
		for _, p := range cs {
			if ch == p.Name {
				p.Parent = c.Name
			}
		}
	}
}

func findAncestors(c *FlatCategory, cs []*FlatCategory) {
	a := []string{}
	if c.Parent != "" {
		a = append(a, c.Parent)
	}
	for _, ch := range c.Childrens {
		for _, ca := range cs {
			if ch == ca.Name {
				a = append(a, ca.Parent)
				ca.Ancestors = dedupeString(a)
			}
		}
	}
}

func dedupeString(e []string) []string {
	m := map[string]bool{}
	r := []string{}
	for v := range e {
		if m[e[v]] == false {
			m[e[v]] = true
			r = append(r, e[v])
		}
	}
	return r
}

func insertCategory(c *FlatCategory) primitive.ObjectID {
	r, err := categories.InsertOne(mongoCtx, c)
	if err != nil {
		fmt.Println("InsertOne ERROR:", err)
	}
	return r.InsertedID.(primitive.ObjectID)
}

func findMongoCat(n string, mcs []*MongoCategory) *MongoCategory {
	mc := &MongoCategory{}
	for _, m := range mcs {
		if m.Name == n {
			mc = m
		}
	}
	return mc
}

func updateCategory(c *MongoCategory) {
	_, err := categories.ReplaceOne(mongoCtx, primitive.M{"_id": c.ID}, MongoCategory{
		Name:        c.Name,
		Slug:        c.Slug,
		Ancestors:   c.Ancestors,
		Childrens:   c.Childrens,
		LastUpdated: c.LastUpdated,
	})
	if err != nil {
		fmt.Printf("Cannot update person: %v", err)
	}
}

func getFlatProducts(s *SampleData, m []*MongoCategory) []*FlatProduct {
	ps := []*FlatProduct{}
	for _, p := range s.Products {
		fp := &FlatProduct{
			Name:     p.Name,
			Slug:     slug.Make(p.Name),
			Quantity: p.Quantity,
			Value:    math.Ceil(p.Value*100) / 100,
		}
		var pc string
		for _, sc := range s.Categories {
			if sc.Id == p.Category {
				pc = sc.Name
			}
		}
		for _, c := range m {
			if c.Name == pc {
				fp.Category = c.ID
			}
		}
		ps = append(ps, fp)
	}
	return ps
}

func insertProduct(p *FlatProduct) {
	p.LastUpdated = timestamppb.Now()
	_, err := products.InsertOne(mongoCtx, p)
	if err != nil {
		fmt.Println("InsertOne ERROR:", err)
	}
}
```

## CSS

I'm not a designer, so to help in this project I'll use bootstrap and create a very simple layout.

`npm install react-bootstrap bootstrap node-sass`

In `public/index.html` we need to add the Bootstrap CSS.

```html
<link
  rel="stylesheet"
  href="https://cdn.jsdelivr.net/npm/bootstrap@4.6.0/dist/css/bootstrap.min.css"
  integrity="sha384-B0vP5xmATw1+K9KRQjQERJvTumQW0nPEzvF6L/Z6nronJ3oUOFUFpCjEUQouq2+l"
  crossorigin="anonymous"
/>
```

And in `src/css/` just 3 files:

- `_colors.scss`: The custom colors of the page. Fell free to change to the colors you like.
- `src/css/app.scss`: This file will import the custom colors and define your custon bootstrap theme.
- `src/css/layout.scss`: And this last file will have some css classes to adjust your page.

## Layout

In `src/App.tsx` we put the routes right bellow the layout to not have to add the layout to each page. The each page will be a child of the layout.

- `src/layout/index.tsx` is the entry point of the layout, will load the css and add the header and the footer.

- `src/layout/footer.tsx` a very simple file created just to show the credits, like who made the page and the icons that was used.

- `src/layout/header.tsx` will show the logo, the page name (from the `.env` file) and the description. The header also load the menu of the page.

- `src/layout/menu.tsx` This top menu will show the page logo, title, the categories, a search fied and the cart button.

## Router

`npm install react-router-dom react-error-boundary && npm install --save-dev @types/react-router-dom`

`src/index.tsx` we added the main Router and use the browser history.

`src/App.tsx`

First we add the ErrorBoundary and create a fallback, this will take care if some route fail to load. Suspense are added to work with Lazy, this make react load just the components that we need at the time.

This page will have very simple routes, just Category, Cart and a Not Found page if the user type a url that do not exists.

## Constants

- `src/constants/index.ts`

It's good to have constants to avoid any typing errors or / and when we have to use the value repeatedly. In our page we will use the constants for the routes, state and toast.

## Interfaces

- `src/interfaces/index.ts`

In typescript we need Interfaces for type checking, and in this file we are going to centralize them so that we don't have to repeating every time they are used. I like to put a `I` in front of the name to easily know that I'm dealing with an interface.

## API and SERVER

- `src/api/index.ts`

It will make calls to the backend server. It is good to leave it separate so we do not need to look all the frontend code when we move to the GO server.

- `src/utils/server.ts`

In this file we are going to simulate our backend server. We will receive calls from the API and respond with the data that we have in `src/utils/sample-data.ts`.

## Handling the page state

- `src/context/index.tsx`

This file will create the context and provider for the state.

- `src/context/reducers.ts`

Here we will have the reducers for the state, so the state can only be changed if done by a dispatch.

For our website we will only have to add, update and remove products from the cart, checkout, login and logout. We will also control to have a Toast appear whenever something is changed in the cart.

To use the state on the entire site we will put the provider in `src/App.tsx`, just above the layout, so on the pages that we are going to use it just need to use the context.

## Pages

- `src/pages/not-found.tsx`

Very simple page, it only shows an alert when the user tries to use a Route that does not exist.

![notfound](imgs/notfound.png)

- `src/pages/category.tsx`

This page is practically the whole site, we will practically use just this page, so each time it is called we will reset the information.

When this page is called the first thing we need to do is see if it has a search, then if we are on the homepage, and finally if is to show some category, so we can make the correct call to the server to obtain the products.

Then we will show the Breadcrumb, which can again be for search, homepage or categories. And if we are in a subcategory, we should show the parents.

If we were in a category, we should check the subcategories and creating a sidemenu with them.

We will also show the products, and to not get a long list we will add a pagination and the user can choose how many products he wants to see.

![category](imgs/category.png)

- `src/pages/cart.tsx`

This page will show the products that the user has added to the cart and allow it to change the quantity or remove it from the cart. Will also show if the user is logged with the Keycloak or not, and allow them to login / logout and checkout.

![cart](imgs/cart.png)

## Components

- `src/components/menu-nav.tsx`

The Menu navigation by categories. We will receive from the server the main categories (which do not have a parent) and for each their direct subcategories (only 1 level), and then set up the list and dropdowns.

![menu-nav](imgs/menu-nav.png)

- `src/components/menu-cart.tsx`

The button for the shopping cart. We will use the state to show a badge with the quantity of products added to the cart, and we will also show a Toast whenever a product is added, updated or removed from the cart.

![menu-cart](imgs/menu-cart.png)

- `src/components/breadcrumb.tsx`

It will show a title if we are on the homepage, or the search term and the number of results, or the category and the parents categories.

![breadcrumb](imgs/breadcrumb.png)

- `src/components/side-menu.tsx`

If you are viewing a category it will show the subcategories.

![sidemenu](imgs/sidemenu.png)

- `src/components/qty-per-page.tsx`

Allows the user to define how many products to see at a time.

![qtyperpage](imgs/qtyperpage.png)

- `src/components/product-card.tsx`

Shows the image, title and value of the product to the user, also allows the user to add the product to the cart.

![productcard](imgs/productcard.png)

- `src/components/pagination.tsx`

To not to show a very long list of products we will use pagination, this file create a list of page numbers for the user be able to browse the product list.

![pagination](imgs/pagination.png)

- `src/components/cart-list.tsx`

Shows the list of products in the cart and the total purchase price.

![cartlist](imgs/cartlist.png)

- `src/components/cart-item.tsx`

Shows the product in the cart list and allows the user to change the quantity or remove it from the list.

![cartitem](imgs/cartitem.png)

- `src/components/login-form.tsx`

Allows the user to login using the Red Hat's Keycloak.

![loginform](imgs/loginform.png)

- `src/components/logged-in.tsx`

When the user is logged in, it shows the user's information and allows him to logout or checkout the products in the cart.

![loggedin](imgs/loggedin.png)

# Red Hat's Keycloak

![dockercompose](imgs/dockercompose.png)

`docker-compose.yml`

To handle user authentication we will use Red Hat's Keycloak. To use it you need to have the docker and docker compose installed (in my previous articles I show you how to install). To start you only need to run: `docker-compose up` and then open it in the browser

http://localhost:8080/auth/admin
user: admin
pass: admin

![keycloak](imgs/keycloak.png)

Pass the mouse over `Master` and add a realm `go-grpc-ecomm-react`. Create a user, set `User Enabled` and `Email Verified` than go to the user `Credentials` and set a password and change temporary to off.

Open http://localhost:8080/auth/realms/go-grpc-ecomm-react/account and click sign-in to test the new user. After login go back to admin console and go to sessions, you will see `account-console` with 1 active session. Click `account-console` and show sessions to see the user.

Click on `Clients` -> `create` and create a client with id `react`, change `Access Type` to `confidential`, set `Standard Flow Enabled` to off, `Direct Access Grants Enabled` to on. Save and go to `Credentials` to see the Secret (change the value of `REACT_APP_KEYCLOAK_CLIENT_SECRET` in `.env` to this secret). Now set `Access Type` back to `public` (you will not see the `Credentials` tab anymore) and add `http://localhost:3000` in `Web Origins`. This will handle CORS problems.

Try to run `http://localhost:8080/auth/realms/go-grpc-ecomm-react/.well-known/openid-configuration` to see if everything is ok. It will open a JSON with many endpoints, we will use the `token_endpoint` now.

![postman](imgs/postman.png)

You can try to login using POSTMAN (https://www.postman.com). Send a POST to `http://localhost:8080/auth/realms/go-grpc-ecomm-react/protocol/openid-connect/token` In Body select `x-www-form-urlencoded` and

```
grant_type : password
client_id : react
client_secret : the client secret from client`s credentials
username : the email of the user you just created
password : the password of the user you just created
```

Hit send and you will receive a JSON like this one:

```json
{
  "access_token": "eyJhbG...",
  "expires_in": 300,
  "refresh_expires_in": 1800,
  "refresh_token": "eyJhbG...",
  "token_type": "Bearer",
  "not-before-policy": 1617433028,
  "session_state": "41252d19-fa5b-4c33-a354-a86e637685ac",
  "scope": "profile email"
}
```

If you go to the keycload admin console and click `Users -> the user id -> Sessions` will show the login info and your keycloak is working.

## Credits

The icons used in this project are from www.flaticon.com
