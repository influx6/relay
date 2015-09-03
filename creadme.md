Controller
Based on the idea of using this structure as below to build reusable,reactive contollers for easy web apps development.

  ```go
  //taking from http://codegangsta.gitbooks.io/building-web-apps-with-go/content/controllers/index.html
  import "net/http"

  // Action defines a standard function signature for us to use when creating
  // controller actions. A controller action is basically just a method attached to
  // a controller.
  type Action func(rw http.ResponseWriter, r *http.Request) error

  // This is our Base Controller
  type AppController struct{}

  // The action function helps with error handling in a controller
  func (c *AppController) Action(a Action) http.Handler {
      return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
          if err := a(rw, r); err != nil {
              http.Error(rw, err.Error(), 500)
          }
      })
  }

  import (
      "net/http"

      "gopkg.in/unrolled/render.v1"
  )

  type MyController struct {
      AppController
      *render.Render
  }

  func (c *MyController) Index(rw http.ResponseWriter, r *http.Request) error {
      c.JSON(rw, 200, map[string]string{"Hello": "JSON"})
      return nil
  }

  func main() {
      c := &MyController{Render: render.New(render.Options{})}
      http.ListenAndServe(":8080", c.Action(c.Index))
  }

  ```

Providing closures to controllers

```go
func MyHandler(database *sql.DB) http.Handler {
  return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
    // you now have access to the *sql.DB here
  })
}

```

Using gorillar/context for request-specific data

```go
func MyHandler(w http.ResponseWriter, r *http.Request) {
    val := context.Get(r, "myKey")

    // returns ("bar", true)
    val, ok := context.GetOk(r, "myKey")
    // ...

}
```
