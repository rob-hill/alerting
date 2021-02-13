package main

import (
    "fmt"
    "strings"
    "io/ioutil"
    "net/http"
    "os"
)

var authheader = "Authorization"
var authkey = "" // don't want this on github
var url string = "https://api.opsgenie.com/v2/alerts/"
var query string = "?identifierType=tiny"


func main() {


// read alert id from command line
if len(os.Args) < 2 {
  fmt.Println("please supply a tinyId")
  os.Exit(1)
}

tinyId := os.Args[1]
query = tinyId + query

// read authkey from file 
data, err := ioutil.ReadFile("authkey")
if err != nil {
    fmt.Println("Cannot read ./authkey", err)
    return
}

//authkey = string(data)
authkey = strings.TrimSuffix(string(data), "\n")

// set up the http client
client := &http.Client{
}

resp, err := client.Get(url)
handleError(err)

req, err := http.NewRequest("GET", url+query, nil)
handleError(err)

// add the authorization header
req.Header.Add(authheader, authkey)
resp, err = client.Do(req)
handleError(err)

defer resp.Body.Close()
body, err := ioutil.ReadAll(resp.Body)
fmt.Println(string(body))

}


func handleError(err error) {

  if err != nil {
        fmt.Print(err.Error())
        os.Exit(1)
    }

}
