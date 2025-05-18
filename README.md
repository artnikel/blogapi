# BlogAPI

BlogAPI is a RESTful API for managing blogs, built with Go, Echo framework, and PostgreSQL database.

## Installation

```bash
# Clone the repository
git clone https://github.com/artnikel/blogapi.git
cd blogapi

# Up Docker containers
make start
```

## Environment Variables

To run the application, export environment variables:

```
BLOG_POSTGRES_PATH="postgres://bloguser:blogpassword@localhost:5432/blogdb"
BLOG_TOKEN_SIGNATURE="blogsignature"
BLOG_SERVER_PORT="8080"
BLOG_POSTGRES_DB="blogdb"
BLOG_POSTGRES_USER="bloguser"
BLOG_POSTGRES_PASSWORD="blogpassword"
```


The API will be available at: `http://localhost:8080`

Service can be stopped 
```bash
make stop
```
or restarted.
```bash
make restart
```

## API Endpoints

### Authentication:

* `POST /signup` — Register a new user
* `POST /signupadmin` — Register a new admin (JWT token required)
* `POST /login` — User login
* `POST /refresh` — Refresh JWT token
* `DELETE /user/:id` — Delete a user (JWT token required)

### Blogs (JWT token required):

* `POST /blog` — Create a new blog 
* `GET /blog/:id` — Get blog by ID 
* `PUT /blog` — Update blog information 
* `DELETE /blog/:id` — Delete blog by ID 
* `DELETE /blogs/user/:id` — Delete all blogs by user ID 
* `GET /blogs` — Get all blogs 
* `GET /blogs/user/:id` — Get all blogs by user ID 


## Testing

To run tests with Dockertest and Go:

```bash
make test
```

## Linter


```bash
make lint
```


---

BlogAPI provides a simple and secure way to manage blogs through RESTful API endpoints. 
