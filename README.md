# HEG Tech Stack
Built Interactive and Performant Webapp in Simplicity without JavaScript (Note: WIP)

![alt text](https://github.com/MrBounty/HEG/blob/main/HEG_logo.png)

# Introduction
You might feel overwhelmed by the variety of frameworks and libraries available to do frontend. Even performing some basic task seems to require complex frameworks, enormous libraries, and extensive boilerplate code.

In this landscape, JavaScript is unavoidable. Whether you like it or not, mastering a JavaScript framework like React or Angular is often seen as essential. However, for those who prefer simplicity and minimalism, there is an alternative: HTMX.

*Note: JavaScript is fine, I just don't like big frameworks looking mandatory. Or more precisely the overall unnecessary hidden complexity behind them.*

# Stack Philosophy: Simplicity
The HEG tech stack is built with simplicity and HATEOAS in mind. 

The state of applications in HEG is the HTML itself. There is no variable in JS that is then use to fill the HTML, the HTML itself IS the state.  
There is no state either on the server. When a route need data, it take it from the database.

So the only 2 places where you can store data/state are either the HTML or the Database. Everything else (JavaScript and Golang route), are just here to update those using those as input.

# Agenda

- The HEG Tech Stack
  - Go
  - EdgeDB
  - HTMX
- Naming convention
- Go + EdgeDB
  - Starting a Database
  - Defining type in EdgeDB
  - Defining type in Go
  - Fetching data
  - Authentification
  - Managing the EdgeDB client
- Go + HTMX + JS
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
- Web Assembly
- Other techs

# The HEG Tech Stack

## Go (Golang)
Go, developed by Google, embodies simplicity and productivity so how can't I choose it? Its fast compilation times make it ideal for rapid development. While not the best in any specific category, Go excels in overall performance and ease of use. [Learn more](https://go.dev/)

## EdgeDB
EdgeDB is an open-source modern database designed with developer experience in mind. Its queries integrate seamlessly with Go, using similar data structures and types. EdgeDB simplifies complex SQL queries and offers numerous quality-of-life features with a UI to explore and query data from a browser. (And a cloud service) [Learn more](https://www.edgedb.com/)

## HTMX
HTMX allows HTML attributes to send HTTP requests to the server that return HTML and use it to update parts of the page. This reduces the need for complex JavaScript frameworks for apps with small to medium interactivity. While HTMX does not eliminate JavaScript entirely, it significantly reduces the amount required by removing the need of a framework. [Learn more](https://htmx.org/)

# Examples

[Step by step simple TODO list intro](https://github.com/MrBounty/HEG-todo-example)

# Naming convention
Before everything here an idea of the naming convention I used:
- `thisIsAnExample`: For Go variable, and global EdgeDB variable
- `ThisIsAnExample`: For Go type and functions, and EdgeDB type
- `this-is-an-id`: For ids in HTML
- `this_is_an_example`: For variable in EdgeDB type

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

## Defining Type in EdgeDB
You define type in `default.esdl` like that:
```esdl
type User {
    name: str;
    email: str;
    avatar: str;
    setting: Setting; # This is a link to another Setting type
}
```

And then migrate the new schema to update the database. using the command `edgedb migration create` and then `edgedb migrate`.

[Learn more about EdgeDB schema](https://docs.edgedb.com/database/datamodel)  
[Learn more about EdgeDB migration](https://www.edgedb.com/showcase/migrations)  
[Learn more about EdgeDB type](https://docs.edgedb.com/database/datamodel/objects)

Here a more advance example with default value:
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

## Defining type in Go
You also define a type in go in a similar way. Note that we need to define all type in both EdgeDB and Go.
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
Fetching data is a 2 step process in Go. First you create an empty variable of the type to extract, then you populate it using a query. Here an example to fetch one user:
```go
var user User
edgeClient.QuerySingle(context.TODO(), `
SELECT User {
    name,
    email
  } 
FILTER .name = "Adrien"
LIMIT 1;
`, &user)
```

*Note that I only get the name and email, so other value will be empty. Like Avatar will be "" and ID 0000-0000-0000-0000*

You can also fetch an array of type and use link.
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

*Note the `.conversation.user` as filter, meaning 2 relationship! It would be a nightmare to do in SQL.*

The query language of EdgeDB is named EdgeQL, here are some links for more infos (higly recommand the interactive tutorial):  
[Learn more about EdgeQL](https://www.edgedb.com/showcase/edgeql)  
[Interactive EdgeQL tutotial](https://docs.edgedb.com/tutorial)

You can use the built-in UI of your database to add some data and run some query. Very usefull during development.

## Authentification
EdgeDB has a built-in auth UI. You can **configure it in the auth section of the EdgeDB UI**, see link below for more details. Once configured, you need to implement the flow in your app. Here the step by step:

1. The user click on a button and is redirect to `/signin`.
2. `/signin` generate a PKCE and verifier. It save the verifier in a cookie and redirect to the built-in UI of your database with the PKCE.
3. The user can choose between different auth provider like google or github. EdgeDB create a new `Identity`.
4. The user is redirect to `callbackSignup` if unknow (first time login) or `/callback` if the user is know. And a cookie that refer to the `Identity` is created.
7. Every request that the user make, use the cookie to authenticate the user on the global EdgeDB client (see next section).
8. To signout, delete the cookie.

Add those route using the 'authentification_example.go' file. 
```go
app.Get("/signin", handleUiSignIn)
app.Get("/signout", handleSignOut)
app.Get("/callback", handleCallback)
app.Get("/callbackSignup", handleCallbackSignup)
```

You also need to enable the auth extension of EdgeDB and create a global currentUser in `default.esdl`.

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

*Note: You will see next section how to use `currentUser`*  
[Learn more about Auth in EdgeDB](https://docs.edgedb.com/guides/auth)

## Managing the EdgeDB client
In `EdgeDatabase.go` first define all your types that are also defined in EdgeDB. And then do a `init` function to create two global variable `edgeGlobalClient` and `edgeCtx`.

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

If you use authentification, you need to provide the auth cookie to get access to the global variable `currentUser`:
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
FILTER .friend = global currentUser;
```

And that's it, you can now query anywhere in your app and every request is potentialy authentify, so no risk of auth crossing or stuffs like that. And don't worry about performance, it is think to be use like that, it cost nothing.

*Note: A function named init in go in any file will be run one time at the start of the server.*

# Go + HTMX + JS

Now that we have setup the database, let's see how to use HTMX and JS. I assume that you know HTML and CSS (You can use a CSS library like bulma).

## Template

*Note: I use the Django template because I like it but the go and other one are good too.*  
[Choose the one you prefer.](https://docs.gofiber.io/guide/templates/)  
[Or learn more abour Django template](https://docs.djangoproject.com/en/5.0/ref/templates/language/)

To use template with GoFiber, you can use the engine and add it to your app. Then you can use the `Render` function to render any templates. You can also use `pongo2.Template` directly.

```go
myTemplate = pongo2.Must(pongo2.FromFile("views/partials/my-template-file.html"))

out, err := myTemplate.Execute(pongo2.Context{"Text": input})
if err != nil {
	return c.SendString(err.Error())
}
```

## HTMX or JS
You will need to ask yourself that a lot because HTMX isn't good for everything (JS either but they don't want to admite it).

So how do I choose what to use ? UX. UX, UX, UX. At the end of the day, the user don't care what tech you use, how it's done or anything. The user care ony about the User Experience. You need to ask yourself:

*Can this action take 0.2s or it need to be instant to prevent building frustration ?*

If yes, take HTMX, if no, take JS.  

Like a popover button [as for example a recent Rails "drama"](https://x.com/noahflk/status/1795758603577545035). It need to be display instantly!  
Otherwise it will very quickly build a lot of frustration, and we aren't objectif when frustrated, this is something you absolutly need to avoid because users will avoid your app. It is better to remove a feature that build up frustration. 

## Using HTMX

### The Five Ws

I like to imagine each element using HTMX answering this:
- Where ? *The route to get the HTML*
- Who ? *The element that will be affected*
- When ? *The trigger*
- With ? *The data attach*
- What ? *The place to put the response*

For more information, please check the HTMX docs, it is well written, easy and fast to read. [Link](https://htmx.org/docs/)

### Where

Send a POST request
```html
<button
hx-post="/clicked
hx-trigger="click">
Click me!
</button>
```

This will send a POST request to `/clicked` and use the response. By default replacing it innerHTML with the response. You can alse use `hx-get`, `hx-delete`, `hx-put`, `hx-patch`.

### Who

The target of the request is by defaut the element making the request but you can also use any CSS selector. Here some example:
- `#my-id`: Find an element with a specific id
- `.my-class`: Find all delement with the class `my-class`
- `closest`, `next`, `previous`: To get one element from a list of element (Only in HTMX, not classics CSS selectors)
- Ect

```html
<button
hx-post="/clicked
hx-trigger="click"
hx-target="#my-id">
Click me!
</button>
```

[Learn more HTMX target](https://htmx.org/docs/#targets)  
[Learn more about CSS selectors](https://www.w3schools.com/cssref/css_selectors.php)  

### When

When to trigger the request. It is usually `click` or `submit` but it can be any events, it can also be trigger within JS.

```html
<button
hx-post="/clicked
hx-trigger="click">
Click me!
</button>
```

### With

With are the data to attach to the reauest. There is 3 ways to do it:

You can attach the data directly to the URL when you create the HTML using a template.
```html
<button
hx-post="/clicked?text={{ Text }}
hx-trigger="click">
Click me!
</button>
```

You can use `hx-include`, that take the name attribute of elements. [Learn more](https://htmx.org/attributes/hx-include/)
```html
<button hx-post="/register" hx-include="[name='email']">
	Register!
</button>
Enter email: <input name="email" type="email"/>
```

Or you can use `hx-vals` for more complexe data and use JS directly. [Learn more](https://htmx.org/attributes/hx-vals/)
```html
 <div hx-get="/example" hx-vals='js:{myVal: calculateValue()}'>
```

### What

And to finish with what to do with the response, or how to swap. For that there is 6 possibility

## Using JS
So JavaScript. I don't like it, I made it clear enough I think but like said in the intro:

*JavaScript is fine, I just don't like big frameworks looking mandatory. Or more precisely the overall unnecessary hidden complexity behind them.*

The only rule to follow is **There is no data living in JS**.

JS need to be use only to update the HTML and that's it. If at some point you have some JS variable that like contain the number of time a user clicked on a button, and you keep it in JS to use it in different JS functions. You do not do HATEOAS!

**Data can only live in the HTML or in the Database.**

I like to attach my function directly in the HTML, to keep things at the same place. You need to be carefull doing that tho, because you can easely send 2 times the same script.  
You can also just create one or multiple .js file and import it in the head of the HTML.

For example, here a function that will hide a popover menu if I click outside of it:
```html
<script>
document.addEventListener('click', function (event) {
    if (!document.getElementById('models-dropdown').contains(event.target)) {
	document.getElementById('models-dropdown').classList.remove('is-active');
    }
});
</script>
```

# Deployment
For the deployment you can use anything as it is just a docker container [Dockerfile](https://github.com/MrBounty/HEG/edit/main/Dockerfile). It is a really small one too, for example my app [JADE](https://jade.bouvai.com) is a 31MB container that run perfectly on a 1 shared CPU and 256MB or RAM on fly.io. So the hosting part isn't an issue as it would cost near to nothing on any cloud platform.

I personally use [fly.io](fly.io) because it is perfect for me. It do exactly what I want and nothing more, it is easy to use and keep the same philosophy as HEG.

TODO: Tuto to deploy an app on fly

# Performance

A big downfall of overly using the database is that the link between the database and the app need to be fast. Otherwise everything will take a will.

I recommand hosting the database and app very close to each other, preferably in on the same server. Because EdgeDB is open-source, you can host it directly on fly.io with you app or on any cloud provider.

# Web Assembly

Sometime you need something instantly but also need to use Go and not JS. That's when Web Assembly is useful.

TODO: Explain how to build the .wasm and import it to the app.

[Learn more about how to use wasm with Go](https://go.dev/wiki/WebAssembly)

# Folder structure

```
├── edgedb.toml
├── main.go
├── otherGolangFile.go
├── Dockerfile
├── dbschema
│   ├── default.esdl
│   ├── migrations
├── static
│   ├── style.css
│   ├── image.png
├── views
│   ├── layout
│   │   ├──main.html
│   ├── partials
│   │   ├──partial1.html
│   ├── page1.html
```

# Other advices
- Dog food your app (you are the user of your app)

# Aditional tech
- [Stripe](https://stripe.com/) for payment

### Frontend toolkit
- [Bulma](https://bulma.io/)
- [Bootstrap](https://getbootstrap.com/)
- [Tailwind CSS](https://tailwindcss.com/)


