package main

import (
	"encoding/json"
	"fmt"
	"log"
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
	file, err := os.OpenFile("./runlog.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
  defer file.Close()

  file.WriteString("Starting run.... StartDate and EndDate are:" + startDate + " " + endDate + "\n")


	// read authkey from file
	data, err := ioutil.ReadFile("authkey")
	if err != nil {
		fmt.Println("Cannot read ./authkey", err)
    file.WriteString("Cannot read ./authkey\n")
		os.Exit(1)
	}

	authkey = strings.TrimSuffix(string(data), "\n")

	//get the list of teams
	//
	// set up the http client
	client := &http.Client{}

	resp, err := client.Get(teams_url)
	handleError(err)

	req, err := http.NewRequest("GET", teams_url, nil)
	handleError(err)

	// add the authorization header
	req.Header.Add(authheader, authkey)
	//fmt.Printf("fetching teams url : %s\n", teams_url)
	file.WriteString("fetching teams url: "+ teams_url + "\n")
	resp, err = client.Do(req)
	handleError(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(body))

	// put our json data in here
	var teamobj Teams

	// unmarshall it
	err = json.Unmarshal([]byte(body), &teamobj)
	if err != nil {
		fmt.Println("error:", err)
    file.WriteString("error unmarshalling " + "\n")
    os.Exit(1)
	}




	//get the 'lerts
	//
	// set up the http client
	client = &http.Client{}

	resp, err = client.Get(url)
	handleError(err)

	req, err = http.NewRequest("GET", url+query, nil)
	handleError(err)

	// add the authorization header
	req.Header.Add(authheader, authkey)
	//fmt.Printf("fetching first url : %s\n", url+query)
	file.WriteString("fetching first url: "+url+query + "\n")
	resp, err = client.Do(req)
	handleError(err)

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	//fmt.Println(string(body))

	// put our json data in here
	var obj AlertList

	// unmarshall it
	err = json.Unmarshal([]byte(body), &obj)
	if err != nil {
		fmt.Println("error:", err)
    file.WriteString("error unmarshalling" + "\n")
    os.Exit(1)
	}

	csv_data := "AlertId,Alias,TinyId,Message,Status,IsSeen,Acknowledged,Snoozed,CreatedAt,UpdatedAt,Count,Owner,Teams,Priority\n"
	csv_data = csv_data + gather_data(obj)

	// pull out the next url and keep fetching each page until we hit the last page

	for {

		if obj.Paging.Next == "" {
			break
		}

		//fmt.Printf("fetching next url : %s\n", obj.Paging.Next)
		file.WriteString("fetching next url: "+obj.Paging.Next + "\n")

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
      file.WriteString("error unmarshalling" + "\n")
      os.Exit(1)
		}
		// add the fields to the csv
    file.WriteString("Composing csv." + "\n")
		csv_data = csv_data + gather_data(obj)

		time.Sleep(2 * time.Second)

	}


  // replace team ids with team names
  file.WriteString("Replacing team names...." + "\n")
  for _, teams := range teamobj.Data {
    //fmt.Printf("Team name is %s, team id is %s\n", teams.Name, teams.ID)
    csv_data = strings.Replace(csv_data, teams.ID, "\""+teams.Name+"\"", -1)
  }

	fmt.Printf(csv_data)
  file.WriteString("Done." + "\n")

} // end main




// pull out all the relevant parts from the json struct and format into a single line for the csv
func gather_data(obj AlertList) string {
	var csv_data = ""
  layout := "2006-01-02T15:04:05"
  loc, _ := time.LoadLocation("Australia/NSW")

	for _, alert := range obj.Data {
    created := alert.CreatedAt[0:19]
    updated := alert.UpdatedAt[0:19]
    var createdTime, err = time.Parse(layout, created)
    var updatedTime, err2 = time.Parse(layout, updated)
    if err != nil {
        fmt.Println(err)
    }
    if err2 != nil {
        fmt.Println(err)
    }

    // Grab alert-specific data


/*
		//fmt.Printf("fetching next url : %s\n", obj.Paging.Next)
		file.WriteString("fetching next url: "+obj.Paging.Next + "\n")

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
      file.WriteString("error unmarshalling" + "\n")
      os.Exit(1)
		}





*/





    // protect against empty Teams array
    var teamId string = ""
    if len(alert.Teams) != 0 {
      teamId = alert.Teams[0].ID
    }

		var csv_line string = alert.ID + ",\"" + alert.Alias + "\"," + alert.TinyID + ",\"" + alert.Message + "\"," + alert.Status + "," + strconv.FormatBool(alert.IsSeen) + "," + strconv.FormatBool(alert.Acknowledged) + "," + strconv.FormatBool(alert.Snoozed) + ",\"" + createdTime.In(loc).String() + "\",\"" + updatedTime.In(loc).String() + "\"," + strconv.FormatInt(alert.Count, 10) + "," + alert.Owner + "," + teamId + "," + alert.Priority + "\n"
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

