package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	authToken       = "466536633:AAEOHtd13X2-VINpprfbUB6T8rGGgjewq6E"
	apiURL          = "https://api.telegram.org/bot" + authToken
	ngrokURL        = "https://f914a605.ngrok.io"
	s1              = spot{Name: "A21", Entrance1: 1, Entrance2: 4, Status: 1} //initially all status is available
	s2              = spot{Name: "A22", Entrance1: 2, Entrance2: 3, Status: 0} //0 available , 1 not , 2 pending
	s3              = spot{Name: "A23", Entrance1: 3, Entrance2: 2, Status: 2}
	s4              = spot{Name: "A24", Entrance1: 4, Entrance2: 1, Status: 0}
	p               = parking{Spots: make(map[int]spot)}
	nodeResp        = nodeResponse{IsPending: false, Pending: ""}
	session         = map[int64]userSession{} //map[chatID]userSession
	initLock        sync.Mutex
	pendingLock     sync.Mutex
	apiResponseLock sync.Mutex //locks when telegram is trying to send a user a response back
	helpMessage     = "Use\n '/start': to initiate SayeS' service \n '/bye' , '/end': to end your session" +
		"\n '/reserve': to reserve a spot \n '/cancel': to cancel a reservation you've already made" +
		"\n '/another': to reserve a spot other than one you already declined or reserved and cancelled"
)

type userSession struct {
	VacancyFound     int  //id of spot (if > 0 then we found one)
	AwaitingEntrance bool //true if we r waiting for a number from the user
	Entrance         string
	Reserved         bool    //true if user already reserved a spot
	UnwantedSpots    [4]bool //if unwanted, unwantedSpts[id] = true
}

type parking struct {
	Spots map[int]spot
}

type spot struct {
	Name      string `json:"name"`
	Entrance1 int    `json:"entrance_1"`
	Entrance2 int    `json:"entrance_2"`
	Status    int    `json:"status"`
}

type webhook struct {
	URL string `json:"url"`
}

type user struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	IsBot     bool   `json:"is_bot"`
}
type chat struct {
	ID int64 `json:"id"`
}

type telegramResponse struct {
	Text   string `json:"text"`
	ChatID int64  `json:"chat_id"`
}

type nodeResponse struct {
	IsPending bool   `json:"is_pending"`
	Pending   string `json:"pending_spots"` //comma-seperated pending spot IDs
}

type message struct {
	MessageID int    `json:"message_id"`
	Text      string `json:"text"`
	From      user   `json:"from"`
	Date      int    `json:"date"`
	Chat      chat   `json:"chat"`
}

type update struct {
	UpdateID      uint32  `json:"update_id"`
	Message       message `json:"message"`
	EditedMessage message `json:"edited_message"`
}
type chatAction struct {
	ChatID int64  `json:"chat_id"`
	Action string `json:"action"`
}

func main() {
	initializeParking()
	setWebhook()
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/updateSpots", updateSpots)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8443" // || "443" || "80" || "88"
	}
	fmt.Println("Server up and running on port : " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func initializeParking() {
	initLock.Lock()
	defer initLock.Unlock()
	p.Spots[1] = s1
	p.Spots[2] = s2
	p.Spots[3] = s3
	p.Spots[4] = s4
}

func updateStatus(id int, status int) {
	pendingLock.Lock()
	defer pendingLock.Unlock()
	switch id {
	case 1:
		s1.Status = status
		break
	case 2:
		s2.Status = status
		break
	case 3:
		s3.Status = status
		break
	case 4:
		s4.Status = status
		break
	}
	initializeParking()
}

func updateSpots(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var (
			spots []string
		)

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Println("error parsing body in updateSpots request")
			return
		}
		updatedSpots := string(b)
		spots = strings.Split(updatedSpots, ",")

		for _, spot := range spots {
			IDToStatus := strings.Split(spot, ":")
			switch IDToStatus[0] {
			case "1":
				s1.Status = customAtoi(IDToStatus[1])
				break
			case "2":
				s2.Status = customAtoi(IDToStatus[1])
				break
			case "3":
				s3.Status = customAtoi(IDToStatus[1])
				break
			case "4":
				s4.Status = customAtoi(IDToStatus[1])
				break
			}
			/*now that we have updated s1, s2, s3, s4, re-initialize parking*/
			initializeParking()
		}

		// resp := &nodeResponse{IsPending: isPending, Pending: pending}
		/*respond with pending array then clear it*/
		b, err = json.Marshal(&nodeResp)
		if err != nil {
			fmt.Println("error encoding json in updating spots response")
			return
		}
		_, err = w.Write(b)
		if err != nil {
			fmt.Println("error writing back response")
			return
		}
		/*clear nodeResp after sending it to the arduino node*/
		nodeResp = nodeResponse{IsPending: false, Pending: ""}
	}
}

func customAtoi(str string) int {
	var res int
	err := fmt.Errorf("err")
	for err != nil {
		res, err = strconv.Atoi(str)
	}
	return res
}

func setWebhook() {
	webhook := webhook{URL: ngrokURL}
	b, err := json.Marshal(&webhook)
	if err != nil {
		fmt.Println("error encoding webhook into JSON")
		return
	}

	temp := fmt.Errorf("webhook not set")
	reader := bytes.NewReader(b)
	if err != nil {
		fmt.Println("error reading byte[] of webhook", reader)
		return
	}
	for temp != nil {
		resp, err := http.Post(apiURL+"/setWebhook", "application/json", reader)
		temp = err
		defer resp.Body.Close()
	}

}

func decodeUpdate(body io.Reader) (update, error) {
	decoder := json.NewDecoder(body)
	var resp update
	if err := decoder.Decode(&resp); err != nil {
		if err.Error() != "EOF" {
			fmt.Println("error decoding update JSON")
			return resp, err
		}
	}
	return resp, nil
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		resp, err := decodeUpdate(r.Body) //resp is an update struct
		if err != nil {
			fmt.Println("error decoding update json")
			return
		}
		/*make bot seem like it's typing*/
		action := chatAction{ChatID: resp.Message.Chat.ID, Action: "typing"}
		b, err := json.Marshal(&action)
		if err != nil {
			fmt.Println("error encoding webhook into JSON")
			return
		}

		reader := bytes.NewReader(b)
		if err != nil {
			fmt.Println("error reading byte[] of webhook")
			return
		}

		http.Post(apiURL+"/sendChatAction", "application/json", reader)

		resMsg, chatID := handleUpdate(resp)
		sendMessage(resMsg, chatID)
	}

}

func sendMessage(msg string, id int64) {
	apiResponseLock.Lock()
	defer apiResponseLock.Unlock()
	res := &telegramResponse{Text: msg, ChatID: id}
	encRes, err := json.Marshal(res)
	if err != nil {
		fmt.Println("error encoding response message")
		return
	}
	reader := bytes.NewReader(encRes)
	http.Post(apiURL+"/sendMessage", "application/json", reader)
}

func handleUpdate(update update) (string, int64) {
	var (
		messageText string
		respText    string
	)

	messageText = update.Message.Text
	if update.EditedMessage.Text != "" {
		messageText = update.EditedMessage.Text
	}
	messageText = strings.ToLower(messageText)

	chatID := update.Message.Chat.ID

	_, idFound := session[chatID]
	if !idFound {
		session[chatID] = userSession{}
	}
	s, _ := session[chatID]

	fmt.Printf("chatID: %d, msg text: %s \n", chatID, update.Message.Text)
	fmt.Println("debug:: session data1: ", session[chatID])
	//special case
	if strings.Contains(messageText, "thank") {
		respText = "My Pleasure."
	}

	if strings.Contains(messageText, "help") {
		return helpMessage, chatID
	}

	if strings.Contains(messageText, "cancel") {
		if s.Reserved {
			spotName := p.Spots[s.VacancyFound].Name
			s.Reserved = false
			s.UnwantedSpots[s.VacancyFound-1] = false
			updateStatus(s.VacancyFound, 0)

			s.UnwantedSpots[s.VacancyFound-1] = true
			s.VacancyFound = 0
			session[chatID] = s

			return "Reservation of spot " + spotName + ` has been removed. You can still reserve a different spot using the '/another'` +
				`command or any spot using the '/reserve' command`, chatID
		}
	}

	if strings.Contains(messageText, "bye") || strings.Contains(messageText, "end") {
		delete(session, chatID)
		return "Goodbye, then.", chatID
	}

	if s.Reserved {
		if p.Spots[s.VacancyFound].Status != 2 { //no longer pending
			fmt.Println("no longer pending")
			/* either pending expired so it's available OR it's been taken by sb else so it's not available */
			/* thus, reset all variables*/
			var unwanted [4]bool
			s.Reserved = false
			s.AwaitingEntrance = false
			s.UnwantedSpots = unwanted
			s.Entrance = ""
			s.VacancyFound = 0
			session[chatID] = s
		} else { //still pending
			return respText + "There's still a spot reserved for you", chatID
		}
	}

	if s.VacancyFound > 0 && s.Entrance != "" {
		/*we found a vacancy and are waiting for a confirmation that he wants it*/
		if strings.Contains(messageText, "yes") {
			respText = `Reserved spot: ` + p.Spots[s.VacancyFound].Name +
				` for you. Please note that reservation holds for 1 minute. Afterwards,` +
				`spot is made available for others.`

			s.Reserved = true
			updateStatus(s.VacancyFound, 2)

			if nodeResp.IsPending { /*someone else also reserved a place and we haven't sent back the response to the node*/
				nodeResp = nodeResponse{IsPending: true, Pending: nodeResp.Pending + "," + strconv.Itoa(s.VacancyFound)}
			} else {
				nodeResp = nodeResponse{IsPending: true, Pending: strconv.Itoa(s.VacancyFound)} //no comma at the beginning
			}
			//close conversation
		} else if strings.Contains(messageText, "no") {
			respText = "Well, if you wish to reserve a different spot, use the '/another' command. If you want to reserve any spot, use the '/reserve' command"
			//save to unwanted spots and find another spot
			s.UnwantedSpots[s.VacancyFound-1] = true
			//leave entrance as is
			s.AwaitingEntrance = false
			s.Reserved = false
			s.VacancyFound = 0
			session[chatID] = s
		} else {
			respText += "Sorry, I don't understand. Is this a yes or a no?"
		}

		session[chatID] = s
		fmt.Println("debug:: session data2: ", session[chatID])

		return respText, chatID
	}

	//reserve is used to reserve and ignore previous unwanted if any
	//another is used to not use unwanted
	if strings.Contains(messageText, "another") {
		if s.Entrance != "" { //already specified an entrance
			e, err := strconv.Atoi(s.Entrance)
			if err != nil {
				return "Internal Error while decoding entrance number", chatID
			}
			available := getNearestAvailable(e, s.UnwantedSpots)
			if available > 0 {
				s.VacancyFound = available
				session[chatID] = s
				return "Next Available Spot to Entrance #" + s.Entrance + " is: " + p.Spots[available].Name + ". Do you want to reserve it?", chatID
			}
			return "There are no other available spots. To look for a spot, again, use the '/reserve' command instead of '/another'", chatID
		}
		//entrance is empty
		s.AwaitingEntrance = true
		session[chatID] = s
		return "Please specify an entrance number, first", chatID
	}

	//correct entrance but no vacancies ---> we keep checking for availability
	if s.VacancyFound == 0 && s.Entrance != "" {
		str, err := strconv.Atoi(s.Entrance)

		if err != nil {
			fmt.Printf("INVALID ENTRANCE NUMBER STORED IN SESSION[%d].ENTRANCE: %s", chatID, session[chatID].Entrance)
			return "Internal Error Occurred", chatID
		}
		//reset unwanted spots
		var unwanted [4]bool
		s.UnwantedSpots = unwanted
		available := getNearestAvailable(str, unwanted)
		if available > 0 { /*b/c all IDs start from 1*/
			respText += "Nearest Available Spot to Entrance#" + s.Entrance + " is: " + p.Spots[available].Name + ". Do you want to reserve it?"
			s.VacancyFound = available
		} else {
			respText += "There are currently no vacancies found, send me anything in a bit, I could check again."
		}

		session[chatID] = s
		fmt.Println("debug:: session data2: ", session[chatID])
		return respText, chatID
	}

	if s.AwaitingEntrance {
		var available int //stores id of nearest available spot according to entrance
		var unwanted [4]bool

		if strings.Contains(messageText, "1") {
			s.Entrance = "1"
			s.AwaitingEntrance = false

			available = getNearestAvailable(1, unwanted)

			if available > 0 { /*b/c all IDs start from 1*/
				s.VacancyFound = available
				respText += "Nearest Available Spot to Entrance #" + s.Entrance + " is: " + p.Spots[available].Name + ". Do you want to reserve it?"
			} else {
				respText += "There are currently no vacancies found, send me anything in a bit, I could check again."
			}
		} else if strings.Contains(messageText, "2") {
			s.Entrance = "2"
			s.AwaitingEntrance = false
			available = getNearestAvailable(2, unwanted)

			if available > 0 {
				s.VacancyFound = available
				respText += "Nearest Available Spot to Entrance #" + s.Entrance + " is: " + p.Spots[available].Name + ". Do you want to reserve it?"
			} else {
				respText += "There are currently no vacancies found, send me anything in a bit, I could check again."
			}

		} else {
			respText += "Invalid Entrance Number. Resend the correct entrance number."
		}

		session[chatID] = s
		fmt.Println("debug:: session data2: ", session[chatID])
		return respText, chatID
	}

	if strings.Contains(messageText, "start") || (strings.Contains(messageText, "reserve") && s.Entrance == "") {
		var unwanted [4]bool
		respText = "Hey there, please specify the entrance number you're at"
		s.AwaitingEntrance = true //this is the only field i needed to change in that case (MSLN)
		s.UnwantedSpots = unwanted
		fmt.Println("debugging: unwanted spots are ", s.UnwantedSpots)

		session[chatID] = s
		return respText, chatID
	}

	fmt.Println("HAVENT REACHED SHIT")
	if respText == "" {
		respText = "Sorry I don't understand. Type '/help' to learn about what I can do"
	}

	return respText, chatID

}

func getNearestAvailable(n int, unwantedSpots [4]bool) int {
	var available int
	if n == 1 { //entrance number
		for id, spot := range p.Spots {
			if spot.Status == 0 && !unwantedSpots[id-1] {
				if available > 0 {
					/*available has been updated before*/
					if spot.Entrance1 < p.Spots[available].Entrance1 {

						/*if the sequence of the current element is smaller than the already-available spot found, update the available*/
						available = id
					}
				} else { //this is the 1st available element we found
					available = id
				}
			}
		}
	}
	if n == 2 { //entrance2
		for id, spot := range p.Spots {
			if spot.Status == 0 && !unwantedSpots[id-1] {
				if available > 0 {
					/*available has been updated before*/
					if spot.Entrance2 < p.Spots[available].Entrance2 {
						/*if the sequence of the current element is smaller than the already-available spot found, update the available*/
						available = id
					}
				} else { //this is the 1st available element we found
					available = id
				}
			}
		}
	}

	return available

}
