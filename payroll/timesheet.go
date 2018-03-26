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

//Timesheet - placeholder waiting for .
type Timesheet struct {

	// The Xero identifier for an timesheet e.g. 297c2dc5-cc47-4afd-8ec8-74990b8761e9
	TimesheetID string `json:"TimesheetID,omitempty" xml:"TimesheetID"`

	StartDateUTC string `json:"StartDate,omitempty" xml:"StartDate"`

	EndDateUTC string `json:"EndDate,omitempty" xml:"EndDate"`

	Status string `json:"Status,omitempty" xml:"Status,omitempty"`

	Hours float64 `json:"Hours,omitempty" xml:"Hours,omitempty"`

	TimesheetLine []TimesheetLine `json:"TimesheetLines,omitempty" xml:"TimesheetLines,omitempty"`
}

type TimesheetLine struct {
	EarningsRateID string    `json:"EarningsRateID,omitempty" xml:"EarningsRateID"`
	NumberOfUnits  []float64 `json:"NumberOfUnits,omitempty" xml:"NumberOfUnits,omitempty"`
}

type NumberOfUnit struct {
	NumberOfUnit float64 `json:"NumberOfUnit,omitempty" xml:"NumberOfUnit,omitempty"`
}

/*
<Timesheet>
    <TimesheetID>5e493b2e-c3ed-4172-95b2-593438101f76</TimesheetID>
    <StartDate>2018-03-25T00:00:00</StartDate>
    <EndDate>2018-04-01T00:00:00</EndDate>
    <Status>Draft</Status>
    <TimesheetLines>
        <TimesheetLine>
            <EarningsRateID>0daff504-2d42-4243-bdac-24f2bae0ce7c</EarningsRateID>
            <NumberOfUnits>
                <NumberOfUnit>8</NumberOfUnit>
                <NumberOfUnit>8</NumberOfUnit>
                <NumberOfUnit>8</NumberOfUnit>
                <NumberOfUnit>8</NumberOfUnit>
                <NumberOfUnit>8</NumberOfUnit>
                <NumberOfUnit>0</NumberOfUnit>
                <NumberOfUnit>0</NumberOfUnit>
            </NumberOfUnits>
        </TimesheetLine>
    </TimesheetLines>
</Timesheet>
*/
//Timesheets contains a collection of Timesheets
type Timesheets struct {
	ID           string `json:"Id,omitempty" xml:"Timesheet"`
	Status       string `json:"Status,omitempty" xml:"Status"`
	ProviderName string `json:"ProviderName,omitempty" xml:"ProviderName"`
	DateTimeUTC  string `json:"DateTimeUTC,omitempty" xml:"-"`

	Timesheets []Timesheet `json:"Timesheets" xml:"Timesheets"`
}

//The Xero API returns Dates based on the .Net JSON date format available at the time of development
//We need to convert these to a more usable format - RFC3339 for consistency with what the API expects to recieve
func (c *Timesheets) convertDates() error {

	var err error
	c.DateTimeUTC, err = helpers.DotNetJSONTimeToRFC3339(c.DateTimeUTC, true)
	if err != nil {
		return err
	}
	/*	var err error
		for n := len(c.Timesheets) - 1; n >= 0; n-- {
			c.Timesheets[n].UpdatedDateUTC, err = helpers.DotNetJSONTimeToRFC3339(c.Timesheets[n].UpdatedDateUTC, true)
			if err != nil {
				return err
			}
		}*/

	return nil
}

func unmarshalTimesheet(timesheetResponseBytes []byte) (*Timesheets, error) {
	var timesheetResponse *Timesheets

	log.Printf("Timesheet: %s\n", string(timesheetResponseBytes))

	err := json.Unmarshal(timesheetResponseBytes, &timesheetResponse)
	if err != nil {
		return nil, err
	}

	err = timesheetResponse.convertDates()
	if err != nil {
		return nil, err
	}

	return timesheetResponse, err
}

//Create will create Timesheets given an Timesheets struct
func (c *Timesheets) Create(provider *xerogolang.Provider, session goth.Session) (*Timesheets, error) {
	additionalHeaders := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/xml",
	}

	body, err := xml.MarshalIndent(c, "  ", "   ")
	if err != nil {
		return nil, err
	}

	timesheetResponseBytes, err := provider.Create(session, "Timesheets", additionalHeaders, body)
	if err != nil {
		return nil, err
	}

	return unmarshalTimesheet(timesheetResponseBytes)
}

//Update will update a Timesheet given a Timesheets struct
//This will only handle single Timesheet - you cannot update multiple Timesheets in a single call
func (c *Timesheets) Update(provider *xerogolang.Provider, session goth.Session) (*Timesheets, error) {
	additionalHeaders := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/xml",
	}

	body, err := xml.MarshalIndent(c, "  ", "   ")
	if err != nil {
		return nil, err
	}

	timesheetResponseBytes, err := provider.Update(session, "Timesheets/"+c.Timesheets[0].TimesheetID, additionalHeaders, body)
	if err != nil {
		return nil, err
	}

	return unmarshalTimesheet(timesheetResponseBytes)
}

//FindTimesheetsModifiedSince
func FindTimesheetsModifiedSince(provider *xerogolang.Provider, session goth.Session, modifiedSince time.Time, querystringParameters map[string]string) (*Timesheets, error) {
	additionalHeaders := map[string]string{
		"Accept": "application/json",
	}

	if !modifiedSince.Equal(dayZero) {
		additionalHeaders["If-Modified-Since"] = modifiedSince.Format(time.RFC3339)
	}

	timesheetResponseBytes, err := provider.FindWithEndpoint(session, "https://api.xero.com/payroll.xro/1.0/", "Timesheets", additionalHeaders, querystringParameters)
	if err != nil {
		return nil, err
	}

	return unmarshalTimesheet(timesheetResponseBytes)
}

//FindTimesheets will get all Timesheets.
func FindTimesheets(provider *xerogolang.Provider, session goth.Session, querystringParameters map[string]string) (*Timesheets, error) {
	return FindTimesheetsModifiedSince(provider, session, dayZero, querystringParameters)
}

//FindTimesheet will get a single Timesheet
func FindTimesheet(provider *xerogolang.Provider, session goth.Session, timesheetID string) (*Timesheets, error) {
	additionalHeaders := map[string]string{
		"Accept": "application/json",
	}

	log.Printf("Calling FindTimesheet: %s\n", timesheetID)

	timesheetResponseBytes, err := provider.FindWithEndpoint(session, "https://api.xero.com/payroll.xro/1.0/", "Timesheets/"+timesheetID, additionalHeaders, nil)
	if err != nil {
		return nil, err
	}

	return unmarshalTimesheet(timesheetResponseBytes)
}
