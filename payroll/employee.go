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

//Employee - placeholder waiting for .
type Employee struct {

	// The Xero identifier for an employee e.g. 297c2dc5-cc47-4afd-8ec8-74990b8761e9
	EmployeeID string `json:"EmployeeID,omitempty"`

	// Current status of an employee – see employee status types
	Status string `json:"Status,omitempty"`

	// First name of an employee (max length = 255)
	FirstName string `json:"FirstName,omitempty"`

	// Last name of an employee (max length = 255)
	LastName string `json:"LastName,omitempty"`

	//
	Email string `json:"Email,omitempty"`

	Gender string `json:"Gender,omitempty"`

	Phone string `json:"Gender,omitempty"`

	UpdatedDateUTC string `json:"UpdatedDateUTC,omitempty" xml:"-"`
}

/*
<Employee>
    <EmployeeID>fb4ebd68-6568-41eb-96ab-628a0f54b4b8</EmployeeID>
    <FirstName>James</FirstName>
    <LastName>Lebron</LastName>
    <Status>ACTIVE</Status>
    <Email>JL@madeup.email.com</Email>
    <DateOfBirth>1978-08-13T00:00:00</DateOfBirth>
    <Gender>M</Gender>
    <Phone>0400-000-123</Phone>
    <Mobile> 408-230-9732</Mobile>
    <StartDate>2012-01-30T00:00:00</StartDate>
    <OrdinaryEarningsRateID>72e962d1-fcac-4083-8a71-742bb3e7ae14</OrdinaryEarningsRateID>
    <PayrollCalendarID>cb8e4706-2fdc-4170-aebd-0ffb855557f5</PayrollCalendarID>
    <UpdatedDateUTC>2013-04-01T23:02:36</UpdatedDateUTC>
  </Employee>
*/
//Employees contains a collection of Employees
type Employees struct {
	Employees []Employee `json:"Employees" xml:"Employee"`
}

var (
	dayZero = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
)

//The Xero API returns Dates based on the .Net JSON date format available at the time of development
//We need to convert these to a more usable format - RFC3339 for consistency with what the API expects to recieve
func (c *Employees) convertDates() error {
	var err error
	for n := len(c.Employees) - 1; n >= 0; n-- {
		c.Employees[n].UpdatedDateUTC, err = helpers.DotNetJSONTimeToRFC3339(c.Employees[n].UpdatedDateUTC, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func unmarshalEmployee(employeeResponseBytes []byte) (*Employees, error) {
	var employeeResponse *Employees

	log.Printf("Employee: %s\n", string(employeeResponseBytes))

	err := json.Unmarshal(employeeResponseBytes, &employeeResponse)
	if err != nil {
		return nil, err
	}

	err = employeeResponse.convertDates()
	if err != nil {
		return nil, err
	}

	return employeeResponse, err
}

//Create will create Employees given an Employees struct
func (c *Employees) Create(provider *xerogolang.Provider, session goth.Session) (*Employees, error) {
	additionalHeaders := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/xml",
	}

	body, err := xml.MarshalIndent(c, "  ", "   ")
	if err != nil {
		return nil, err
	}

	employeeResponseBytes, err := provider.Create(session, "Employees", additionalHeaders, body)
	if err != nil {
		return nil, err
	}

	return unmarshalEmployee(employeeResponseBytes)
}

//Update will update a Employee given a Employees struct
//This will only handle single Employee - you cannot update multiple Employees in a single call
func (c *Employees) Update(provider *xerogolang.Provider, session goth.Session) (*Employees, error) {
	additionalHeaders := map[string]string{
		"Accept":       "application/json",
		"Content-Type": "application/xml",
	}

	body, err := xml.MarshalIndent(c, "  ", "   ")
	if err != nil {
		return nil, err
	}

	employeeResponseBytes, err := provider.Update(session, "https://api.xero.com/payroll.xro/1.0/Employees/"+c.Employees[0].EmployeeID, additionalHeaders, body)
	if err != nil {
		return nil, err
	}

	return unmarshalEmployee(employeeResponseBytes)
}

//FindEmployeesModifiedSince
func FindEmployeesModifiedSince(provider *xerogolang.Provider, session goth.Session, modifiedSince time.Time, querystringParameters map[string]string) (*Employees, error) {
	additionalHeaders := map[string]string{
		"Accept": "application/json",
	}

	if !modifiedSince.Equal(dayZero) {
		additionalHeaders["If-Modified-Since"] = modifiedSince.Format(time.RFC3339)
	}

	employeeResponseBytes, err := provider.Find(session, "https://api.xero.com/payroll.xro/1.0/Employees", additionalHeaders, querystringParameters)
	if err != nil {
		return nil, err
	}

	return unmarshalEmployee(employeeResponseBytes)
}

//FindEmployees will get all Employees.
func FindEmployees(provider *xerogolang.Provider, session goth.Session, querystringParameters map[string]string) (*Employees, error) {
	return FindEmployeesModifiedSince(provider, session, dayZero, querystringParameters)
}

//FindEmployee will get a single Employee
func FindEmployee(provider *xerogolang.Provider, session goth.Session, employeeID string) (*Employees, error) {
	additionalHeaders := map[string]string{
		"Accept": "application/json",
	}

	log.Printf("Calling FindEmployee: %s\n", employeeID)

	employeeResponseBytes, err := provider.Find(session, "https://api.xero.com/payroll.xro/1.0/Employees/"+employeeID, additionalHeaders, nil)
	if err != nil {
		return nil, err
	}

	return unmarshalEmployee(employeeResponseBytes)
}
