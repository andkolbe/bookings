package forms

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
)

// hold all information associated with our form either when it is rendered for the first time,
// or after it is submitted and there might be one or more errors
type Form struct {
	url.Values
	Errors errors
}

// returns true if there are no errors, otherwise return false
func (f *Form) Valid() bool {
	return len(f.Errors) == 0
}

// initializes the form struct
func New(data url.Values) *Form {
	return &Form{
		data,
		errors(map[string][]string{}),
	}
}

// checks for required fields
func (f *Form) Required(fields ...string) { // ...string means you can pass in as many string parameters as you want
	for _, field := range fields {
		value := f.Get(field)
		if strings.TrimSpace(value) == "" { // removes any extraneous spaces the user may have filled in by mistake
			f.Errors.Add(field, "This field cannot be blank")
		}
	}
}

// checks if form field is in post and not empty
func (f *Form) Has(field string) bool {
	x := f.Get(field)
	// if any of the specified fields are left blank, display an error on the screen
	if x == "" {
		f.Errors.Add(field, "This field cannot be blank")
		return false
	}
	return true
}

// checks for string minimum length
func (f *Form) MinLength(field string, length int) bool {
	x := f.Get(field)
	if len(x) < length {
		f.Errors.Add(field, fmt.Sprintf("This field must be at least %d characters long", length))
		return false
	}
	return true
}

// checks for valid email address
func (f *Form) IsEmail(field string) {
	if !govalidator.IsEmail(f.Get(field)) {
		f.Errors.Add(field, "Invalid email address")
	}
}
