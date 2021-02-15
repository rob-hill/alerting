package main

import (
    "time"
    "fmt"
    "strings"
    "strconv"
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
fmt.Printf("fetching first url : %s\n", url+query)
resp, err = client.Do(req)
handleError(err)

defer resp.Body.Close()
body, err := ioutil.ReadAll(resp.Body)
//fmt.Println(string(body))

// put our json data in here
var obj AlertList

// unmarshall it
err = json.Unmarshal([]byte(body), &obj)
if err != nil {
    fmt.Println("error:", err)
}


csv_data :=  "AlertId,Alias,TinyId,Message,Status,IsSeen,Acknowledged,Snoozed,CreatedAt,UpdatedAt,Count,Owner,Teams\n" 
csv_data = csv_data + compose_csv(obj)


// pull out the next url and keep fetching each page until we hit the last page

for  {

    if obj.Paging.Next == "" {
      break
    }

    //fmt.Printf("fetching next url : %s\n", obj.Paging.Next)

    req, err = http.NewRequest("GET", obj.Paging.Next, nil)
    handleError(err)

    // add the authorization header
    req.Header.Add(authheader, authkey)
    resp, err = client.Do(req)
    handleError(err)

    defer resp.Body.Close()
    body, err = ioutil.ReadAll(resp.Body)

    // clear out the last alert
    obj = AlertList{}
    // unmarshall it
    err = json.Unmarshal([]byte(body), &obj)
    if err != nil {
        fmt.Println("error:", err)
    }
    // add the fields to the csv
    csv_data = csv_data + compose_csv(obj)

    time.Sleep(2 * time.Second)

}

fmt.Printf(csv_data)

} // end main

// pull out all the relevant parts from the json struct and format into a single line for the csv
func compose_csv(obj AlertList) string {
  var csv_data = ""
  for _, alert := range obj.Data {
    var csv_line string = alert.ID + "," + alert.Alias + "," + alert.TinyID + ",\"" + alert.Message + "\"," + alert.Status + "," + strconv.FormatBool(alert.IsSeen) + "," + strconv.FormatBool(alert.Acknowledged) + "," + strconv.FormatBool(alert.Snoozed) + "," + alert.CreatedAt + "," + alert.UpdatedAt + "," + strconv.FormatInt(alert.Count, 10) + "," + alert.Owner + "," + alert.Teams[0].ID + "\n"
    csv_data = csv_data + csv_line
  }
  return csv_data

}

// deal with it
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


