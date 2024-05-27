# HEG Tech Stack
Simple webapp tech stack without JS

![alt text](https://github.com/MrBounty/HEG/blob/main/HEG_logo.png)

## Introduction
You might feel overwhelmed by the variety of frameworks and libraries available to do frontend. Even performing some basic task seems to require complex frameworks, enormous libraries, and extensive boilerplate code.

In this landscape, JavaScript is unavoidable. Whether you like it or not, mastering a JavaScript framework like React or Angular is often seen as essential. However, for those who prefer simplicity and minimalism, there is an alternative: HTMX.

## Stack Philosophy: Simplicity
The HEG tech stack is built on simplicity. My definition of simplicity is straightforward: code that can be understood without documentation and minimal comments in less than five seconds. So if I come back in 2 years, I'm not lost. If someone else read my code, it's not too difficult. If something break, I can easely identify it.

This philosophy guided my choice of technologies.

## The HEG Tech Stack

### Go (Golang)
Go, developed by Google, embodies simplicity and productivity. Its fast compilation times make it ideal for rapid development. While not the best in any specific category, Go excels in overall performance and ease of use.

### EdgeDB
EdgeDB is a modern database designed with developer experience in mind. Its queries integrate seamlessly with Go, using similar data structures and types. EdgeDB simplifies complex SQL queries and offers numerous quality-of-life features.

### HTMX
HTMX allows you to handle user inputs by sending requests to the server and updating parts of the HTML using attributes. This reduces the need for complex JavaScript frameworks for apps with small to medium interactivity. While HTMX does not eliminate JavaScript entirely, it significantly reduces the amount required.

## EdgeDB + Go

EdgeDB and Go work seamlessly together, providing a streamlined experience for defining and using data types.

### Defining a Type in EdgeDB
In EdgeDB, you define a type with its required fields and relationships like this:
```esdl
type User {
    required setting: Setting;
    required stripe_id: str;
    required name: str;
    required email: str;
    required avatar: str;
    required identity: ext::auth::Identity {
        on source delete delete target;
    }
}
```

### Defining a Type in GO
In EdgeDB, you define a type with its required fields and relationships like this:
```go
type User struct {
    ID       edgedb.UUID `edgedb:"id"`
    Setting  Setting     `edgedb:"setting"`
    StripeID string      `edgedb:"stripe_id"`
    Email    string      `edgedb:"email"`
    Name     string      `edgedb:"name"`
    Avatar   string      `edgedb:"avatar"`
}
```

### Relationships
In both EdgeDB and Go, relationships between types can be defined, such as one-to-one, one-to-many, or many-to-many. For example, a `User` type having a unique `Setting` type or a `Cart` type with multiple `Item` links.

This close alignment between EdgeDB and Go types simplifies data management and enhances developer productivity.

## Templates
Templates are the secret cherry on top of the HEG stack. Using Go, you can generate HTML directly from your types and send it to the client. This approach keeps your code clean and maintainable.

Here's a simple example of using Go templates to loop over all `Items` and display a list of `Item.Name`:
```html
<ul>
    {% for item in Items %}
    <li>{{ item.Name }}</li>
    {% endfor %}
</ul>
```

*Note: I use the Django template because I like it but the go and other one are good too. [Choose the one you prefer.](https://docs.gofiber.io/guide/templates/)*

## Deployment
Deploying your application built with the HEG stack is straightforward, as it runs as a Go server. You can deploy it almost anywhere, but I personally prefer Google Cloud with Cloud Run due to its ease of use and scalability features.

Here the docker file I use to build the image:
```dockerfile
############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache 'git=~2'

# Install dependencies
ENV GO111MODULE=on
WORKDIR $GOPATH/src/packages/goginapp/
COPY . .

# Fetch dependencies.
# Using go get.
RUN go get -d -v

# Build the binary.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/main .

############################
# STEP 2 build a small image
############################
FROM alpine:3

WORKDIR /

# Copy our static executable.
COPY --from=builder /go/main /go/main
COPY views /go/views
COPY static /go/static

ENV PORT 8080
ENV GIN_MODE release
EXPOSE 8080

WORKDIR /go

# Run the Go Gin binary.
ENTRYPOINT ["/go/main"]
```

# Examples

## Set up the project

## Chatbot
(TODO: Add example)

## Interactive table
(TODO: Add example)
