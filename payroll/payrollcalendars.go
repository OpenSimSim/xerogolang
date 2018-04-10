package payroll

import (
	"encoding/json"
	"encoding/xml"
	"log"
	"time"

	"github.com/markbates/goth"
	"github.com/opensimsim/xerogolang"
	"github.com/opensimsim/xerogolang/helpers"
)

//PayrollCalendar - placeholder waiting for .
type PayrollCalendar struct {

	// The Xero identifier for an payrollCalendar e.g. 297c2dc5-cc47-4afd-8ec8-74990b8761e9
	PayrollCalendarID string `json:"PayrollCalendarID,omitempty" xml:"Name,omitempty"`

	Name string `json:"Name,omitempty" xml:"Name,omitempty"`

	CalendarType string `json:"CalendarType,omitempty" xml:"CalendarType,omitempty"`

	PaymentDate string `json:"PaymentDate,omitempty" xml:"PaymentDate,omitempty"`
	StartDate   string `json:"StartDate,omitempty" xml:"StartDate,omitempty"`
}

/*
<PayrollCalendars>
  <PayrollCalendar>
    <CalendarType>FORTNIGHTLY</CalendarType>
    <Name>Fortnightly Calendar</Name>
    <PaymentDate>2012-08-17T00:00:00Z</PaymentDate>
    <PayrollCalendarID>a17394fe-fa23-4d4a-8e2f-a19217bc6b4f</PayrollCalendarID>
    <StartDate>2012-08-01T00:00:00</StartDate>
  </PayrollCalendar>
  <PayrollCalendar>
    <CalendarType>WEEKLY</CalendarType>
    <Name>Weekly Calendar</Name>
    <PaymentDate>2012-05-20T00:00:00Z</PaymentDate>
    <PayrollCalendarID>bfac31bd-ea62-4fc8-a5e7-7965d9504b15</PayrollCalendarID>
    <StartDate>2012-05-13T00:00:00</StartDate>
  </PayrollCalendar>
  <PayrollCalendar>
    <CalendarType>WEEKLY</CalendarType>
    <Name>What</Name>
    <PaymentDate>2012-11-16T00:00:00Z</PaymentDate>
    <PayrollCalendarID>49713875-ad73-492c-b6ac-2d265a5fe862</PayrollCalendarID>
    <StartDate>2012-11-08T00:00:00</StartDate>
  </PayrollCalendar>
</PayrollCalendars>
*/

//PayrollCalendars contains a collection of PayrollCalendars
type PayrollCalendars struct {
	PayrollCalendars []PayrollCalendar `json:"PayrollCalendars" xml:"PayrollCalendar"`
}

//The Xero API returns Dates based on the .Net JSON date format available at the time of development
//We need to convert these to a more usable format - RFC3339 for consistency with what the API expects to recieve
func (c *PayrollCalendars) convertDates() error {
	var err error
	for n := len(c.PayrollCalendars) - 1; n >= 0; n-- {
		c.PayrollCalendars[n].PaymentDate, err = helpers.DotNetJSONTimeToRFC3339(c.PayrollCalendars[n].PaymentDate, true)
		if err != nil {
			return err
		}
		c.PayrollCalendars[n].StartDate, err = helpers.DotNetJSONTimeToRFC3339(c.PayrollCalendars[n].StartDate, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func unmarshalPayrollCalendar(payrollCalendarResponseBytes []byte) (*PayrollCalendars, error) {
	var payrollCalendarResponse *PayrollCalendars

	err := json.Unmarshal(payrollCalendarResponseBytes, &payrollCalendarResponse)
	if err != nil {
		return nil, err
	}

	err = payrollCalendarResponse.convertDates()
	if err != nil {
		return nil, err
	}

	return payrollCalendarResponse, err
}

//Create will create PayrollCalendars given an PayrollCalendars struct
func (c *PayrollCalendars) Create(provider *xerogolang.Provider, session goth.Session) (*PayrollCalendars, error) {
	additionalHeaders := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/xml",
	}

	body, err := xml.MarshalIndent(c, "  ", "   ")
	if err != nil {
		return nil, err
	}

	payrollCalendarResponseBytes, err := provider.Create(session, "PayrollCalendars", additionalHeaders, body)
	if err != nil {
		return nil, err
	}

	return unmarshalPayrollCalendar(payrollCalendarResponseBytes)
}

//Update will update a PayrollCalendar given a PayrollCalendars struct
//This will only handle single PayrollCalendar - you cannot update multiple PayrollCalendars in a single call
func (c *PayrollCalendars) Update(provider *xerogolang.Provider, session goth.Session) (*PayrollCalendars, error) {
	additionalHeaders := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/xml",
	}

	body, err := xml.MarshalIndent(c, "  ", "   ")
	if err != nil {
		return nil, err
	}

	payrollCalendarResponseBytes, err := provider.Update(session, "https://api.xero.com/payroll.xro/1.0/PayrollCalendars/"+c.PayrollCalendars[0].PayrollCalendarID, additionalHeaders, body)
	if err != nil {
		return nil, err
	}

	return unmarshalPayrollCalendar(payrollCalendarResponseBytes)
}

//FindPayrollCalendarsModifiedSince
func FindPayrollCalendarsModifiedSince(provider *xerogolang.Provider, session goth.Session, modifiedSince time.Time, querystringParameters map[string]string) (*PayrollCalendars, error) {
	additionalHeaders := map[string]string{
		"Accept": "application/json",
	}

	if !modifiedSince.Equal(dayZero) {
		additionalHeaders["If-Modified-Since"] = modifiedSince.Format(time.RFC3339)
	}

	payrollCalendarResponseBytes, err := provider.Find(session, "https://api.xero.com/payroll.xro/1.0/PayrollCalendars", additionalHeaders, querystringParameters)
	if err != nil {
		return nil, err
	}

	return unmarshalPayrollCalendar(payrollCalendarResponseBytes)
}

//FindPayrollCalendars will get all PayrollCalendars.
func FindPayrollCalendars(provider *xerogolang.Provider, session goth.Session, querystringParameters map[string]string) (*PayrollCalendars, error) {
	return FindPayrollCalendarsModifiedSince(provider, session, dayZero, querystringParameters)
}

//FindPayrollCalendar will get a single PayrollCalendar
func FindPayrollCalendar(provider *xerogolang.Provider, session goth.Session, payrollCalendarID string) (*PayrollCalendars, error) {
	additionalHeaders := map[string]string{
		"Accept": "application/json",
	}

	log.Printf("Calling FindPayrollCalendar: %s\n", payrollCalendarID)

	payrollCalendarResponseBytes, err := provider.Find(session, "https://api.xero.com/payroll.xro/1.0/PayrollCalendars/"+payrollCalendarID, additionalHeaders, nil)
	if err != nil {
		return nil, err
	}

	return unmarshalPayrollCalendar(payrollCalendarResponseBytes)
}
