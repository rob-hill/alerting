package main

import (
    "fmt"
    "strings"
    "io/ioutil"
    "net/http"
    "os"
)


func main() {

var authheader = "Authorization"
var authkey = "" // don't want this on github
var url string = "https://api.opsgenie.com/v2/alerts"


// read startDate, endDate from command line
if len(os.Args) < 3 {
  fmt.Println("please supply a startDate and endDate in the form DD-MM-YYYY")
  os.Exit(1)
}


startDate := os.Args[1]
endDate := os.Args[2]

var query string = "?query=createdAt%3E" + startDate + "%20AND%20createdAt%3C" + endDate + "&limit=20&sort=createdAt&order=desc"


// read authkey from file 
data, err := ioutil.ReadFile("authkey")
if err != nil {
    fmt.Println("Cannot read ./authkey", err)
    return
}

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
