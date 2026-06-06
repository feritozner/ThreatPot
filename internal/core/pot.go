package core

import (
	"sync"
)

type RequestDump struct {
	ID      int
	Time    string
	Method  string
	Path    string
	IP      string
	RawData string
}

type Pot struct {
	Name         string
	Description  string
	Port         int
	BindIP       string
	ServerHeader string
	Delay        int
	IsRunning    bool
	Logs         []string
	LogChans     []chan string
	Routes       map[string]string
	ReqDumps     []RequestDump
	ReqCount     int
	Mutex        sync.Mutex
}

var Pots = make(map[string]*Pot)

func InitPots() {
	Pots["pot_1"] = &Pot{
		Name:         "pot_1",
		Description:  "Default VPN Honeypot",
		Port:         80,
		BindIP:       "0.0.0.0",
		ServerHeader: "Corporate VPN Server",
		Delay:        1000,
		IsRunning:    false,
		Logs:         []string{},
		LogChans:     []chan string{},
		Routes: map[string]string{
			"/":                "templates/index.html",
			"/login":           "templates/login.html",
			"/api/v1/auth":     "templates/api_auth_fail.json",
			"/api/v2/helpdesk": "templates/api_helpdesk_fail.json",
			"/api/v2/status":   "templates/api_status_fail.json",
		},
		ReqDumps: []RequestDump{},
		ReqCount: 0,
	}
}

func CreatePot(name string) {
	Pots[name] = &Pot{
		Name:         name,
		Description:  "Newly created honeypot",
		Port:         8080,
		BindIP:       "0.0.0.0",
		ServerHeader: "Corporate VPN Server",
		Delay:        0,
		IsRunning:    false,
		Logs:         []string{},
		LogChans:     []chan string{},
		Routes: map[string]string{
			"/":                "templates/index.html",
			"/login":           "templates/login.html",
			"/api/v1/auth":     "templates/api_auth_fail.json",
			"/api/v2/helpdesk": "templates/api_helpdesk_fail.json",
			"/api/v2/status":   "templates/api_status_fail.json",
		},
		ReqDumps: []RequestDump{},
		ReqCount: 0,
	}
}

func (p *Pot) AddLog(logLine string) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	p.Logs = append(p.Logs, logLine)
	for _, ch := range p.LogChans {
		select {
		case ch <- logLine:
		default:
		}
	}
}

func (p *Pot) Subscribe() chan string {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	ch := make(chan string, 100)
	p.LogChans = append(p.LogChans, ch)
	return ch
}

func (p *Pot) Unsubscribe(ch chan string) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	for i, c := range p.LogChans {
		if c == ch {
			p.LogChans = append(p.LogChans[:i], p.LogChans[i+1:]...)
			close(ch)
			break
		}
	}
}
