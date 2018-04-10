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

type EarningsRate struct {
	// The Xero identifier for an EarningsRateID e.g. 297c2dc5-cc47-4afd-8ec8-74990b8761e9
	EarningsRateID string `json:"EarningsRateID,omitempty" xml:"EarningsRateID,omitempty"`

	// Name
	Name         string `json:"Name,omitempty"  xml:"Name,omitempty"`
	EarningsType string `json:"EarningsType,omitempty"  xml:"EarningsType,omitempty"`
	RateType     string `json:"RateType,omitempty"  xml:"RateType,omitempty"`
	//RatePerUnit
	// Multiplier
	AccountCode       string `json:"AccountCode,omitempty"  xml:"AccountCode,omitempty"`
	TypeOfUnits       string `json:"TypeOfUnits,omitempty"  xml:"TypeOfUnits,omitempty"`
	IsExemptFromTax   bool   `json:"IsExemptFromTax,omitempty"  xml:"IsExemptFromTax,omitempty"`
	IsExemptFromSuper bool   `json:"IsExemptFromSuper,omitempty"  xml:"IsExemptFromSuper,omitempty"`
	IsReportableAsW1  bool   `json:"IsReportableAsW1,omitempty"  xml:"IsReportableAsW1,omitempty"`
	UpdatedDateUTC    string `json:"UpdatedDateUTC,omitempty"  xml:"UpdatedDateUTC,omitempty"`
}

type DeductionType struct {
	// The Xero identifier for an EarningsRateID e.g. 297c2dc5-cc47-4afd-8ec8-74990b8761e9
	DeductionTypeID string `json:"DeductionTypeID,omitempty" xml:"DeductionTypeID,omitempty"`
	Name            string `json:"Name,omitempty"  xml:"Name,omitempty"`
	AccountCode     string `json:"AccountCode,omitempty"  xml:"AccountCode,omitempty"`
	ReducesTax      bool   `json:"IsExemptFromTax,omitempty"  xml:"IsExemptFromTax,omitempty"`
	ReducesSuper    bool   `json:"IsExemptFromSuper,omitempty"  xml:"IsExemptFromSuper,omitempty"`
	IsExemptFromW1  bool   `json:"IsReportableAsW1,omitempty"  xml:"IsReportableAsW1,omitempty"`
	UpdatedDateUTC  string `json:"UpdatedDateUTC,omitempty"  xml:"UpdatedDateUTC,omitempty"`
}

type ReimbursementType struct {
	ReimbursementTypeID string `json:"ReimbursementTypeID,omitempty" xml:"ReimbursementTypeID,omitempty"`
	Name                string `json:"Name,omitempty"  xml:"Name,omitempty"`
	AccountCode         string `json:"AccountCode,omitempty"  xml:"AccountCode,omitempty"`
	UpdatedDateUTC      string `json:"UpdatedDateUTC,omitempty"  xml:"UpdatedDateUTC,omitempty"`
}

type LeaveType struct {
	LeaveTypeID       string  `json:"LeaveTypeID,omitempty" xml:"LeaveTypeID,omitempty"`
	Name              string  `json:"Name,omitempty"  xml:"Name,omitempty"`
	TypeOfUnits       string  `json:"TypeOfUnits,omitempty"  xml:"TypeOfUnits,omitempty"`
	NormalEntitlement float64 `json:"NormalEntitlement,omitempty"  xml:"NormalEntitlement,omitempty"`
	IsPaidLeave       bool    `json:"IsPaidLeave,omitempty"  xml:"IsPaidLeave,omitempty"`
	ShowOnPayslip     bool    `json:"ShowOnPayslip,omitempty"  xml:"ShowOnPayslip,omitempty"`
	UpdatedDateUTC    string  `json:"UpdatedDateUTC,omitempty"  xml:"UpdatedDateUTC,omitempty"`
}

type PayItems struct {
	EarningsRates     []EarningsRate      `json:"EarningsRates" xml:"EarningsRates"`
	DeductionTypes    []DeductionType     `json:"DeductionTypes" xml:"DeductionTypes"`
	ReimbursementType []ReimbursementType `json:"ReimbursementType" xml:"ReimbursementType"`
	LeaveTypes        []LeaveType         `json:"LeaveTypes" xml:"LeaveTypes"`
}

//PayItems contains a collection of PayItems
type PayItem struct {
	ID           string   `json:"Id,omitempty" xml:"Id,omitempty"`
	Status       string   `json:"Status,omitempty" xml:"Status,omitempty"`
	ProviderName string   `json:"ProviderName,omitempty" xml:"ProviderName,omitempty"`
	DateTimeUTC  string   `json:"DateTimeUTC,omitempty" xml:"DateTimeUTC,omitempty"`
	PayItems     PayItems `json:"PayItems" xml:"PayItem"`
}

//The Xero API returns Dates based on the .Net JSON date format available at the time of development
//We need to convert these to a more usable format - RFC3339 for consistency with what the API expects to recieve
func (c *PayItem) convertDates() error {
	var err error
	// TODO Use reflection to do this.

	c.DateTimeUTC, err = helpers.DotNetJSONTimeToRFC3339(c.DateTimeUTC, false)

	for n := len(c.PayItems.EarningsRates) - 1; n >= 0; n-- {
		c.PayItems.EarningsRates[n].UpdatedDateUTC, err = helpers.DotNetJSONTimeToRFC3339(c.PayItems.EarningsRates[n].UpdatedDateUTC, true)
		if err != nil {
			return err
		}

	}

	return nil
}

func unmarshalPayItem(payItemResponseBytes []byte) (*PayItem, error) {
	var payItemResponse *PayItem

	log.Printf("PayItem: %s\n", string(payItemResponseBytes))

	err := json.Unmarshal(payItemResponseBytes, &payItemResponse)
	if err != nil {
		return nil, err
	}

	err = payItemResponse.convertDates()
	if err != nil {
		return nil, err
	}

	return payItemResponse, err
}

//Create will create PayItems given an PayItems struct
func (c *PayItem) Create(provider *xerogolang.Provider, session goth.Session) (*PayItem, error) {
	additionalHeaders := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/xml",
	}

	body, err := xml.MarshalIndent(c, "  ", "   ")
	if err != nil {
		return nil, err
	}

	payItemResponseBytes, err := provider.Create(session, "PayItems", additionalHeaders, body)
	if err != nil {
		return nil, err
	}

	return unmarshalPayItem(payItemResponseBytes)
}

//Update will update a PayItem given a PayItems struct
//This will only handle single PayItem - you cannot update multiple PayItems in a single call
func (c *PayItem) Update(provider *xerogolang.Provider, session goth.Session) (*PayItem, error) {
	additionalHeaders := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/xml",
	}

	body, err := xml.MarshalIndent(c, "  ", "   ")
	if err != nil {
		return nil, err
	}

	payItemResponseBytes, err := provider.Update(session, "PayItems/"+c.PayItems.EarningsRates[0].EarningsRateID, additionalHeaders, body)
	if err != nil {
		return nil, err
	}

	return unmarshalPayItem(payItemResponseBytes)
}

//FindPayItemsModifiedSince
func FindPayItemsModifiedSince(provider *xerogolang.Provider, session goth.Session, modifiedSince time.Time, querystringParameters map[string]string) (*PayItem, error) {
	additionalHeaders := map[string]string{
		"Accept": "application/json",
	}

	if !modifiedSince.Equal(dayZero) {
		additionalHeaders["If-Modified-Since"] = modifiedSince.Format(time.RFC3339)
	}

	payItemResponseBytes, err := provider.FindWithEndpoint(session, "https://api.xero.com/payroll.xro/1.0/", "PayItems", additionalHeaders, querystringParameters)
	if err != nil {
		return nil, err
	}

	return unmarshalPayItem(payItemResponseBytes)
}

//FindPayItems will get all PayItems.
func FindPayItems(provider *xerogolang.Provider, session goth.Session, querystringParameters map[string]string) (*PayItem, error) {
	return FindPayItemsModifiedSince(provider, session, dayZero, querystringParameters)
}

//FindPayItem will get a single PayItem
func FindPayItem(provider *xerogolang.Provider, session goth.Session, payItemID string) (*PayItem, error) {
	additionalHeaders := map[string]string{
		"Accept": "application/json",
	}

	log.Printf("Calling FindPayItem: %s\n", payItemID)

	payItemResponseBytes, err := provider.FindWithEndpoint(session, "https://api.xero.com/payroll.xro/1.0/", "PayItems/"+payItemID, additionalHeaders, nil)
	if err != nil {
		return nil, err
	}

	return unmarshalPayItem(payItemResponseBytes)
}
