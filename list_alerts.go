package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var authheader = "Authorization"
var authkey = "" // don't want this on github
var url = "https://api.opsgenie.com/v2/alerts"
var teams_url = "https://api.opsgenie.com/v2/teams"
var logfile *os.File
var err error

// get date layout and location so we can adjust from UTC
var layout = "2006-01-02T15:04:05"
var loc, _ = time.LoadLocation("Australia/NSW")



func main() {


	// read startDate, endDate from command line
	if len(os.Args) < 3 {
		fmt.Println("please supply a startDate and endDate in the form DD-MM-YYYY")
		os.Exit(1)
	}

	startDate := os.Args[1]
	endDate := os.Args[2]

	var query string = "?query=createdAt%3E" + startDate + "%20AND%20createdAt%3C" + endDate + "&limit=20&sort=createdAt&order=desc"


  // create a log file
  // If the file doesn't exist, create it, or append to the file
  logfile, err = os.OpenFile("./runlog.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	handleError(err)
  defer logfile.Close()

  logfile.WriteString("Starting run.... StartDate and EndDate are:" + startDate + " " + endDate + "\n")


	// read authkey from file
	data, err := ioutil.ReadFile("authkey")
	if err != nil {
		fmt.Println("Cannot read ./authkey", err)
    logfile.WriteString("Cannot read ./authkey\n")
		os.Exit(1)
	}

	authkey = strings.TrimSuffix(string(data), "\n")

	//get the list of teams
  body := get_url(teams_url)

  // put our json data in here
	var teamobj Teams

	// unmarshall it
	err = json.Unmarshal([]byte(body), &teamobj)
	handleError(err)


	//get the alerts
  body = get_url(url+query)

	// put our json data in here
	var obj AlertList

	// unmarshall it
	err = json.Unmarshal([]byte(body), &obj)
	handleError(err)

  // create csv header
	csv_data := "AlertId,Alias,TinyId,Message,Status,IsSeen,Acknowledged,Snoozed,CreatedAt,UpdatedAt,Count,Owner,Teams,Priority\n"

  // pass the first page to gather_data
	csv_data = csv_data + gather_data(obj)

	// iterate over next paging url till done, passing json struct to gather_data
	for {

		if obj.Paging.Next == "" {
			break
		}

    body = get_url(obj.Paging.Next)

		// clear out the last alert
		obj = AlertList{}
		// unmarshall it
		err = json.Unmarshal([]byte(body), &obj)
    handleError(err)

		// add the fields to the csv
		csv_data = csv_data + gather_data(obj)


	}


  // replace team ids with team names
  logfile.WriteString("Replacing team names...." + "\n")
  for _, teams := range teamobj.Data {
    //fmt.Printf("Team name is %s, team id is %s\n", teams.Name, teams.ID)
    csv_data = strings.Replace(csv_data, teams.ID, "\""+teams.Name+"\"", -1)
  }

	fmt.Printf(csv_data)
  logfile.WriteString("Done." + "\n")

} // end main



// fetch the body given a url
//

func get_url(url string) ([]byte) {

	// set up the http client
	client := &http.Client{}

	resp, err := client.Get(url)
	handleError(err)

	req, err := http.NewRequest("GET", url, nil)
	handleError(err)

	// add the authorization header
	req.Header.Add(authheader, authkey)
	//fmt.Printf("fetching %s\n", url)
	logfile.WriteString("fetching "+ url + "\n")
	resp, err = client.Do(req)
	handleError(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(body))
	handleError(err)

  return body
}

// pull out all the relevant parts from the json struct and format into a single line for the csv
//
func gather_data(obj AlertList) string {

  logfile.WriteString("Composing csv." + "\n")
  var csv_data = ""
  var alertTinyID = ""

	for _, alert := range obj.Data {
    created := alert.CreatedAt[0:19] // fist 19 chars contain the datetime info we want
    updated := alert.UpdatedAt[0:19]
    var createdTime, err = time.Parse(layout, created)
    handleError(err)
    var updatedTime, err2 = time.Parse(layout, updated)
    handleError(err2)

    // protect against empty Teams array
    var teamId string = ""
    if len(alert.Teams) != 0 {
      teamId = alert.Teams[0].ID
    }

    alertTinyID = alert.TinyID
		var csv_line string = alert.ID + ",\"" + alert.Alias + "\"," + alert.TinyID + ",\"" + alert.Message + "\"," + alert.Status + "," + strconv.FormatBool(alert.IsSeen) + "," + strconv.FormatBool(alert.Acknowledged) + "," + strconv.FormatBool(alert.Snoozed) + ",\"" + createdTime.In(loc).String() + "\",\"" + updatedTime.In(loc).String() + "\"," + strconv.FormatInt(alert.Count, 10) + "," + alert.Owner + "," + teamId + "," + alert.Priority + "\n"
		csv_data = csv_data + csv_line
	}

  // Grab alert-specific data here
  url = "https://api.opsgenie.com/v2/alerts/" + alertTinyID + "?identifierType=tiny"
  var body = get_url(url)
  var details = AlertDetails{}
  // unmarshall it
  err = json.Unmarshal([]byte(body), &details)
  handleError(err)

  //Details.Backend is what we want
  var backend = details.Data.Details.Backend
  var frontend = details.Data.Details.Frontend
  var host = details.Data.Details.Host
  var class = details.Data.Details.Class
  logfile.WriteString("found backend " + backend + "\n")
  logfile.WriteString("found frontend " + frontend + "\n")
  logfile.WriteString("found host " + host + "\n")
  logfile.WriteString("found class " + class + "\n")

  //append to csv



  // sleep here as we call this function after almost all fetches
  time.Sleep(1 * time.Second)

  return csv_data

}


// deal with errors
func handleError(err error) {

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

}

// structure to hold alert data
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

// struct to hold team info
type Teams struct {
	Data []struct {
		Description string `json:"description"`
		ID          string `json:"id"`
		Links       struct {
			API string `json:"api"`
			Web string `json:"web"`
		} `json:"links"`
		Name string `json:"name"`
	} `json:"data"`
	RequestID string  `json:"requestId"`
	Took      float64 `json:"took"`
}


type AlertDetails struct {
	Data struct {
		Acknowledged bool     `json:"acknowledged"`
		Actions      []string `json:"actions"`
		Alias        string   `json:"alias"`
		Count        int64    `json:"count"`
		CreatedAt    string   `json:"createdAt"`
		Description  string   `json:"description"`
		Details      struct {
			Backend      string `json:"Backend"`
			Class        string `json:"Class"`
			Count        string `json:"Count"`
			Frontend     string `json:"Frontend"`
			Host         string `json:"Host"`
			Raw          string `json:"Raw"`
			Results_Link string `json:"Results Link"`
			Sdkb         string `json:"SDKB"`
			SdkbID       string `json:"SDKB_ID"`
			Severity     string `json:"Severity"`
			Total        string `json:"Total"`
		} `json:"details"`
		Entity      string `json:"entity"`
		ID          string `json:"id"`
		Integration struct {
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
	RequestID string  `json:"requestId"`
	Took      float64 `json:"took"`
}

