# HEG Tech Stack
Built low cost webapp in simplicity without JavaScript

![alt text](https://github.com/MrBounty/HEG/blob/main/HEG_logo.png)

# Introduction
You might feel overwhelmed by the variety of frameworks and libraries available to do frontend. Even performing some basic task seems to require complex frameworks, enormous libraries, and extensive boilerplate code.

In this landscape, JavaScript is unavoidable. Whether you like it or not, mastering a JavaScript framework like React or Angular is often seen as essential. However, for those who prefer simplicity and minimalism, there is an alternative: HTMX.

*Note: JavaScript is fine, I just don't like it being mandatory.*

# Stack Philosophy: Simplicity
The HEG tech stack is built on simplicity. My definition of simplicity is: **code/structure that can be understood without documentation and minimal comments in less than like 20 seconds** but that's for code, here we are talking tech stack but this philosophy guided my choice of technologies.

# The HEG Tech Stack

## Go (Golang)
Go, developed by Google, embodies simplicity and productivity so how can't I choose it? Its fast compilation times make it ideal for rapid development. While not the best in any specific category, Go excels in overall performance and ease of use. [Learn more](https://go.dev/)

## EdgeDB
EdgeDB is an open-source modern database designed with developer experience in mind. Its queries integrate seamlessly with Go, using similar data structures and types. EdgeDB simplifies complex SQL queries and offers numerous quality-of-life features with a UI to explore and query data from a browser. (And a cloud service) [Learn more](https://www.edgedb.com/)

## HTMX
HTMX allows HTML attributes to send requests to the server that return HTML and use it to update parts of the page. This reduces the need for complex JavaScript frameworks for apps with small to medium interactivity. While HTMX does not eliminate JavaScript entirely, it significantly reduces the amount required by removing the need of a framework. [Learn more](https://htmx.org/)

# EdgeDB + Go

Go and EdgeDB manage data in similar way, through the usage of type and struct. Making it very easy to manipualte data.  

### Defining a Type in EdgeDB
```esdl
type User {
    name: str;
    email: str;
    avatar: str;
    setting: Setting; # This is a link to another Setting type
}
```
[Learn more about EdgeDB type](https://docs.edgedb.com/database/datamodel/objects)

### Defining a Type in GO
```go
type User struct {
    ID       edgedb.UUID `edgedb:"id"`
    Email    string      `edgedb:"email"`
    Name     string      `edgedb:"name"`
    Avatar   string      `edgedb:"avatar"`
    Setting  Setting     `edgedb:"setting"`
}
```

## Links
EdgeDB support links for easy relationship. For example, a `User` having a unique `Setting` or a `Cart` with multiple `Item`.

[Learn more about links](https://docs.edgedb.com/database/datamodel/links)

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

*Note: I use the Django template because I like it but the go and other one are good too. [Choose the one you prefer.](https://docs.gofiber.io/guide/templates/)* [Or learn more abour Django template](https://docs.djangoproject.com/en/5.0/ref/templates/language/)

## Authentification
EdgeDB have a built-in authentification. You can use commun provider like Google, Github, Apple, email and more

[Learn more](https://docs.edgedb.com/guides/auth)

## Deployment
You can use Google Cloud run, it's ok, I tried. But I highly recommand [fly.io](https://fly.io/) because of how efficient Go is.  
You can easily run a decently optimize Go app with just a 1 shared + 256 MB of RAM. Making it the cheapest option available at just $1.94/mo! So if in average a user spend 4h per day, you would pay around 0.30$/mo!
To compare with a flask app running on cloud run. Idk if I set up it wrong, but I would pay like 20x this, if not more. 

Here the docker file I use to build the image (TODO Add file link)

# Aditional tech
Obviously you need more tech for specific things. Here a list of some usefull one.

- [Stripe](https://stripe.com/) for payment

### Frontend toolkit
- [Bulma](https://bulma.io/)
- [Bootstrap](https://getbootstrap.com/)
- [Tailwind CSS](https://tailwindcss.com/)

# Examples

## Chatbot
(TODO: Add example)

## Interactive table
(TODO: Add example)


