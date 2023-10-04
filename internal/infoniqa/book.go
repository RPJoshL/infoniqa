package infoniqa

import (
	"fmt"
	"net/url"
	"strings"
)

func (inf Infoniqa) Kommen() error {
	if inf.lastBookingStatus == 1 {
		return fmt.Errorf("last booking was already 'kommen'")
	}

	return inf.book(1)
}

func (inf *Infoniqa) Gehen() error {
	if inf.lastBookingStatus == 2 {
		return fmt.Errorf("last booking was already 'gehen'")
	}

	return inf.book(2)
}

// book books the given action that is identified behind the hotkey
func (inf *Infoniqa) book(hotkey int) error {

	// Build body with x-www-form-urlencoded content type (First without password and second with callback)
	data := url.Values{}
	data.Set("__WPPS", `u`)
	data.Set("__EVENTARGUMENT", ``)
	data.Set("__EVENTTARGET", ``)
	data.Set("__VIEWSTATEGENERATOR", inf.viewStateGenerator)
	data.Set("__VIEWSTATE", inf.viewstate)
	data.Set("HotKey_SI_KTO_NR", fmt.Sprintf("%d", hotkey))

	// Request with password
	req := inf.getRequest("POST", "/includes/checkworkflow.aspx", strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Origin", "https://hama.infoniqa.co.at")

	// Execute request
	res, err := inf.client.Do(req)
	if err != nil {
		return err
	}

	// Check status code
	if res.StatusCode != 200 {
		return fmt.Errorf("invalid status code (%d)", res.StatusCode)
	}

	inf.lastBookingStatus = hotkey

	return nil
}
