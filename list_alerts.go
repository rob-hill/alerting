package main

import (
    "fmt"
    "strings"
    "io/ioutil"
    "encoding/json"
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
//fmt.Println(string(body))

// json data
var obj AlertList

// unmarshall it
err = json.Unmarshal([]byte(body), &obj)
if err != nil {
    fmt.Println("error:", err)
}

// can access using struct now
fmt.Printf("Next URL : %s\n", obj.Paging.Next);


} // end main

func handleError(err error) {

  if err != nil {
        fmt.Print(err.Error())
        os.Exit(1)
    }

}


type AlertList struct {
    Data []struct {
      Acknowledged bool   `json:"acknowledged"`
      Alias        string `json:"alias"`
      Count        int64  `json:"count"`
      CreatedAt    string `json:"createdAt"`
      ID           string `json:"id"`
      Integration  struct {
        ID   string `json:"id"`
        Name string `json:"name"`
        Type string `json:"type"`
      } `json:"integration"`
      IsSeen         bool   `json:"isSeen"`
      LastOccurredAt string `json:"lastOccurredAt"`
      Message        string `json:"message"`
      Owner          string `json:"owner"`
      OwnerTeamID    string `json:"ownerTeamId"`
      Priority       string `json:"priority"`
      Report         struct {
        AckTime        int64  `json:"ackTime"`
        AcknowledgedBy string `json:"acknowledgedBy"`
        CloseTime      int64  `json:"closeTime"`
        ClosedBy       string `json:"closedBy"`
      } `json:"report"`
      Responders []struct {
        ID   string `json:"id"`
        Type string `json:"type"`
      } `json:"responders"`
      Seen    bool     `json:"seen"`
      Snoozed bool     `json:"snoozed"`
      Source  string   `json:"source"`
      Status  string   `json:"status"`
      Tags    []string `json:"tags"`
      Teams   []struct {
        ID string `json:"id"`
      } `json:"teams"`
      TinyID    string `json:"tinyId"`
      UpdatedAt string `json:"updatedAt"`
    } `json:"data"`
    Paging struct {
      First string `json:"first"`
      Last  string `json:"last"`
      Next  string `json:"next"`
    } `json:"paging"`
    RequestID string  `json:"requestId"`
    Took      float64 `json:"took"`
}


