package infoniqa

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"git.rpjosh.de/RPJosh/go-logger"
)

// Infoniqa is the base struct with all available functions that are
// implemented within this program
type Infoniqa struct {
	// URL of the infoniqa instance
	BaseUrl string

	// Username to log in
	Username string
	// Password for the user
	Password string

	// The viewstate that was provided by infoniqa from the last request
	viewstate string
	// The last view state generator that was provided by infoniqa from the last request
	viewStateGenerator string
	// Client for executing the request. This does also store the cookies
	client http.Client

	// Last booking status (0 = unknown, 1 = kommen, 2 = gehen)
	lastBookingStatus int
}

// NewInfoniqa creates a new infoniqa instance with the provided credentials.
// This function will execute an initialization sequence so that the other functions
// of this struct can be used.
// When this initialization sequence does fail, it will return an error
func NewInfoniqa(baseUrl string, username string, password string) (*Infoniqa, error) {

	// Create instance with parameters
	inf := &Infoniqa{
		BaseUrl:  baseUrl,
		Username: username,
		Password: password,
		client:   http.Client{Timeout: 5 * time.Second, Jar: NewJar()},
	}

	// Get the login page to login
	_, res, err := inf.executeRequest(inf.getRequest("GET", "/Default.aspx", nil))
	if err != nil {
		return nil, fmt.Errorf("fetching of login page failed: %s", err)
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("unable to contact infoniqa site. Got status %d", res.StatusCode)
	}

	// Execute the login
	if err := inf.login(); err != nil {
		return nil, fmt.Errorf("failed to login: %s", err)
	}

	return inf, nil
}

// getRequest returns a new request to the infoniqa API
func (inf *Infoniqa) getRequest(method string, path string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, inf.BaseUrl+path, body)
	if err != nil {
		logger.Error("Failed to get infoniqa request: %s", err)
	}

	// Set common headers
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")

	return req
}

// executeRequest executes the given requests and updates aspx specific variables like the viewstate
// accordingly
func (inf *Infoniqa) executeRequest(req *http.Request) (body string, resp *http.Response, err error) {

	// Execute the request
	res, err := inf.client.Do(req)
	if err != nil {
		return "", nil, err
	}

	// Read the body
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", res, fmt.Errorf("reading of body failed: %s", err)
	}
	strBody := string(resBody)

	// Get the viewstate from the body
	if st, err := inf.findHiddenValue("__VIEWSTATE", strBody); err != nil {
		return "", res, fmt.Errorf("no viewstate found in response")
	} else {
		inf.viewstate = st
	}
	if st, err := inf.findHiddenValue("__VIEWSTATEGENERATOR", strBody); err != nil {
		return "", res, fmt.Errorf("no viewstateGenerator found in response")
	} else {
		inf.viewStateGenerator = st
	}

	return string(resBody), res, nil
}

// findHiddenValue searches the given aspx value within the
// response body as a hidden input type.
func (inf *Infoniqa) findHiddenValue(name string, body string) (value string, err error) {
	// Get the viewstate from the body
	regex, err := regexp.Compile(`<input.type="hidden".name="` + name + `".id="` + name + `" value="(?P<Viewstate>.*)".\/>`)
	if err != nil {
		return "", fmt.Errorf("failed to compile regex: %s", err)
	}
	matches := regex.FindStringSubmatch(body)
	index := regex.SubexpIndex("Viewstate")
	if index >= len(matches) {
		return "", fmt.Errorf("no viewstate found in response")
	}

	return matches[index], nil
}

// login calls the login endpoint of infoniqa and sets the cookie and viewstate
// for all further requests correctly
func (inf *Infoniqa) login() error {

	// Build body with x-www-form-urlencoded content type (First without password and second with callback)
	data := url.Values{}
	data.Set("__EVENTTARGET", `ctl00$ContentPlaceHolder1$PanelLogin$PageControl$Login1$btnApgLogin`)
	data.Set("__EVENTARGUMENT", `Click`)
	data.Set("__VIEWSTATE", inf.viewstate)
	data.Set("__VIEWSTATEGENERATOR", inf.viewStateGenerator)
	data.Set("ctl00$Logininfo1$CheckPopupControlState", `{"windowsState":"0:0:-1:0:0:0:-10000:-10000:1:0:0:0"}`)
	data.Set("ctl00$ContentPlaceHolder1$PanelLogin$PageControl", `{"activeTabIndex":0}`)
	data.Set("ctl00$ContentPlaceHolder1$PanelLogin$PageControl$Login1$UserName$State", `{"validationState":""}`)
	data.Set("ctl00$ContentPlaceHolder1$PanelLogin$PageControl$Login1$UserName", inf.Username)
	data.Set("ctl00$ContentPlaceHolder1$PanelLogin$PageControl$Login1$Password$State", `{"validationState":""}`)
	data.Set("ctl00$ContentPlaceHolder1$PanelLogin$PageControl$Login1$Password", inf.Password)
	data.Set("ctl00$ContentPlaceHolder1$PanelLogin$PageControl$PasswordRecovery$UserNameContainerID$UserName$State", `{"validationState":""}`)
	data.Set("ctl00$ContentPlaceHolder1$PanelLogin$PageControl$PasswordRecovery$UserNameContainerID$UserName", "")

	// Request with password
	req := inf.getRequest("POST", "/Default.aspx", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Origin", "https://hama.infoniqa.co.at")

	// Execute request
	body, res, err := inf.executeRequest(req)
	if err != nil {
		return err
	}

	// Check status code
	if res.StatusCode != 200 {
		return fmt.Errorf("login failed (%d)", res.StatusCode)
	}

	// Get the last booking status
	regex := regexp.MustCompile(`<td.*return overlib\('(?P<State>.*)', CAPTION.*\).*id="Zeitleiste".*<\/td>`)
	matches := regex.FindStringSubmatch(body)
	index := regex.SubexpIndex("State")
	if index >= len(matches) {
		logger.Debug("Couldn't find the last booking state")
	} else if strings.HasPrefix(matches[index], "KO") {
		inf.lastBookingStatus = 1
	} else {
		inf.lastBookingStatus = 2
		logger.Debug("Found last booking state state %q", matches[index])
	}

	return nil
}
