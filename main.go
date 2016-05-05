package main

import (
  "fmt"
  "github.com/mauricio/gocached/store"
)

func main() {
  items := store.New()
  items.Put("me", nil)

  fmt.Println("We're going out!")
}
