package notionapi

import (
	"encoding/json"
	"fmt"
	"time"
)

type PropertyType string

type Property interface {
	GetID() string
	GetType() PropertyType
}

type PropertyArray []Property

func (arr *PropertyArray) UnmarshalJSON(data []byte) error {
	var err error
	mapArr := make([]map[string]interface{}, 0)
	if err = json.Unmarshal(data, &mapArr); err != nil {
		return err
	}

	result := make([]Property, len(mapArr))
	for i, prop := range mapArr {
		if result[i], err = decodeProperty(prop); err != nil {
			return err
		}
	}

	if err = json.Unmarshal(data, &result); err != nil {
		return err
	}

	*arr = result
	return nil
}

type TitleProperty struct {
	ID    PropertyID   `json:"id,omitempty"`
	Type  PropertyType `json:"type,omitempty"`
	Title []RichText   `json:"title"`
}

func (p TitleProperty) GetID() string {
	return p.ID.String()
}

func (p TitleProperty) GetType() PropertyType {
	return p.Type
}

type RichTextProperty struct {
	ID       PropertyID   `json:"id,omitempty"`
	Type     PropertyType `json:"type,omitempty"`
	RichText []RichText   `json:"rich_text"`
}

func (p RichTextProperty) GetID() string {
	return p.ID.String()
}

func (p RichTextProperty) GetType() PropertyType {
	return p.Type
}

type TextProperty struct {
	ID   PropertyID   `json:"id,omitempty"`
	Type PropertyType `json:"type,omitempty"`
	Text []RichText   `json:"text"`
}

func (p TextProperty) GetID() string {
	return p.ID.String()
}

func (p TextProperty) GetType() PropertyType {
	return p.Type
}

type NumberProperty struct {
	ID     PropertyID   `json:"id,omitempty"`
	Type   PropertyType `json:"type,omitempty"`
	Number float64      `json:"number"`
}

func (p NumberProperty) GetID() string {
	return p.ID.String()
}

func (p NumberProperty) GetType() PropertyType {
	return p.Type
}

type SelectProperty struct {
	ID     ObjectID     `json:"id,omitempty"`
	Type   PropertyType `json:"type,omitempty"`
	Select Option       `json:"select"`
}

func (p SelectProperty) GetID() string {
	return p.ID.String()
}

func (p SelectProperty) GetType() PropertyType {
	return p.Type
}

type MultiSelectProperty struct {
	ID          ObjectID     `json:"id,omitempty"`
	Type        PropertyType `json:"type,omitempty"`
	MultiSelect []Option     `json:"multi_select"`
}

func (p MultiSelectProperty) GetID() string {
	return p.ID.String()
}

func (p MultiSelectProperty) GetType() PropertyType {
	return p.Type
}

type Option struct {
	ID    PropertyID `json:"id,omitempty"`
	Name  string     `json:"name"`
	Color Color      `json:"color,omitempty"`
}

type DateProperty struct {
	ID   ObjectID     `json:"id,omitempty"`
	Type PropertyType `json:"type,omitempty"`
	Date *DateObject  `json:"date"`
}

type DateObject struct {
	Start *Date `json:"start"`
	End   *Date `json:"end"`
}

func (p DateProperty) GetID() string {
	return p.ID.String()
}

func (p DateProperty) GetType() PropertyType {
	return p.Type
}

type FormulaProperty struct {
	ID      ObjectID     `json:"id,omitempty"`
	Type    PropertyType `json:"type,omitempty"`
	Formula Formula      `json:"formula"`
}

type FormulaType string

type Formula struct {
	Type    FormulaType `json:"type,omitempty"`
	String  string      `json:"string,omitempty"`
	Number  float64     `json:"number,omitempty"`
	Boolean bool        `json:"boolean,omitempty"`
	Date    *DateObject `json:"date,omitempty"`
}

func (p FormulaProperty) GetID() string {
	return p.ID.String()
}

func (p FormulaProperty) GetType() PropertyType {
	return p.Type
}

type RelationProperty struct {
	ID       ObjectID     `json:"id,omitempty"`
	Type     PropertyType `json:"type,omitempty"`
	Relation []Relation   `json:"relation"`
}

type Relation struct {
	ID PageID `json:"id"`
}

func (p RelationProperty) GetID() string {
	return p.ID.String()
}

func (p RelationProperty) GetType() PropertyType {
	return p.Type
}

type RollupProperty struct {
	ID     ObjectID     `json:"id,omitempty"`
	Type   PropertyType `json:"type,omitempty"`
	Rollup Rollup       `json:"rollup"`
}

type RollupType string

type Rollup struct {
	Type   RollupType    `json:"type,omitempty"`
	Number float64       `json:"number,omitempty"`
	Date   *DateObject   `json:"date,omitempty"`
	Array  PropertyArray `json:"array,omitempty"`
}

func (p RollupProperty) GetID() string {
	return p.ID.String()
}

func (p RollupProperty) GetType() PropertyType {
	return p.Type
}

type PeopleProperty struct {
	ID     ObjectID     `json:"id,omitempty"`
	Type   PropertyType `json:"type,omitempty"`
	People []User       `json:"people"`
}

func (p PeopleProperty) GetID() string {
	return p.ID.String()
}

func (p PeopleProperty) GetType() PropertyType {
	return p.Type
}

type FilesProperty struct {
	ID    ObjectID     `json:"id,omitempty"`
	Type  PropertyType `json:"type,omitempty"`
	Files []File       `json:"files"`
}

func (p FilesProperty) GetID() string {
	return p.ID.String()
}

func (p FilesProperty) GetType() PropertyType {
	return p.Type
}

type CheckboxProperty struct {
	ID       ObjectID     `json:"id,omitempty"`
	Type     PropertyType `json:"type,omitempty"`
	Checkbox bool         `json:"checkbox"`
}

func (p CheckboxProperty) GetID() string {
	return p.ID.String()
}

func (p CheckboxProperty) GetType() PropertyType {
	return p.Type
}

type URLProperty struct {
	ID   ObjectID     `json:"id,omitempty"`
	Type PropertyType `json:"type,omitempty"`
	URL  string       `json:"url"`
}

func (p URLProperty) GetID() string {
	return p.ID.String()
}

func (p URLProperty) GetType() PropertyType {
	return p.Type
}

type EmailProperty struct {
	ID    PropertyID   `json:"id,omitempty"`
	Type  PropertyType `json:"type,omitempty"`
	Email string       `json:"email"`
}

func (p EmailProperty) GetID() string {
	return p.ID.String()
}

func (p EmailProperty) GetType() PropertyType {
	return p.Type
}

type PhoneNumberProperty struct {
	ID          ObjectID     `json:"id,omitempty"`
	Type        PropertyType `json:"type,omitempty"`
	PhoneNumber string       `json:"phone_number"`
}

func (p PhoneNumberProperty) GetID() string {
	return p.ID.String()
}

func (p PhoneNumberProperty) GetType() PropertyType {
	return p.Type
}

type CreatedTimeProperty struct {
	ID          ObjectID     `json:"id,omitempty"`
	Type        PropertyType `json:"type,omitempty"`
	CreatedTime time.Time    `json:"created_time"`
}

func (p CreatedTimeProperty) GetID() string {
	return p.ID.String()
}

func (p CreatedTimeProperty) GetType() PropertyType {
	return p.Type
}

type CreatedByProperty struct {
	ID        ObjectID     `json:"id,omitempty"`
	Type      PropertyType `json:"type,omitempty"`
	CreatedBy User         `json:"created_by"`
}

func (p CreatedByProperty) GetID() string {
	return p.ID.String()
}

func (p CreatedByProperty) GetType() PropertyType {
	return p.Type
}

type LastEditedTimeProperty struct {
	ID             ObjectID     `json:"id,omitempty"`
	Type           PropertyType `json:"type,omitempty"`
	LastEditedTime time.Time    `json:"last_edited_time"`
}

func (p LastEditedTimeProperty) GetID() string {
	return p.ID.String()
}

func (p LastEditedTimeProperty) GetType() PropertyType {
	return p.Type
}

type LastEditedByProperty struct {
	ID           ObjectID     `json:"id,omitempty"`
	Type         PropertyType `json:"type,omitempty"`
	LastEditedBy User         `json:"last_edited_by"`
}

func (p LastEditedByProperty) GetID() string {
	return p.ID.String()
}

func (p LastEditedByProperty) GetType() PropertyType {
	return p.Type
}

type StatusProperty struct {
	ID     ObjectID     `json:"id,omitempty"`
	Type   PropertyType `json:"type,omitempty"`
	Status Status       `json:"status"`
}

func (p StatusProperty) GetID() string {
	return p.ID.String()
}

func (p StatusProperty) GetType() PropertyType {
	return p.Type
}

type UniqueIDProperty struct {
	ID       ObjectID     `json:"id,omitempty"`
	Type     PropertyType `json:"type,omitempty"`
	UniqueID UniqueID     `json:"unique_id"`
}

func (p UniqueIDProperty) GetID() string {
	return p.ID.String()
}

func (p UniqueIDProperty) GetType() PropertyType {
	return p.Type
}

type VerificationProperty struct {
	ID           ObjectID     `json:"id,omitempty"`
	Type         PropertyType `json:"type,omitempty"`
	Verification Verification `json:"verification"`
}

func (p VerificationProperty) GetID() string {
	return p.ID.String()
}

func (p VerificationProperty) GetType() PropertyType {
	return p.Type
}

type ButtonProperty struct {
	ID     ObjectID     `json:"id,omitempty"`
	Type   PropertyType `json:"type,omitempty"`
	Button Button       `json:"button"`
}

func (p ButtonProperty) GetID() string {
	return p.ID.String()
}

func (p ButtonProperty) GetType() PropertyType {
	return p.Type
}

type Properties map[string]Property

func (p *Properties) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	props, err := parsePageProperties(raw)
	if err != nil {
		return err
	}

	*p = props
	return nil
}

func parsePageProperties(raw map[string]interface{}) (map[string]Property, error) {
	result := make(map[string]Property)
	for k, v := range raw {
		switch rawProperty := v.(type) {
		case map[string]interface{}:
			p, err := decodeProperty(rawProperty)
			if err != nil {
				return nil, err
			}
			b, err := json.Marshal(rawProperty)
			if err != nil {
				return nil, err
			}

			if err = json.Unmarshal(b, &p); err != nil {
				return nil, err
			}

			result[k] = p
		default:
			return nil, fmt.Errorf("unsupported property format %T", v)
		}
	}

	return result, nil
}

func decodeProperty(raw map[string]interface{}) (Property, error) {
	var p Property
	switch PropertyType(raw["type"].(string)) {
	case PropertyTypeTitle:
		p = &TitleProperty{}
	case PropertyTypeRichText:
		p = &RichTextProperty{}
	case PropertyTypeText:
		p = &RichTextProperty{}
	case PropertyTypeNumber:
		p = &NumberProperty{}
	case PropertyTypeSelect:
		p = &SelectProperty{}
	case PropertyTypeMultiSelect:
		p = &MultiSelectProperty{}
	case PropertyTypeDate:
		p = &DateProperty{}
	case PropertyTypeFormula:
		p = &FormulaProperty{}
	case PropertyTypeRelation:
		p = &RelationProperty{}
	case PropertyTypeRollup:
		p = &RollupProperty{}
	case PropertyTypePeople:
		p = &PeopleProperty{}
	case PropertyTypeFiles:
		p = &FilesProperty{}
	case PropertyTypeCheckbox:
		p = &CheckboxProperty{}
	case PropertyTypeURL:
		p = &URLProperty{}
	case PropertyTypeEmail:
		p = &EmailProperty{}
	case PropertyTypePhoneNumber:
		p = &PhoneNumberProperty{}
	case PropertyTypeCreatedTime:
		p = &CreatedTimeProperty{}
	case PropertyTypeCreatedBy:
		p = &CreatedByProperty{}
	case PropertyTypeLastEditedTime:
		p = &LastEditedTimeProperty{}
	case PropertyTypeLastEditedBy:
		p = &LastEditedByProperty{}
	case PropertyTypeStatus:
		p = &StatusProperty{}
	case PropertyTypeUniqueID:
		p = &UniqueIDProperty{}
	case PropertyTypeVerification:
		p = &VerificationProperty{}
	case PropertyTypeButton:
		p = &ButtonProperty{}
	default:
		return nil, fmt.Errorf("unsupported property type: %s", raw["type"].(string))
	}

	return p, nil
}
