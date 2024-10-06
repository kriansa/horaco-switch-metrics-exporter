package scraper

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/go-resty/resty/v2"
)

var ErrorInvalidAuth = errors.New("invalid auth token")

type SwitchScraper struct {
	BaseURL      string
	User         string
	Pass         string
	sessionToken string
	portStats    map[string]*PortStats
}

func NewScraper(baseUrl, user, pass string) *SwitchScraper {
	// The session token is a md5 hash of the user and pass
	md5Hash := md5.Sum([]byte(user + pass))
	sessionToken := hex.EncodeToString(md5Hash[:])

	return &SwitchScraper{
		BaseURL:      baseUrl,
		User:         user,
		Pass:         pass,
		sessionToken: sessionToken,
		portStats:    make(map[string]*PortStats, 0),
	}
}

func (s *SwitchScraper) FetchData() (map[string]*PortStats, error) {
	// In case the existing session token is already authenticated, we can skip the sign in to avoid
	// one extra HTTP call. In case it really is invalid, then we attempt to sign in and retry this
	// call.
	err := s.fetchPortStatistics()

	if err == ErrorInvalidAuth {
		if s.signIn() != nil {
			return nil, err
		}

		if s.fetchPortStatistics() != nil {
			return nil, err
		}
	}

	if s.fetchPortSettings() != nil {
		return nil, err
	}

	return s.portStats, nil
}

func (s SwitchScraper) signIn() error {
	payload := map[string]string{
		"username": s.User,
		"password": s.Pass,
		"language": "EN",
		"Response": s.sessionToken,
	}

	client := resty.New()
	client.SetTimeout(1 * time.Second)
	_, err := client.R().
		SetFormData(payload).
		Post(s.BaseURL + "/login.cgi")

	if err != nil {
		return fmt.Errorf("unable to send the login request: %w", err)
	}

	return nil
}

func (s SwitchScraper) get(path string, queryParams map[string]string) (string, error) {
	client := resty.New()
	client.SetTimeout(1 * time.Second)

	cookie := &http.Cookie{
		Name:  "admin",
		Value: s.sessionToken,
	}

	resp, err := client.R().
		SetCookie(cookie).
		SetQueryParams(queryParams).
		Get(s.BaseURL + path)

	if err != nil {
		return "", fmt.Errorf("unable to send the request to %s: %w", path, err)
	}

	if strings.Contains(resp.String(), "window.top.location.replace(\"/login.cgi\");") {
		return "", ErrorInvalidAuth
	}

	return resp.String(), nil
}

// Get the statistics from the `Monitoring > Port Statistics` page
func (s SwitchScraper) fetchPortStatistics() error {
	result, err := s.get("/port.cgi", map[string]string{"page": "stats"})
	if err != nil {
		return fmt.Errorf("unable to fetch port statistics: %w", err)
	}

	doc, _ := htmlquery.Parse(strings.NewReader(result))
	row, _ := htmlquery.QueryAll(doc, "//table/tbody/tr")

	for _, item := range row {
		cells, _ := htmlquery.QueryAll(item, "//td")
		if len(cells) < 7 {
			continue
		}

		portName := htmlquery.InnerText(cells[0])
		if s.portStats[portName] == nil {
			s.portStats[portName] = new(PortStats)
		}

		s.portStats[portName].Enabled = htmlquery.InnerText(cells[1]) == "Enable"
		s.portStats[portName].Connected = htmlquery.InnerText(cells[2]) == "Link Up"
		s.portStats[portName].TxGood, _ = strconv.Atoi(htmlquery.InnerText(cells[3]))
		s.portStats[portName].TxBad, _ = strconv.Atoi(htmlquery.InnerText(cells[4]))
		s.portStats[portName].RxGood, _ = strconv.Atoi(htmlquery.InnerText(cells[5]))
		s.portStats[portName].RxBad, _ = strconv.Atoi(htmlquery.InnerText(cells[6]))
	}

	return nil
}

// Get the statistics from the `System > Port Setting` page
func (s SwitchScraper) fetchPortSettings() error {
	result, err := s.get("/port.cgi", nil)
	if err != nil {
		return fmt.Errorf("unable to fetch port statistics: %w", err)
	}

	doc, _ := htmlquery.Parse(strings.NewReader(result))
	rows, _ := htmlquery.QueryAll(doc, "/html/body/center/fieldset/table/tbody/tr")

	for _, row := range rows {
		cells, _ := htmlquery.QueryAll(row, "//td[position()=1 or position()=4]")
		if len(cells) < 2 {
			continue
		}

		portName := htmlquery.InnerText(cells[0])
		if s.portStats[portName] == nil {
			s.portStats[portName] = new(PortStats)
		}

		speedText := htmlquery.InnerText(cells[1])
		if strings.HasPrefix(speedText, "10000") {
			s.portStats[portName].Speed = LinkSpeed10Gbps
		} else if strings.HasPrefix(speedText, "2500") {
			s.portStats[portName].Speed = LinkSpeed2_5Gbps
		} else if strings.HasPrefix(speedText, "1000") {
			s.portStats[portName].Speed = LinkSpeed1Gbps
		} else if strings.HasPrefix(speedText, "100") {
			s.portStats[portName].Speed = LinkSpeed100Mbps
		} else if strings.HasPrefix(speedText, "10") {
			s.portStats[portName].Speed = LinkSpeed10Mbps
		} else {
			s.portStats[portName].Speed = ""
		}

		if strings.HasSuffix(speedText, "Full") {
			s.portStats[portName].TransmissionMode = LinkTransmissionModeFullDuplex
		} else if strings.HasSuffix(speedText, "Half") {
			s.portStats[portName].TransmissionMode = LinkTransmissionModeHalfDuplex
		} else {
			s.portStats[portName].TransmissionMode = ""
		}
	}

	return nil
}
