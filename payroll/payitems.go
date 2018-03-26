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

//PayItem - placeholder waiting for .
type EarningsRate struct {

	// The Xero identifier for an EarningsRateID e.g. 297c2dc5-cc47-4afd-8ec8-74990b8761e9
	EarningsRateID string `json:"EarningsRateID,omitempty" xml:"EarningsRateID,omitempty"`

	// Name
	Name string `json:"Name,omitempty"  xml:"Name,omitempty"`

	// TODO add other variables - Dont need the others right now.

	UpdatedDateUTC string `json:"UpdatedDateUTC,omitempty"  xml:"UpdatedDateUTC,omitempty"`
}

/*
<PayItems>
  <EarningsRates>
    <EarningsRate>
      <EarningsRateID>eca71b79-edab-4c3f-967f-a405453bac08</EarningsRateID>
      <Name>Ordinary Hours</Name>
      <EarningsType>ORDINARYTIMEEARNINGS</EarningsType>
      <RateType>RATEPERUNIT</RateType>
      <AccountCode>477</AccountCode>
      <TypeOfUnits>Hours</TypeOfUnits>
      <IsExemptFromTax>false</IsExemptFromTax>
      <IsExemptFromSuper>false</IsExemptFromSuper>
      <IsReportableAsW1>false</IsReportableAsW1>
      <UpdatedDateUTC>2013-04-09T23:45:25</UpdatedDateUTC>
    </EarningsRate>
    ...

</PayItems>
*/

type EarningsRates struct {
	EarningsRates []EarningsRate `json:"EarningsRates" xml:"EarningsRates"`
}

//PayItems contains a collection of PayItems
type PayItems struct {
	PayItems []EarningsRates `json:"PayItems" xml:"PayItem"`
}

//The Xero API returns Dates based on the .Net JSON date format available at the time of development
//We need to convert these to a more usable format - RFC3339 for consistency with what the API expects to recieve
func (c *PayItems) convertDates() error {

	// TODO Use reflection to do this.
	var err error
	for n := len(c.PayItems) - 1; n >= 0; n-- {
		for m := len(c.PayItems[n].EarningsRates) - 1; m >= 0; m-- {
			c.PayItems[n].EarningsRates[m].UpdatedDateUTC, err = helpers.DotNetJSONTimeToRFC3339(c.PayItems[n].EarningsRates[m].UpdatedDateUTC, true)
			if err != nil {
				return err
			}

		}
	}

	return nil
}

func unmarshalPayItem(payItemResponseBytes []byte) (*PayItems, error) {
	var payItemResponse *PayItems

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
func (c *PayItems) Create(provider *xerogolang.Provider, session goth.Session) (*PayItems, error) {
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
func (c *PayItems) Update(provider *xerogolang.Provider, session goth.Session) (*PayItems, error) {
	additionalHeaders := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/xml",
	}

	body, err := xml.MarshalIndent(c, "  ", "   ")
	if err != nil {
		return nil, err
	}

	payItemResponseBytes, err := provider.Update(session, "PayItems/"+c.PayItems[0].EarningsRates[0].EarningsRateID, additionalHeaders, body)
	if err != nil {
		return nil, err
	}

	return unmarshalPayItem(payItemResponseBytes)
}

//FindPayItemsModifiedSince
func FindPayItemsModifiedSince(provider *xerogolang.Provider, session goth.Session, modifiedSince time.Time, querystringParameters map[string]string) (*PayItems, error) {
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
func FindPayItems(provider *xerogolang.Provider, session goth.Session, querystringParameters map[string]string) (*PayItems, error) {
	return FindPayItemsModifiedSince(provider, session, dayZero, querystringParameters)
}

//FindPayItem will get a single PayItem
func FindPayItem(provider *xerogolang.Provider, session goth.Session, payItemID string) (*PayItems, error) {
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
