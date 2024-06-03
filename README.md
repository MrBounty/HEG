# HEG Tech Stack
Built Interactive and Performant Webapp in Simplicity without JavaScript (Note: WIP)

![alt text](https://github.com/MrBounty/HEG/blob/main/HEG_logo.png)

# Introduction
You might feel overwhelmed by the variety of frameworks and libraries available to do frontend. Even performing some basic task seems to require complex frameworks, enormous libraries, and extensive boilerplate code.

In this landscape, JavaScript is unavoidable. Whether you like it or not, mastering a JavaScript framework like React or Angular is often seen as essential. However, for those who prefer simplicity and minimalism, there is an alternative: HTMX.

*Note: JavaScript is fine, I just don't like big frameworks looking mandatory. Or more precisely the overall unnecessary hidden complexity behind them.*

# Stack Philosophy: Simplicity
The HEG tech stack is built on simplicity. My definition of simplicity is something you understand almost instantly, even months or years later. This philosophy guided my choice of technologies and how I build and organise my app.

# Agenda

- The HEG Tech Stack
  - Go
  - EdgeDB
  - HTMX
- Go + EdgeDB
  - Starting a Database
  - Defining type
  - Fetching data
  - Authentification
  - Managing the EdgeDB client
- Go + HTMX
  - Template
  - HTMX or JS ?
  - Using HTMX
  - Using JS
- Deployment
  - Dockerize the app
  - Create an app on fly.io
  - Deploy update
  - Monitor
  - Scale
- Performance
- Other techs

# The HEG Tech Stack

## Go (Golang)
Go, developed by Google, embodies simplicity and productivity so how can't I choose it? Its fast compilation times make it ideal for rapid development. While not the best in any specific category, Go excels in overall performance and ease of use. [Learn more](https://go.dev/)

## EdgeDB
EdgeDB is an open-source modern database designed with developer experience in mind. Its queries integrate seamlessly with Go, using similar data structures and types. EdgeDB simplifies complex SQL queries and offers numerous quality-of-life features with a UI to explore and query data from a browser. (And a cloud service) [Learn more](https://www.edgedb.com/)

## HTMX
HTMX allows HTML attributes to send requests to the server that return HTML and use it to update parts of the page. This reduces the need for complex JavaScript frameworks for apps with small to medium interactivity. While HTMX does not eliminate JavaScript entirely, it significantly reduces the amount required by removing the need of a framework. [Learn more](https://htmx.org/)

# Go + EdgeDB

## Starting a Database
First thing to do is to follow the [EdgeDB quickstart](https://docs.edgedb.com/get-started/quickstart). It will install necessary stuff. You sould end up with something like that:
```
├── edgedb.toml
├── dbschema
│   ├── default.esdl
│   ├── migrations
```

You can either create a local db or use a cloud free tier to start developping.

## Defining Type
You define type in `default.esdl` like that:
```esdl
type User {
    name: str;
    email: str;
    avatar: str;
    setting: Setting; # This is a link to another Setting type
}
```
[Learn more about EdgeDB schema](https://docs.edgedb.com/database/datamodel)
[Learn more about EdgeDB type](https://docs.edgedb.com/database/datamodel/objects)

Here a more advance example if default value:
```esdl
type Conversation {
	required name: str;
	required position: int32;
	required selected: bool {
	    default := false; 
	};
	required user: User {
	    on target delete delete source; # When the user is delete, delete the conv
	};
	required date: datetime {
	    default := datetime_current();
	}
}
```

You also define a type in go in a similar way.
```go
type User struct {
    ID       edgedb.UUID `edgedb:"id"`
    Email    string      `edgedb:"email"`
    Name     string      `edgedb:"name"`
    Avatar   string      `edgedb:"avatar"`
    Setting  Setting     `edgedb:"setting"`
}
```

## Fetching data
Get User:
```go
var user User
edgeClient.QuerySingle(context.TODO(), `
SELECT User {
    name,
    email
  } 
FILTER .name = "Adrien";
`, &user)
```

*Note that I only get the name and email, so other value will be empty. Like Avatar is ""*

Get an array of Message with other type inside.
```go
var Messages []Message
err = edgeClient.Query(context.TODO(), `
SELECT Message {
    id,
    selected,
    role,
    content,
    date,
    llm : {
        name,
        modelInfo : {
            modelID,
            name,
            company : {
                icon
            }
        }
    }
  } 
FILTER .conversation.user = global currentUser
ORDER BY .date ASC
`, &Messages)
```

*Note that .conversation.user, meaning 2 relationship! It would be a nightmare to do in SQL.*

*Note: I use the Django template because I like it but the go and other one are good too. [Choose the one you prefer.](https://docs.gofiber.io/guide/templates/)* [Or learn more abour Django template](https://docs.djangoproject.com/en/5.0/ref/templates/language/)

## Authentification
EdgeDB have a built-in auth UI. You can configure it in the auth section of the EdgeDB UI. This is the flow:

1. The user click on a button and is redirect to `/signin`.
2. `/signin` generate a PKCE and verifier. It save the verifier in a cookie and redirect to the built-in UI of your database with the PKCE.
3. The user can choose between different auth provider like google or github.
4. The user is redirect to `callbackSignup` if unknow (first time login) or `/callback` if the user is know.
5. The user know have a cookie that you need to rename (by default `my-cookie-name-auth-token` in the example).
6. Every request that the user make, use the cookie to authenticate the user on the global EdgeDB client see next section.
7. To signout, delete the cookie.

Add those route using the 'authentification_example.go' file. 
```go
app.Get("/signin", handleUiSignIn)
app.Get("/signout", handleSignOut)
app.Get("/callback", handleCallback)
app.Get("/callbackSignup", handleCallbackSignup)
```

You also need to enable the auth extension of EdgeDB and create a global currentUser.

```esdl
using extension auth;

module default {
    global currentUser := (
        assert_single((
        select User
        filter .identity = global ext::auth::ClientTokenIdentity
        ))
    );
```

*Note: You will see nect section how to use `currentUser`
[Learn more](https://docs.edgedb.com/guides/auth)

## Managing the EdgeDB client
In `EdgeDatabase.go` first define all your types that are also defined in EdgeDB. And then I do a `init` function to create two global variable `edgeGlobalClient` and `edgeCtx`.

```go
var edgeCtx context.Context
var edgeGlobalClient *edgedb.Client

func init() {
	var ctx = context.Background()
	client, err := edgedb.CreateClient(ctx, edgedb.Options{})
	if err != nil {
		fmt.Println("Error connecting to edgedb")
		panic(err)
	}

	edgeCtx = ctx
	edgeGlobalClient = client
}
```

You can then call the global client from anywhere:
```go
var user User
edgeGlobalClient.QuerySingle(edgeCtx, `
SELECT User {
    name,
    email
  } 
FILTER .name = "Adrien"
LIMIT 1;
`, &user)
```

If you use authentification, you need to provide the auth cookie to get access to the global `currentUser`:
```go
var lastArea Area
err := edgeGlobalClient.WithGlobals(map[string]interface{}{"ext::auth::client_token": c.Cookies("my-auth-token")}).QuerySingle(edgeCtx, `
SELECT global currentUser;
`, &lastArea)
```

You can then use the global `currentUser` anywhere in your query. For example:
```edgeql
SELECT User {
    name,
    email
  } 
FILTER .friend = global currentUser
```

And that's it, you can now query aywhere in your app and every request is potentialy authentify, so no risk of auth crossing of stuffs like that. And don't worry about performance, it is think to be use like that, it cost nothing.

*Note: A function named init in go in any file will be run one time at the beguinning.*

# Go + HTMX + JS

Ok now that we have setup everything, we need to start building the app. Idk what you are trying to build but I will try to give my idea of it.

## Template

## HTMX or JS
You will need to ask yourself that a lot because HTMX isn't good for everything (JS either but they don't want to admite it).

So how do I choose what to use ? UX. UX, UX, UX. At the end of the day, the user don't care what tech you use, how it's done or anything. The user care ony about the User Experience. You need to ask yourself:

*Is this action can take 0.2s to load or it need to be instant to prevent building frustration ?*

If yes, take HTMX, if no, take JS.  
For example a popover button. It need to be display instantly... TODO

### The Five Ws

I like to imagine each element using HTMX answering this:
- Where ? *The route to get the HTML*
- Who ? *The element that will be affected*
- When ? *The trigger*
- With ? *The data attach*
- What ? *The place to put the response*

For more information, please check the HTMX docs, it is well build, easy and fast to read. [Link](https://htmx.org/docs/)

### Where

Send a POST request
```html
<button
hx-post="/clicked
hx-trigger="click">
Click me!
</button>
```

This will send a POST request to `/clicked` and use the response. By default replacing it inside with the response. You can alse use `hx-get`, `hx-delete`, `hx-put`, `hx-patch`.

### Who

The target of the request is by defaut the element making the request

## Templates
You can generate HTML directly from Go passing your types as input and send it to the client.  
A tipical route is: update the database -> retrieve data from database -> generate HTML -> send HTML

Here's a simple example of using Go templates to loop over all `Items` and display a list of `Item.Name`:
```html
<ul>
    {% for item in Items %}
    <li>{{ item.Name }}</li>
    {% endfor %}
</ul>
```

### When
TODO

### With
TODO

### What
TODO

# Deployment
For the deployment you can use anything as it is just a docker container [Dockerfile](https://github.com/MrBounty/HEG/edit/main/Dockerfile). It is a really small one too, for example my app [JADE](https://jade.bouvai.com) is a 31MB container that run perfectly on a 1 shared CPU and 256MB or RAM on fly.io. So the hosting part isn't an issue as it would cost near to nothing on any cloud platform.

I personally use [fly.io](fly.io) because it is perfect for me. It do exactly what I want and nothing more, it is easy to use and keep the same philosophy as HEG.

TODO: Tuto to deploy an app on fly

# Performance
TODO: Need data with more users

# Aditional tech
- [Stripe](https://stripe.com/) for payment

### Frontend toolkit
- [Bulma](https://bulma.io/)
- [Bootstrap](https://getbootstrap.com/)
- [Tailwind CSS](https://tailwindcss.com/)


