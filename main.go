package main

import (
  "os"
  "fmt"
  "net/http"
  "github.com/virajago/go-scs-eureka"
)

func main(){

  http.HandleFunc("/",hello)

  port := os.Getenv("VCAP_APP_PORT")

  if port == "" {
    port = "8080"
  }

  err := eureka.RegisterSCS(true)

  if err!= nil {
    fmt.Println("Error: ",err)
  }

  go eureka.SendHearbeatSCS(true)

  err = http.ListenAndServe(":"+port,nil)
  if err!=nil {
    panic(err)
  }

}

func hello(res http.ResponseWriter, req *http.Request) {
    fmt.Fprintln(res,"Hello World")
}
