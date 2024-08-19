package notionapi

const (
	ObjectTypeDatabase ObjectType = "database"
	ObjectTypeBlock    ObjectType = "block"
	ObjectTypePage     ObjectType = "page"
	ObjectTypeList     ObjectType = "list"
	ObjectTypeText     ObjectType = "text"
	ObjectTypeUser     ObjectType = "user"
	ObjectTypeError    ObjectType = "error"
	ObjectTypeComment  ObjectType = "comment"
)

const (
	PropertyConfigTypeTitle       PropertyConfigType = "title"
	PropertyConfigTypeRichText    PropertyConfigType = "rich_text"
	PropertyConfigTypeNumber      PropertyConfigType = "number"
	PropertyConfigTypeSelect      PropertyConfigType = "select"
	PropertyConfigTypeMultiSelect PropertyConfigType = "multi_select"
	PropertyConfigTypeDate        PropertyConfigType = "date"
	PropertyConfigTypePeople      PropertyConfigType = "people"
	PropertyConfigTypeFiles       PropertyConfigType = "files"
	PropertyConfigTypeCheckbox    PropertyConfigType = "checkbox"
	PropertyConfigTypeURL         PropertyConfigType = "url"
	PropertyConfigTypeEmail       PropertyConfigType = "email"
	PropertyConfigTypePhoneNumber PropertyConfigType = "phone_number"
	PropertyConfigTypeFormula     PropertyConfigType = "formula"
	PropertyConfigTypeRelation    PropertyConfigType = "relation"
	PropertyConfigTypeRollup      PropertyConfigType = "rollup"
	PropertyConfigCreatedTime     PropertyConfigType = "created_time"
	PropertyConfigCreatedBy       PropertyConfigType = "created_by"
	PropertyConfigLastEditedTime  PropertyConfigType = "last_edited_time"
	PropertyConfigLastEditedBy    PropertyConfigType = "last_edited_by"
	PropertyConfigStatus          PropertyConfigType = "status"
	PropertyConfigUniqueID        PropertyConfigType = "unique_id"
	PropertyConfigVerification    PropertyConfigType = "verification"
)

const (
	PropertyTypeTitle          PropertyType = "title"
	PropertyTypeRichText       PropertyType = "rich_text"
	PropertyTypeText           PropertyType = "text"
	PropertyTypeNumber         PropertyType = "number"
	PropertyTypeSelect         PropertyType = "select"
	PropertyTypeMultiSelect    PropertyType = "multi_select"
	PropertyTypeDate           PropertyType = "date"
	PropertyTypeFormula        PropertyType = "formula"
	PropertyTypeRelation       PropertyType = "relation"
	PropertyTypeRollup         PropertyType = "rollup"
	PropertyTypePeople         PropertyType = "people"
	PropertyTypeFiles          PropertyType = "files"
	PropertyTypeCheckbox       PropertyType = "checkbox"
	PropertyTypeURL            PropertyType = "url"
	PropertyTypeEmail          PropertyType = "email"
	PropertyTypePhoneNumber    PropertyType = "phone_number"
	PropertyTypeCreatedTime    PropertyType = "created_time"
	PropertyTypeCreatedBy      PropertyType = "created_by"
	PropertyTypeLastEditedTime PropertyType = "last_edited_time"
	PropertyTypeLastEditedBy   PropertyType = "last_edited_by"
	PropertyTypeStatus         PropertyType = "status"
	PropertyTypeUniqueID       PropertyType = "unique_id"
	PropertyTypeVerification   PropertyType = "verification"
	PropertyTypeButton         PropertyType = "button"
)

const (
	FormatNumber           FormatType = "number"
	FormatNumberWithCommas FormatType = "number_with_commas"
	FormatPercent          FormatType = "percent"
	FormatDollar           FormatType = "dollar"
	FormatCanadianDollar   FormatType = "canadian_dollar"
	FormatEuro             FormatType = "euro"
	FormatPound            FormatType = "pound"
	FormatYen              FormatType = "yen"
	FormatRuble            FormatType = "ruble"
	FormatRupee            FormatType = "rupee"
	FormatWon              FormatType = "won"
	FormatYuan             FormatType = "yuan"
	FormatReal             FormatType = "real"
	FormatLira             FormatType = "lira"
	FormatRupiah           FormatType = "rupiah"
	FormatFranc            FormatType = "franc"
	FormatHongKongDollar   FormatType = "hong_kong_dollar"
	FormatNewZealandDollar FormatType = "hong_kong_dollar"
	FormatKrona            FormatType = "krona"
	FormatNorwegianKrone   FormatType = "norwegian_krone"
	FormatMexicanPeso      FormatType = "mexican_peso"
	FormatRand             FormatType = "rand"
	FormatNewTaiwanDollar  FormatType = "new_taiwan_dollar"
	FormatDanishKrone      FormatType = "danish_krone"
	FormatZloty            FormatType = "zloty"
	FormatBath             FormatType = "baht"
	FormatForint           FormatType = "forint"
	FormatKoruna           FormatType = "koruna"
	FormatShekel           FormatType = "shekel"
	FormatChileanPeso      FormatType = "chilean_peso"
	FormatPhilippinePeso   FormatType = "philippine_peso"
	FormatDirham           FormatType = "dirham"
	FormatColombianPeso    FormatType = "colombian_peso"
	FormatRiyal            FormatType = "riyal"
	FormatRinggit          FormatType = "ringgit"
	FormatLeu              FormatType = "leu"
	FormatArgentinePeso    FormatType = "argentine_peso"
	FormatUruguayanPeso    FormatType = "uruguayan_peso"
	FormatSingaporeDollar  FormatType = "singapore_dollar"
)

const (
	ColorDefault           Color = "default"
	ColorGray              Color = "gray"
	ColorBrown             Color = "brown"
	ColorOrange            Color = "orange"
	ColorYellow            Color = "yellow"
	ColorGreen             Color = "green"
	ColorBlue              Color = "blue"
	ColorPurple            Color = "purple"
	ColorPink              Color = "pink"
	ColorRed               Color = "red"
	ColorDefaultBackground Color = "default_background"
	ColorGrayBackground    Color = "gray_background"
	ColorBrownBackground   Color = "brown_background"
	ColorOrangeBackground  Color = "orange_background"
	ColorYellowBackground  Color = "yellow_background"
	ColorGreenBackground   Color = "green_background"
	ColorBlueBackground    Color = "blue_background"
	ColorPurpleBackground  Color = "purple_background"
	ColorPinkBackground    Color = "pink_background"
	ColorRedBackground     Color = "red_background"
)

const (
	FilterOperatorAND FilterOperator = "and"
	FilterOperatorOR  FilterOperator = "or"
)

const (
	FunctionCountAll          FunctionType = "count_all"
	FunctionCountValues       FunctionType = "count_values"
	FunctionCountUniqueValues FunctionType = "count_unique_values"
	FunctionCountEmpty        FunctionType = "count_empty"
	FunctionCountNotEmpty     FunctionType = "count_not_empty"
	FunctionPercentEmpty      FunctionType = "percent_empty"
	FunctionPercentNotEmpty   FunctionType = "percent_not_empty"
	FunctionSum               FunctionType = "sum"
	FunctionAverage           FunctionType = "average"
	FunctionMedian            FunctionType = "median"
	FunctionMin               FunctionType = "min"
	FunctionMax               FunctionType = "max"
	FunctionRange             FunctionType = "range"
)

const (
	ConditionEquals         Condition = "equals"
	ConditionDoesNotEqual   Condition = "does_not_equal"
	ConditionContains       Condition = "contains"
	ConditionDoesNotContain Condition = "does_not_contain"
	ConditionDoesStartsWith Condition = "starts_with"
	ConditionDoesEndsWith   Condition = "ends_with"
	ConditionDoesIsEmpty    Condition = "is_empty"
	ConditionGreaterThan    Condition = "greater_than"
	ConditionLessThan       Condition = "less_than"

	ConditionGreaterThanOrEqualTo Condition = "greater_than_or_equal_to"
	ConditionLessThanOrEqualTo    Condition = "greater_than_or_equal_to"

	ConditionBefore     Condition = "before"
	ConditionAfter      Condition = "after"
	ConditionOnOrBefore Condition = "on_or_before"
	ConditionOnOrAfter  Condition = "on_or_after"
	ConditionPastWeek   Condition = "past_week"
	ConditionPastMonth  Condition = "past_month"
	ConditionPastYear   Condition = "past_year"
	ConditionNextWeek   Condition = "next_week"
	ConditionNextMonth  Condition = "next_month"
	ConditionNextYear   Condition = "next_year"

	ConditionText     Condition = "text"
	ConditionCheckbox Condition = "checkbox"
	ConditionNumber   Condition = "number"
	ConditionDate     Condition = "date"
)

const (
	TimestampCreated    TimestampType = "created_time"
	TimestampLastEdited TimestampType = "last_edited_time"
)

const (
	SortOrderASC  SortOrder = "ascending"
	SortOrderDESC SortOrder = "descending"
)

const (
	ParentTypeDatabaseID ParentType = "database_id"
	ParentTypePageID     ParentType = "page_id"
	ParentTypeWorkspace  ParentType = "workspace"
	ParentTypeBlockID    ParentType = "block_id"
)

const (
	UserTypePerson UserType = "person"
	UserTypeBot    UserType = "bot"
)

// See https://developers.notion.com/reference/block
const (
	BlockTypeParagraph BlockType = "paragraph"
	BlockTypeHeading1  BlockType = "heading_1"
	BlockTypeHeading2  BlockType = "heading_2"
	BlockTypeHeading3  BlockType = "heading_3"

	BlockTypeBulletedListItem BlockType = "bulleted_list_item"
	BlockTypeNumberedListItem BlockType = "numbered_list_item"

	BlockTypeToDo          BlockType = "to_do"
	BlockTypeToggle        BlockType = "toggle"
	BlockTypeChildPage     BlockType = "child_page"
	BlockTypeChildDatabase BlockType = "child_database"

	BlockTypeEmbed           BlockType = "embed"
	BlockTypeImage           BlockType = "image"
	BlockTypeVideo           BlockType = "video"
	BlockTypeFile            BlockType = "file"
	BlockTypePdf             BlockType = "pdf"
	BlockTypeBookmark        BlockType = "bookmark"
	BlockTypeCode            BlockType = "code"
	BlockTypeDivider         BlockType = "divider"
	BlockCallout             BlockType = "callout"
	BlockQuote               BlockType = "quote"
	BlockTypeTableOfContents BlockType = "table_of_contents"
	BlockTypeEquation        BlockType = "equation"
	BlockTypeBreadcrumb      BlockType = "breadcrumb"
	BlockTypeColumn          BlockType = "column"
	BlockTypeColumnList      BlockType = "column_list"
	BlockTypeLinkPreview     BlockType = "link_preview"
	BlockTypeLinkToPage      BlockType = "link_to_page"
	BlockTypeTemplate        BlockType = "template"
	BlockTypeSyncedBlock     BlockType = "synced_block"
	BlockTypeTableBlock      BlockType = "table"
	BlockTypeTableRowBlock   BlockType = "table_row"
	BlockTypeUnsupported     BlockType = "unsupported"
)

const (
	FileTypeFile     FileType = "file"
	FileTypeExternal FileType = "external"
)

const (
	FormulaTypeString  FormulaType = "string"
	FormulaTypeNumber  FormulaType = "number"
	FormulaTypeBoolean FormulaType = "boolean"
	FormulaTypeDate    FormulaType = "date"
)

const (
	RollupTypeNumber RollupType = "number"
	RollupTypeDate   RollupType = "date"
	RollupTypeArray  RollupType = "array"
)

const (
	MentionTypeDatabase        MentionType = "database"
	MentionTypePage            MentionType = "page"
	MentionTypeUser            MentionType = "user"
	MentionTypeDate            MentionType = "date"
	MentionTypeTemplateMention MentionType = "template_mention"
)

const (
	TemplateMentionTypeUser TemplateMentionType = "template_mention_user"
	TemplateMentionTypeDate TemplateMentionType = "template_mention_date"
)

const (
	RelationSingleProperty RelationConfigType = "single_property"
	RelationDualProperty   RelationConfigType = "dual_property"
)

const (
	VerificationStateVerified   VerificationState = "verified"
	VerificationStateUnverified VerificationState = "unverified"
)
