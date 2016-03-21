package main

import (
	"flag"
	"github.com/yfujita/monitoring-elasticsearch-fluent/slack"
	"net/http"
	"time"
	"io/ioutil"
	"errors"
	"encoding/json"
	"strconv"
)

type Options struct {
	webHookUrl    string
	channel       string
	botName       string
	botIcon       string
	redmineHost   string
	redmineKey    string
	redmineUserId string
}

type Issue struct {
	Id int
	Project map[string]interface{}
	Subject	string
	UpdatedOn string `json:"updated_on"`
	AssignedTo map[string]*interface{} `json:"assigned_to"`
}

type Issues struct {
	Issues	[]Issue
}

func main() {
	options := parseOptions()

	issues, err := getRedmineIssues(options.redmineHost, options.redmineKey, options.redmineUserId)
	if err != nil {
		panic(err.Error())
	}
	err = sendToSlack(options, issues)
	if err != nil {
		panic(err.Error())
	}
}

func parseOptions() *Options {
	options := new(Options)
	flag.StringVar(&options.webHookUrl, "webHookUrl", "blank", "webHookUrl")
	flag.StringVar(&options.channel, "channel", "blank", "channel")
	flag.StringVar(&options.botName, "botName", "blank", "botName")
	flag.StringVar(&options.botIcon, "botIcon", ":ghost:", "botName")
	flag.StringVar(&options.redmineHost, "redmineHost", "blank", "redmineHost")
	flag.StringVar(&options.redmineKey, "redmineKey", "blank", "redmineKey")
	flag.StringVar(&options.redmineUserId, "redmineUserId", "blank", "redmineUserId")
	flag.Parse()

	return options
}

func getRedmineIssues(host, key, userId string) ([]Issue, error) {
	body, err := requestToRedmine(host, key, userId)
	if err != nil {
		return nil, err
	}
	var issues Issues
	err2 := json.Unmarshal(body, &issues)
	if err2 != nil {
		return nil, err2;
	}
	return issues.Issues, nil
}

func requestToRedmine(host, key, userId string) ([]byte, error) {
	apiUrl := host + "/issues.json?limit=100&status_id=open&key=" + key + "&assigned_to_id=" + userId
	req, err := http.NewRequest(
		"GET",
		apiUrl,
		nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: time.Duration(15 * time.Second) }
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(b))
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func sendToSlack(options *Options, issues []Issue) error {
	bot := slack.NewBot(options.webHookUrl, options.channel, options.botName, options.botIcon)

	var title string
	var msg string
	if len(issues) == 0 {
		title = "解決中のチケットはありません"
		msg = ""
	} else {
		title = (*issues[0].AssignedTo["name"]).(string) + " の"+ strconv.Itoa(len(issues)) + "件の解決中のチケット"
		for _, issue := range issues {
			msg += issue.UpdatedOn + " "
			msg += issue.Subject + " "
			msg += options.redmineHost + "/issues/" + strconv.Itoa(issue.Id)
			msg += "\n"
		}
	}

	return bot.Message(title, msg)
}