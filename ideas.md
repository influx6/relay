#Controller

main idea is to have constructable structs

```go
type admin struct{

}

admin := &admin{}

func(a *admin) Login(req *http.Request){
}


AdminControllers.Route('/admin/:id',admin.Login)

Requests Control:
 Requests.Bind(AdminControllers)

http.ListenAndServe(":8080",Requests)

```
