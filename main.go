package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/google/uuid"
)

type LoggerType int

const (
	INFO LoggerType = iota
	WARNING
	CRITICAL
	BLOCKER
)

var currTime int64 = 1629395509

// stores LogType to List of Timestamps
var logCache map[LoggerType][]int64

type LogStream struct {
	Observable
	logs []LogData
}

func NewLogStream(logs []LogData) *LogStream {
	return &LogStream{Observable{new(list.List)}, logs}
}

func (l *LogStream) Process(cstore ConfigStore) {
	for _, log := range l.logs {
		time.Sleep(time.Second / 2)
		log.Save()
		config := cstore.Fetch(log.Type)

		fmt.Println("currTime ", currTime, " config duration ", int64(config.Duration))
		previousLogs := log.FetchPreviousLogsByTimetsamp(currTime - int64(config.Duration))
		fmt.Println("Found ", len(previousLogs), " timestamps")
		fmt.Println(previousLogs)
		if len(previousLogs) >= config.Frequency {
			l.Fire(LimitReached{log.Type, config.WaitTime}) // Pass on more data if required
		}
	}
}

// Unmarshals Incoming Logstream data
type LogData struct {
	Type     LoggerType `json:"type"`
	Timstamp int64      `json:"time"`
	Content  string
}

// Fetch all the log streams from past 'time' till now
func (l *LogData) FetchPreviousLogsByTimetsamp(time int64) []int64 {
	fmt.Println("Fetching LogType ", l.Type, " for last ", time)
	allLogs := logCache[l.Type]
	res := make([]int64, 0)
	for _, v := range allLogs {
		if v >= time {
			res = append(res, v)
		}
	}
	return res
}

// saves log in a key-value store with partitionKey as LogType ans sortKey as Timestamp
func (l *LogData) Save() {
	if logCache == nil {
		logCache = make(map[LoggerType][]int64)
	}
	fmt.Println("Storing ", l.Timstamp, " to LogType", l.Type)
	logCache[l.Type] = append(logCache[l.Type], l.Timstamp)
	// ToDo: Sorting the timestamps might be required here assuming incoming logstreams are not sorted based on timestamp
}

type TopicStore struct {
	o      Observable
	topics []*Topic
}

// Similar to SNS Topic, to which Users can subscribe to
type Topic struct {
	LogType LoggerType
	Users   []*User

	LastTimeNotified int64
}

func NewTopic(log LoggerType, users []*User) *Topic {
	return &Topic{log, users, 0}
}

func NewStubTopicStore(o Observable) *TopicStore {

	infoUsers := []*User{NewUser("ram"), NewUser("sam"), NewUser("dam")}
	criticalUsers := []*User{NewUser("dibya"), NewUser("jyoti"), NewUser("hazra")}
	warningUsers := []*User{NewUser("dibya"), NewUser("jyoti"), NewUser("hazra")}

	infoTopic := NewTopic(INFO, infoUsers)
	criticalTopic := NewTopic(CRITICAL, criticalUsers)
	warningTopic := NewTopic(WARNING, warningUsers)

	return &TopicStore{o: o, topics: []*Topic{infoTopic, criticalTopic, warningTopic}}
}

func (t *TopicStore) Notify(data interface{}) {
	// notify users if lasttimenotified - currentTime > waittime
	if lr, ok := data.(LimitReached); ok {
		for _, topic := range t.topics {
			if topic.LogType == lr.LogType {
				fmt.Println("lastTimeNotified", topic.LastTimeNotified)
				if (int)(currTime-topic.LastTimeNotified) > lr.Value.(int) {
					topic.LastTimeNotified = currTime
					for _, u := range topic.Users {
						u.Notify()
					}
					fmt.Printf("Subscribers have been notified that LogType %v has exceeded its limit \n", topic.LogType)
				} else {
					fmt.Println("Wait time ", lr.Value.(int), " is not yet over for LogType", topic.LogType)
				}
			}
		}
	}
}

type Subscriber interface {
	FetchSubscribers() []User
	Subscribe(User)
}

func (t *Topic) FetchSubscribers() []User {
	panic("not implemented") // TODO: Implement
}

func (t *Topic) Subscribe(_ User) {
	panic("not implemented") // TODO: Implement
}

// User has a mail ID
type User struct {
	ID   string
	Data string
}

func NewUser(name string) *User {
	return &User{uuid.New().String(), name}
}

func (u *User) Notify() {
	// Sends a mail to the User
	fmt.Println("User ", u.Data, " has been notified!")
}

type Config struct {
	LogType   LoggerType
	Frequency int
	Duration  int
	WaitTime  int
}

func NewConfig(log LoggerType, frequency, duration, waittime int) *Config {
	return &Config{log, frequency, duration, waittime}
}

func (c *Config) String() string {
	sb := strings.Builder{}
	sb.WriteString("[ " + fmt.Sprint(c.LogType) +
		", " + fmt.Sprint(c.Frequency) +
		", " + fmt.Sprint(c.Duration) +
		", " + fmt.Sprint(c.WaitTime) + " ]")

	return sb.String()
}

func main() {
	var logs []LogData
	content, err := ioutil.ReadFile("log_stream.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal([]byte(string(content)), &logs)
	if err != nil {
		panic(err)
	}

	configStore := NewStubConfigStore()

	logstream := NewLogStream(logs)
	topicStore := NewStubTopicStore(logstream.Observable)
	logstream.Subscribe(topicStore)

	logstream.Process(configStore)
}
