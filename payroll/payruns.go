package payroll

import (
	"encoding/json"
	"encoding/xml"
	"time"

	"github.com/markbates/goth"
	"github.com/opensimsim/xerogolang"
	"github.com/opensimsim/xerogolang/helpers"
)

//PayRun - placeholder waiting for .
type PayRun struct {
	PayRunID          string `json:"PayRunID,omitempty" xml:"PayRunID,omitempty"`
	PayrollCalendarID string `json:"PayrollCalendarID,omitempty" xml:"PayrollCalendarID,omitempty"`

	PayRunPeriodEndDate   string `json:"PayRunPeriodEndDate,omitempty" xml:"PayRunPeriodEndDate,omitempty"`
	PayRunPeriodStartDate string `json:"PayRunPeriodStartDate,omitempty" xml:"PayRunPeriodStartDate,omitempty"`

	PayRunStatus string `json:"PayRunStatus,omitempty" xml:"PayRunStatus,omitempty"`
	PaymentDate  string `json:"PaymentDate,omitempty" xml:"PaymentDate,omitempty"`

	Deductions float64 `json:"Deductions,omitempty" xml:"Deductions,omitempty"`
	NetPay     float64 `json:"NetPay,omitempty" xml:"NetPay,omitempty"`
	Super      float64 `json:"Super,omitempty" xml:"Super,omitempty"`
	Tax        float64 `json:"Tax,omitempty" xml:"Tax,omitempty"`
	Wages      float64 `json:"Wages,omitempty" xml:"Wages,omitempty"`
}

/*
<PayRuns>
  <PayRun>
    <Deductions>260.04</Deductions>
    <NetPay>18831.25</NetPay>
    <PayRunID>e3bdb2f7-2b20-45e6-ac8d-ec67d17de9f4</PayRunID>
    <PayRunPeriodEndDate>2012-01-07T00:00:00</PayRunPeriodEndDate>
    <PayRunPeriodStartDate>2012-01-01T00:00:00</PayRunPeriodStartDate>
    <PayRunStatus>Posted</PayRunStatus>
    <PaymentDate>2012-01-08T00:00:00</PaymentDate>
    <PayrollCalendarID>bfac31bd-ea62-4fc8-a5e7-7965d9504b15</PayrollCalendarID>
    <Super>2539.97</Super>
    <Tax>6651.00</Tax>
    <Wages>25742.29</Wages>
  </PayRun>
  <PayRun>
    <Deductions>260.04</Deductions>
    <NetPay>22463.25</NetPay>
    <PayRunID>7c998e04-1cee-4a19-bfe6-3cbfd5cb9cea</PayRunID>
    <PayRunPeriodEndDate>2012-01-14T00:00:00</PayRunPeriodEndDate>
    <PayRunPeriodStartDate>2012-01-08T00:00:00</PayRunPeriodStartDate>
    <PayRunStatus>Posted</PayRunStatus>
    <PaymentDate>2012-01-15T00:00:00</PaymentDate>
    <PayrollCalendarID>bfac31bd-ea62-4fc8-a5e7-7965d9504b15</PayrollCalendarID>
    <Super>2892.78</Super>
    <Tax>6939.00</Tax>
    <Wages>29662.29</Wages>
  </PayRun>
</PayRuns>
*/

//PayRuns contains a collection of PayRuns
type PayRuns struct {
	PayRuns []PayRun `json:"PayRuns" xml:"PayRun"`
}

//The Xero API returns Dates based on the .Net JSON date format available at the time of development
//We need to convert these to a more usable format - RFC3339 for consistency with what the API expects to recieve
func (c *PayRuns) convertDates() error {
	var err error
	for n := len(c.PayRuns) - 1; n >= 0; n-- {
		c.PayRuns[n].PayRunPeriodStartDate, err = helpers.DotNetJSONTimeToRFC3339(c.PayRuns[n].PayRunPeriodStartDate, true)
		if err != nil {
			return err
		}
		c.PayRuns[n].PayRunPeriodEndDate, err = helpers.DotNetJSONTimeToRFC3339(c.PayRuns[n].PayRunPeriodEndDate, true)
		if err != nil {
			return err
		}

		c.PayRuns[n].PaymentDate, err = helpers.DotNetJSONTimeToRFC3339(c.PayRuns[n].PaymentDate, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func unmarshalPayRun(payRunResponseBytes []byte) (*PayRuns, error) {
	var payRunResponse *PayRuns

	err := json.Unmarshal(payRunResponseBytes, &payRunResponse)
	if err != nil {
		return nil, err
	}

	err = payRunResponse.convertDates()
	if err != nil {
		return nil, err
	}

	return payRunResponse, err
}

//Create will create PayRuns given an PayRuns struct
func (c *PayRuns) Create(provider *xerogolang.Provider, session goth.Session) (*PayRuns, error) {
	additionalHeaders := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/xml",
	}

	body, err := xml.MarshalIndent(c, "  ", "   ")
	if err != nil {
		return nil, err
	}

	payRunResponseBytes, err := provider.Create(session, "PayRuns", additionalHeaders, body)
	if err != nil {
		return nil, err
	}

	return unmarshalPayRun(payRunResponseBytes)
}

//Update will update a PayRun given a PayRuns struct
//This will only handle single PayRun - you cannot update multiple PayRuns in a single call
func (c *PayRuns) Update(provider *xerogolang.Provider, session goth.Session) (*PayRuns, error) {
	additionalHeaders := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/xml",
	}

	body, err := xml.MarshalIndent(c, "  ", "   ")
	if err != nil {
		return nil, err
	}

	payRunResponseBytes, err := provider.Update(session, "https://api.xero.com/payroll.xro/1.0/PayRuns/"+c.PayRuns[0].PayRunID, additionalHeaders, body)
	if err != nil {
		return nil, err
	}

	return unmarshalPayRun(payRunResponseBytes)
}

//FindPayRunsModifiedSince
func FindPayRunsModifiedSince(provider *xerogolang.Provider, session goth.Session, modifiedSince time.Time, querystringParameters map[string]string) (*PayRuns, error) {
	additionalHeaders := map[string]string{
		"Accept": "application/json",
	}

	if !modifiedSince.Equal(dayZero) {
		additionalHeaders["If-Modified-Since"] = modifiedSince.Format(time.RFC3339)
	}

	payRunResponseBytes, err := provider.Find(session, "https://api.xero.com/payroll.xro/1.0/PayRuns", additionalHeaders, querystringParameters)
	if err != nil {
		return nil, err
	}

	return unmarshalPayRun(payRunResponseBytes)
}

//FindPayRuns will get all PayRuns.
func FindPayRuns(provider *xerogolang.Provider, session goth.Session, querystringParameters map[string]string) (*PayRuns, error) {
	return FindPayRunsModifiedSince(provider, session, dayZero, querystringParameters)
}

//FindPayRun will get a single PayRun
func FindPayRun(provider *xerogolang.Provider, session goth.Session, payRunID string) (*PayRuns, error) {
	additionalHeaders := map[string]string{
		"Accept": "application/json",
	}

	payRunResponseBytes, err := provider.Find(session, "https://api.xero.com/payroll.xro/1.0/PayRuns/"+payRunID, additionalHeaders, nil)
	if err != nil {
		return nil, err
	}

	return unmarshalPayRun(payRunResponseBytes)
}
