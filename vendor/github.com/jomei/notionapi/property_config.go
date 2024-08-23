package notionapi

import (
	"encoding/json"
	"fmt"
)

type PropertyConfigType string

type PropertyConfig interface {
	GetType() PropertyConfigType
	GetID() PropertyID
}

type TitlePropertyConfig struct {
	ID    PropertyID         `json:"id,omitempty"`
	Type  PropertyConfigType `json:"type"`
	Title struct{}           `json:"title"`
}

func (p TitlePropertyConfig) GetType() PropertyConfigType {
	return p.Type
}

func (p TitlePropertyConfig) GetID() PropertyID {
	return p.ID
}

type RichTextPropertyConfig struct {
	ID       PropertyID         `json:"id,omitempty"`
	Type     PropertyConfigType `json:"type"`
	RichText struct{}           `json:"rich_text"`
}

func (p RichTextPropertyConfig) GetType() PropertyConfigType {
	return p.Type
}

func (p RichTextPropertyConfig) GetID() PropertyID {
	return p.ID
}

type NumberPropertyConfig struct {
	ID     PropertyID         `json:"id,omitempty"`
	Type   PropertyConfigType `json:"type"`
	Number NumberFormat       `json:"number"`
}

type FormatType string

func (ft FormatType) String() string {
	return string(ft)
}

type NumberFormat struct {
	Format FormatType `json:"format"`
}

func (p NumberPropertyConfig) GetType() PropertyConfigType {
	return p.Type
}

func (p NumberPropertyConfig) GetID() PropertyID {
	return p.ID
}

type SelectPropertyConfig struct {
	ID     PropertyID         `json:"id,omitempty"`
	Type   PropertyConfigType `json:"type"`
	Select Select             `json:"select"`
}

func (p SelectPropertyConfig) GetType() PropertyConfigType {
	return p.Type
}

func (p SelectPropertyConfig) GetID() PropertyID {
	return p.ID
}

type MultiSelectPropertyConfig struct {
	ID          PropertyID         `json:"id,omitempty"`
	Type        PropertyConfigType `json:"type"`
	MultiSelect Select             `json:"multi_select"`
}

type Select struct {
	Options []Option `json:"options"`
}

func (p MultiSelectPropertyConfig) GetType() PropertyConfigType {
	return p.Type
}

func (p MultiSelectPropertyConfig) GetID() PropertyID {
	return p.ID
}

type DatePropertyConfig struct {
	ID   PropertyID         `json:"id,omitempty"`
	Type PropertyConfigType `json:"type"`
	Date struct{}           `json:"date"`
}

func (p DatePropertyConfig) GetType() PropertyConfigType {
	return p.Type
}

func (p DatePropertyConfig) GetID() PropertyID {
	return p.ID
}

type PeoplePropertyConfig struct {
	ID     PropertyID         `json:"id,omitempty"`
	Type   PropertyConfigType `json:"type"`
	People struct{}           `json:"people"`
}

func (p PeoplePropertyConfig) GetType() PropertyConfigType {
	return p.Type
}

func (p PeoplePropertyConfig) GetID() PropertyID {
	return p.ID
}

type FilesPropertyConfig struct {
	ID    PropertyID         `json:"id,omitempty"`
	Type  PropertyConfigType `json:"type"`
	Files struct{}           `json:"files"`
}

func (p FilesPropertyConfig) GetType() PropertyConfigType {
	return p.Type
}

func (p FilesPropertyConfig) GetID() PropertyID {
	return p.ID
}

type CheckboxPropertyConfig struct {
	ID       PropertyID         `json:"id,omitempty"`
	Type     PropertyConfigType `json:"type"`
	Checkbox struct{}           `json:"checkbox"`
}

func (p CheckboxPropertyConfig) GetType() PropertyConfigType {
	return p.Type
}

func (p CheckboxPropertyConfig) GetID() PropertyID {
	return p.ID
}

type URLPropertyConfig struct {
	ID   PropertyID         `json:"id,omitempty"`
	Type PropertyConfigType `json:"type"`
	URL  struct{}           `json:"url"`
}

func (p URLPropertyConfig) GetType() PropertyConfigType {
	return p.Type
}

func (p URLPropertyConfig) GetID() PropertyID {
	return p.ID
}

type EmailPropertyConfig struct {
	ID    PropertyID         `json:"id,omitempty"`
	Type  PropertyConfigType `json:"type"`
	Email struct{}           `json:"email"`
}

func (p EmailPropertyConfig) GetType() PropertyConfigType {
	return p.Type
}

func (p EmailPropertyConfig) GetID() PropertyID {
	return p.ID
}

type PhoneNumberPropertyConfig struct {
	ID          PropertyID         `json:"id,omitempty"`
	Type        PropertyConfigType `json:"type"`
	PhoneNumber struct{}           `json:"phone_number"`
}

func (p PhoneNumberPropertyConfig) GetType() PropertyConfigType {
	return p.Type
}

func (p PhoneNumberPropertyConfig) GetID() PropertyID {
	return p.ID
}

type FormulaPropertyConfig struct {
	ID      PropertyID         `json:"id,omitempty"`
	Type    PropertyConfigType `json:"type"`
	Formula FormulaConfig      `json:"formula"`
}

type FormulaConfig struct {
	Expression string `json:"expression"`
}

func (p FormulaPropertyConfig) GetType() PropertyConfigType {
	return p.Type
}

func (p FormulaPropertyConfig) GetID() PropertyID {
	return p.ID
}

type RelationPropertyConfig struct {
	Type     PropertyConfigType `json:"type"`
	Relation RelationConfig     `json:"relation"`
}

type RelationConfigType string

func (rp RelationConfigType) String() string {
	return string(rp)
}

type SingleProperty struct{}

type DualProperty struct{}

type RelationConfig struct {
	DatabaseID         DatabaseID         `json:"database_id"`
	SyncedPropertyID   PropertyID         `json:"synced_property_id,omitempty"`
	SyncedPropertyName string             `json:"synced_property_name,omitempty"`
	Type               RelationConfigType `json:"type,omitempty"`
	SingleProperty     *SingleProperty    `json:"single_property,omitempty"`
	DualProperty       *DualProperty      `json:"dual_property,omitempty"`
}

func (p RelationPropertyConfig) GetType() PropertyConfigType {
	return p.Type
}

func (p RelationPropertyConfig) GetID() PropertyID {
	return ""
}

type RollupPropertyConfig struct {
	ID     PropertyID         `json:"id,omitempty"`
	Type   PropertyConfigType `json:"type"`
	Rollup RollupConfig       `json:"rollup"`
}

type RollupConfig struct {
	RelationPropertyName string       `json:"relation_property_name"`
	RelationPropertyID   PropertyID   `json:"relation_property_id"`
	RollupPropertyName   string       `json:"rollup_property_name"`
	RollupPropertyID     PropertyID   `json:"rollup_property_id"`
	Function             FunctionType `json:"function"`
}

func (p RollupPropertyConfig) GetType() PropertyConfigType {
	return p.Type
}

func (p RollupPropertyConfig) GetID() PropertyID {
	return p.ID
}

type CreatedTimePropertyConfig struct {
	ID          PropertyID         `json:"id,omitempty"`
	Type        PropertyConfigType `json:"type"`
	CreatedTime struct{}           `json:"created_time"`
}

func (p CreatedTimePropertyConfig) GetType() PropertyConfigType {
	return p.Type
}

func (p CreatedTimePropertyConfig) GetID() PropertyID {
	return p.ID
}

type CreatedByPropertyConfig struct {
	ID        PropertyID         `json:"id"`
	Type      PropertyConfigType `json:"type"`
	CreatedBy struct{}           `json:"created_by"`
}

func (p CreatedByPropertyConfig) GetType() PropertyConfigType {
	return p.Type
}

func (p CreatedByPropertyConfig) GetID() PropertyID {
	return p.ID
}

type LastEditedTimePropertyConfig struct {
	ID             PropertyID         `json:"id"`
	Type           PropertyConfigType `json:"type"`
	LastEditedTime struct{}           `json:"last_edited_time"`
}

func (p LastEditedTimePropertyConfig) GetType() PropertyConfigType {
	return p.Type
}

func (p LastEditedTimePropertyConfig) GetID() PropertyID {
	return p.ID
}

type LastEditedByPropertyConfig struct {
	ID           PropertyID         `json:"id"`
	Type         PropertyConfigType `json:"type"`
	LastEditedBy struct{}           `json:"last_edited_by"`
}

func (p LastEditedByPropertyConfig) GetType() PropertyConfigType {
	return p.Type
}

func (p LastEditedByPropertyConfig) GetID() PropertyID {
	return p.ID
}

type StatusPropertyConfig struct {
	ID     PropertyID         `json:"id"`
	Type   PropertyConfigType `json:"type"`
	Status StatusConfig       `json:"status"`
}

func (p StatusPropertyConfig) GetType() PropertyConfigType {
	return p.Type
}

func (p StatusPropertyConfig) GetID() PropertyID {
	return p.ID
}

type StatusConfig struct {
	Options []Option      `json:"options"`
	Groups  []GroupConfig `json:"groups"`
}

type GroupConfig struct {
	ID        ObjectID   `json:"id"`
	Name      string     `json:"name"`
	Color     string     `json:"color"`
	OptionIDs []ObjectID `json:"option_ids"`
}

type UniqueIDPropertyConfig struct {
	ID       PropertyID         `json:"id,omitempty"`
	Type     PropertyConfigType `json:"type"`
	UniqueID UniqueIDConfig     `json:"unique_id"`
}

type UniqueIDConfig struct {
	Prefix string `json:"prefix"`
}

func (p UniqueIDPropertyConfig) GetType() PropertyConfigType {
	return ""
}

func (p UniqueIDPropertyConfig) GetID() PropertyID {
	return p.ID
}

type VerificationPropertyConfig struct {
	ID           PropertyID         `json:"id,omitempty"`
	Type         PropertyConfigType `json:"type,omitempty"`
	Verification Verification       `json:"verification"`
}

func (p VerificationPropertyConfig) GetType() PropertyConfigType {
	return p.Type
}

func (p VerificationPropertyConfig) GetID() PropertyID {
	return p.ID
}

type PropertyConfigs map[string]PropertyConfig

func (p *PropertyConfigs) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	props, err := parsePropertyConfigs(raw)
	if err != nil {
		return err
	}

	*p = props
	return nil
}

func parsePropertyConfigs(raw map[string]interface{}) (PropertyConfigs, error) {
	result := make(PropertyConfigs)
	for k, v := range raw {
		var p PropertyConfig
		switch rawProperty := v.(type) {
		case map[string]interface{}:
			switch PropertyConfigType(rawProperty["type"].(string)) {
			case PropertyConfigTypeTitle:
				p = &TitlePropertyConfig{}
			case PropertyConfigTypeRichText:
				p = &RichTextPropertyConfig{}
			case PropertyConfigTypeNumber:
				p = &NumberPropertyConfig{}
			case PropertyConfigTypeSelect:
				p = &SelectPropertyConfig{}
			case PropertyConfigTypeMultiSelect:
				p = &MultiSelectPropertyConfig{}
			case PropertyConfigTypeDate:
				p = &DatePropertyConfig{}
			case PropertyConfigTypePeople:
				p = &PeoplePropertyConfig{}
			case PropertyConfigTypeFiles:
				p = &FilesPropertyConfig{}
			case PropertyConfigTypeCheckbox:
				p = &CheckboxPropertyConfig{}
			case PropertyConfigTypeURL:
				p = &URLPropertyConfig{}
			case PropertyConfigTypeEmail:
				p = &EmailPropertyConfig{}
			case PropertyConfigTypePhoneNumber:
				p = &PhoneNumberPropertyConfig{}
			case PropertyConfigTypeFormula:
				p = &FormulaPropertyConfig{}
			case PropertyConfigTypeRelation:
				p = &RelationPropertyConfig{}
			case PropertyConfigTypeRollup:
				p = &RollupPropertyConfig{}
			case PropertyConfigCreatedTime:
				p = &CreatedTimePropertyConfig{}
			case PropertyConfigCreatedBy:
				p = &CreatedTimePropertyConfig{}
			case PropertyConfigLastEditedTime:
				p = &LastEditedTimePropertyConfig{}
			case PropertyConfigLastEditedBy:
				p = &LastEditedByPropertyConfig{}
			case PropertyConfigStatus:
				p = &StatusPropertyConfig{}
			case PropertyConfigUniqueID:
				p = &UniqueIDPropertyConfig{}
			case PropertyConfigVerification:
				p = &VerificationPropertyConfig{}
			default:

				return nil, fmt.Errorf("unsupported property type: %s", rawProperty["type"].(string))
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
