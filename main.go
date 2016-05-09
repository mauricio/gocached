package main

import (
  "fmt"
  "github.com/mauricio/gocached/store"
  "github.com/mauricio/gocached/server"
)

func main() {
  items := store.New()

  server := server.New(10000, "localhost", items)
  server.Start()

  fmt.Println("Started godached server!")
  defer server.Stop()
}
