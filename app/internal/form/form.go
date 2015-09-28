package form

// Validator is a form that knows how to validate itself. The Validate
// method should write to the receiver's embedded Validation struct
// (see the Validation docs for more info).
type Validator interface {
	// Validate validates the form and writes to the receiver's
	// embedded Validation struct (see the Validation docs for more
	// info).
	Validate()
}

// Validation is a helper type for performing and displaying form
// validation.
//
// Each form should have 2 user-defined types: one type "T" to hold
// *only* the data, and one type that embeds "T" and this type
// (Validation). For example:
//
//  type myForm struct { Name string }
//  type myFormValidated struct { Name string; form.Validation }
//
// The myFormValidated type should implement the Validator
// interface.
type Validation struct {
	// Errors on the overall form (displayed at the top of the form,
	// not next to any specific field).
	Errors []string

	// fieldErrors holds errors on specific form fields (keyed on
	// their name).
	fieldErrors map[string][]string

	// fieldSuccess stores whether specific form fields were
	// successful (keyed on their name).
	fieldSuccess map[string]bool
}

// HasErrors returns true if the form or any of its fields have
// errors.
func (v Validation) HasErrors() bool {
	return len(v.Errors) != 0 || len(v.fieldErrors) != 0
}

// FormError returns the first error on the form, if any (displayed at
// the top of the form, not next to any specific fields).
func (v Validation) FormError() string {
	if len(v.Errors) == 0 {
		return ""
	}
	return v.Errors[0]
}

// AddFieldError adds an error on the named field.
func (v *Validation) AddFieldError(field, error string) {
	if v.fieldErrors == nil {
		v.fieldErrors = map[string][]string{}
	}
	v.fieldErrors[field] = append(v.fieldErrors[field], error)
}

// SetFieldSuccess sets the name field as successful
func (v *Validation) SetFieldSuccess(field string) {
	if v.fieldSuccess == nil {
		v.fieldSuccess = map[string]bool{}
	}
	v.fieldSuccess[field] = true
}

// FieldClass returns the Bootstrap form validation state class
// (has-success, has-warning, or has-error) for the named field.
func (v Validation) FieldClass(field string) string {
	if hasErrors := len(v.fieldErrors[field]) != 0; hasErrors {
		return "has-error"
	}
	if v.fieldSuccess[field] {
		return "has-success"
	}
	return ""
}

// FieldErrors returns the errors, if any, on the named form field.
func (v Validation) FieldErrors(field string) []string {
	return v.fieldErrors[field]
}

// FieldError returns the first error, if any, on the named form field.
func (v Validation) FieldError(field string) string {
	errors := v.fieldErrors[field]
	if len(errors) == 0 {
		return ""
	}
	return errors[0]
}
