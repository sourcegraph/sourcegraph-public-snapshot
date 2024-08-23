package notionapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type BlockID string

func (bID BlockID) String() string {
	return string(bID)
}

type BlockService interface {
	AppendChildren(context.Context, BlockID, *AppendBlockChildrenRequest) (*AppendBlockChildrenResponse, error)
	Get(context.Context, BlockID) (Block, error)
	GetChildren(context.Context, BlockID, *Pagination) (*GetChildrenResponse, error)
	Update(ctx context.Context, id BlockID, request *BlockUpdateRequest) (Block, error)
	Delete(context.Context, BlockID) (Block, error)
}

type BlockClient struct {
	apiClient *Client
}

// Creates and appends new children blocks to the parent block_id specified.
// Blocks can be parented by other blocks, pages, or databases.

// Returns a paginated list of newly created first level children block objects.

// Existing blocks cannot be moved using this endpoint. Blocks are appended to
// the bottom of the parent block. Once a block is appended as a child, it can't
// be moved elsewhere via the API.

// For blocks that allow children, we allow up to two levels of nesting in a
// single request.
//
// See https://developers.notion.com/reference/patch-block-children
func (bc *BlockClient) AppendChildren(ctx context.Context, id BlockID, requestBody *AppendBlockChildrenRequest) (*AppendBlockChildrenResponse, error) {
	res, err := bc.apiClient.request(ctx, http.MethodPatch, fmt.Sprintf("blocks/%s/children", id.String()), nil, requestBody)
	if err != nil {
		return nil, err
	}

	defer func() {
		if errClose := res.Body.Close(); errClose != nil {
			log.Println("failed to close body, should never happen")
		}
	}()

	var response AppendBlockChildrenResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

type AppendBlockChildrenRequest struct {
	// Append new children after a specific block. If empty, new children with be appended to the bottom of the parent block.
	After BlockID `json:"after,omitempty"`
	// Child content to append to a container block as an array of block objects.
	Children []Block `json:"children"`
}

// Retrieves a Block object using the ID specified.
//
// Get https://developers.notion.com/reference/retrieve-a-block
func (bc *BlockClient) Get(ctx context.Context, id BlockID) (Block, error) {
	res, err := bc.apiClient.request(ctx, http.MethodGet, fmt.Sprintf("blocks/%s", id.String()), nil, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		if errClose := res.Body.Close(); errClose != nil {
			log.Println("failed to close body, should never happen")
		}
	}()

	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return decodeBlock(response)
}

// Returns a paginated array of child block objects contained in the block using
// the ID specified. In order to receive a complete representation of a block,
// you may need to recursively retrieve the block children of child blocks.
//
// See https://developers.notion.com/reference/get-block-children
func (bc *BlockClient) GetChildren(ctx context.Context, id BlockID, pagination *Pagination) (*GetChildrenResponse, error) {
	res, err := bc.apiClient.request(ctx, http.MethodGet, fmt.Sprintf("blocks/%s/children", id.String()), pagination.ToQuery(), nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		if errClose := res.Body.Close(); errClose != nil {
			log.Println("failed to close body, should never happen")
		}
	}()

	response := &GetChildrenResponse{}
	err = json.NewDecoder(res.Body).Decode(response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

type GetChildrenResponse struct {
	Object     ObjectType `json:"object"`
	Results    Blocks     `json:"results"`
	NextCursor string     `json:"next_cursor"`
	HasMore    bool       `json:"has_more"`
}

// Updates the content for the specified block_id based on the block type.
// Supported fields based on the block object type (see Block object for
// available fields and the expected input for each field).
//
// Note: The update replaces the entire value for a given field. If a field is
// omitted (ex: omitting checked when updating a to_do block), the value will not be changed.
//
// See https://developers.notion.com/reference/update-a-block
func (bc *BlockClient) Update(ctx context.Context, id BlockID, requestBody *BlockUpdateRequest) (Block, error) {
	res, err := bc.apiClient.request(ctx, http.MethodPatch, fmt.Sprintf("blocks/%s", id.String()), nil, requestBody)
	if err != nil {
		return nil, err
	}

	defer func() {
		if errClose := res.Body.Close(); errClose != nil {
			log.Println("failed to close body, should never happen")
		}
	}()

	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return decodeBlock(response)
}

type BlockUpdateRequest struct {
	Paragraph        *Paragraph `json:"paragraph,omitempty"`
	Heading1         *Heading   `json:"heading_1,omitempty"`
	Heading2         *Heading   `json:"heading_2,omitempty"`
	Heading3         *Heading   `json:"heading_3,omitempty"`
	BulletedListItem *ListItem  `json:"bulleted_list_item,omitempty"`
	NumberedListItem *ListItem  `json:"numbered_list_item,omitempty"`
	Code             *Code      `json:"code,omitempty"`
	ToDo             *ToDo      `json:"to_do,omitempty"`
	Toggle           *Toggle    `json:"toggle,omitempty"`
	Embed            *Embed     `json:"embed,omitempty"`
	Image            *Image     `json:"image,omitempty"`
	Video            *Video     `json:"video,omitempty"`
	File             *BlockFile `json:"file,omitempty"`
	Pdf              *Pdf       `json:"pdf,omitempty"`
	Bookmark         *Bookmark  `json:"bookmark,omitempty"`
	Template         *Template  `json:"template,omitempty"`
	Callout          *Callout   `json:"callout,omitempty"`
	Equation         *Equation  `json:"equation,omitempty"`
	Quote            *Quote     `json:"quote,omitempty"`
	TableRow         *TableRow  `json:"table_row,omitempty"`
}

// Sets a Block object, including page blocks, to archived: true using the ID
// specified. Note: in the Notion UI application, this moves the block to the
// "Trash" where it can still be accessed and restored.
//
// To restore the block with the API, use the Update a block or Update page respectively.
//
// See https://developers.notion.com/reference/delete-a-block
func (bc *BlockClient) Delete(ctx context.Context, id BlockID) (Block, error) {
	res, err := bc.apiClient.request(ctx, http.MethodDelete, fmt.Sprintf("blocks/%s", id.String()), nil, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		if errClose := res.Body.Close(); errClose != nil {
			log.Println("failed to close body, should never happen")
		}
	}()

	var response map[string]interface{}
	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return decodeBlock(response)
}

type BlockType string

func (bt BlockType) String() string {
	return string(bt)
}

type Block interface {
	GetType() BlockType
	GetID() BlockID
	GetObject() ObjectType
	GetCreatedTime() *time.Time
	GetLastEditedTime() *time.Time
	GetCreatedBy() *User
	GetLastEditedBy() *User
	GetHasChildren() bool
	GetArchived() bool
	GetParent() *Parent
	GetRichTextString() string
}

type Blocks []Block

func (b *Blocks) UnmarshalJSON(data []byte) error {
	var err error
	mapArr := make([]map[string]interface{}, 0)
	if err = json.Unmarshal(data, &mapArr); err != nil {
		return err
	}

	result := make([]Block, len(mapArr))
	for i, prop := range mapArr {
		if result[i], err = decodeBlock(prop); err != nil {
			return err
		}
	}

	*b = result
	return nil
}

// BasicBlock defines the common fields of all Notion block types.
// See https://developers.notion.com/reference/block for the list.
// BasicBlock implements the Block interface.
type BasicBlock struct {
	Object         ObjectType `json:"object"`
	ID             BlockID    `json:"id,omitempty"`
	Type           BlockType  `json:"type"`
	CreatedTime    *time.Time `json:"created_time,omitempty"`
	LastEditedTime *time.Time `json:"last_edited_time,omitempty"`
	CreatedBy      *User      `json:"created_by,omitempty"`
	LastEditedBy   *User      `json:"last_edited_by,omitempty"`
	HasChildren    bool       `json:"has_children,omitempty"`
	Archived       bool       `json:"archived,omitempty"`
	Parent         *Parent    `json:"parent,omitempty"`
}

func (b BasicBlock) GetType() BlockType {
	return b.Type
}

func (b BasicBlock) GetID() BlockID {
	return b.ID
}

func (b BasicBlock) GetObject() ObjectType {
	return b.Object
}

func (b BasicBlock) GetCreatedTime() *time.Time {
	return b.CreatedTime
}

func (b BasicBlock) GetLastEditedTime() *time.Time {
	return b.LastEditedTime
}

func (b BasicBlock) GetCreatedBy() *User {
	return b.CreatedBy
}

func (b BasicBlock) GetLastEditedBy() *User {
	return b.LastEditedBy
}

func (b BasicBlock) GetHasChildren() bool {
	return b.HasChildren
}

func (b BasicBlock) GetArchived() bool {
	return b.Archived
}

func (b BasicBlock) GetParent() *Parent {
	return b.Parent
}
func concatenateRichText(richtext []RichText) string {
	var result string
	for _, rt := range richtext {
		result += rt.PlainText
	}
	return result
}

func (h Heading1Block) GetRichTextString() string {
	return concatenateRichText(h.Heading1.RichText)
}

func (p ParagraphBlock) GetRichTextString() string {
	return concatenateRichText(p.Paragraph.RichText)
}

func (h Heading2Block) GetRichTextString() string {
	return concatenateRichText(h.Heading2.RichText)
}

func (h Heading3Block) GetRichTextString() string {
	return concatenateRichText(h.Heading3.RichText)
}

func (c CalloutBlock) GetRichTextString() string {
	return concatenateRichText(c.Callout.RichText)
}

func (q QuoteBlock) GetRichTextString() string {
	return concatenateRichText(q.Quote.RichText)
}

func (b BulletedListItemBlock) GetRichTextString() string {
	return concatenateRichText(b.BulletedListItem.RichText)
}

func (n NumberedListItemBlock) GetRichTextString() string {
	return concatenateRichText(n.NumberedListItem.RichText)
}

func (t ToDoBlock) GetRichTextString() string {
	return concatenateRichText(t.ToDo.RichText)
}

func (b ToggleBlock) GetRichTextString() string {
	return concatenateRichText(b.Toggle.RichText)
}

func (b EmbedBlock) GetRichTextString() string {
	return concatenateRichText(b.Embed.Caption)
}

func (b ImageBlock) GetRichTextString() string {
	return concatenateRichText(b.Image.Caption)
}

func (b AudioBlock) GetRichTextString() string {
	return concatenateRichText(b.Audio.Caption)
}

func (b VideoBlock) GetRichTextString() string {
	return concatenateRichText(b.Video.Caption)
}

func (b FileBlock) GetRichTextString() string {
	return concatenateRichText(b.File.Caption)
}

func (b PdfBlock) GetRichTextString() string {
	return concatenateRichText(b.Pdf.Caption)
}

func (b BookmarkBlock) GetRichTextString() string {
	return concatenateRichText(b.Bookmark.Caption)
}

func (b TemplateBlock) GetRichTextString() string {
	return concatenateRichText(b.Template.RichText)
}

func (b LinkPreviewBlock) GetRichTextString() string {
	return b.LinkPreview.URL
}

func (b EquationBlock) GetRichTextString() string {
	return b.Equation.Expression
}

func (b BasicBlock) GetRichTextString() string {
	return "No rich text of a basic block."
}
var _ Block = (*BasicBlock)(nil)

type ParagraphBlock struct {
	BasicBlock
	Paragraph Paragraph `json:"paragraph"`
}

type Paragraph struct {
	RichText []RichText `json:"rich_text"`
	Children Blocks     `json:"children,omitempty"`
	Color    string     `json:"color,omitempty"`
}

type Heading1Block struct {
	BasicBlock
	Heading1 Heading `json:"heading_1"`
}

type Heading struct {
	RichText     []RichText `json:"rich_text"`
	Children     Blocks     `json:"children,omitempty"`
	Color        string     `json:"color,omitempty"`
	IsToggleable bool       `json:"is_toggleable,omitempty"`
}

type Heading2Block struct {
	BasicBlock
	Heading2 Heading `json:"heading_2"`
}

type Heading3Block struct {
	BasicBlock
	Heading3 Heading `json:"heading_3"`
}

type CalloutBlock struct {
	BasicBlock
	Callout Callout `json:"callout"`
}

type Callout struct {
	RichText []RichText `json:"rich_text"`
	Icon     *Icon      `json:"icon,omitempty"`
	Children Blocks     `json:"children,omitempty"`
	Color    string     `json:"color,omitempty"`
}

type QuoteBlock struct {
	BasicBlock
	Quote Quote `json:"quote"`
}

type Quote struct {
	RichText []RichText `json:"rich_text"`
	Children Blocks     `json:"children,omitempty"`
	Color    string     `json:"color,omitempty"`
}

type TableBlock struct {
	BasicBlock
	Table Table `json:"table"`
}

type Table struct {
	TableWidth      int    `json:"table_width"`
	HasColumnHeader bool   `json:"has_column_header"`
	HasRowHeader    bool   `json:"has_row_header"`
	Children        Blocks `json:"children,omitempty"`
}

type TableRowBlock struct {
	BasicBlock
	TableRow TableRow `json:"table_row"`
}

type TableRow struct {
	Cells [][]RichText `json:"cells"`
}

type BulletedListItemBlock struct {
	BasicBlock
	BulletedListItem ListItem `json:"bulleted_list_item"`
}

type ListItem struct {
	RichText []RichText `json:"rich_text"`
	Children Blocks     `json:"children,omitempty"`
	Color    string     `json:"color,omitempty"`
}

type NumberedListItemBlock struct {
	BasicBlock
	NumberedListItem ListItem `json:"numbered_list_item"`
}

type ToDoBlock struct {
	BasicBlock
	ToDo ToDo `json:"to_do"`
}

type ToDo struct {
	RichText []RichText `json:"rich_text"`
	Children Blocks     `json:"children,omitempty"`
	Checked  bool       `json:"checked"`
	Color    string     `json:"color,omitempty"`
}

type ToggleBlock struct {
	BasicBlock
	Toggle Toggle `json:"toggle"`
}

type Toggle struct {
	RichText []RichText `json:"rich_text"`
	Children Blocks     `json:"children,omitempty"`
	Color    string     `json:"color,omitempty"`
}

type ChildPageBlock struct {
	BasicBlock
	ChildPage struct {
		Title string `json:"title"`
	} `json:"child_page"`
}

type EmbedBlock struct {
	BasicBlock
	Embed Embed `json:"embed"`
}

type Embed struct {
	Caption []RichText `json:"caption,omitempty"`
	URL     string     `json:"url"`
}

type ImageBlock struct {
	BasicBlock
	Image Image `json:"image"`
}

type Image struct {
	Caption  []RichText  `json:"caption,omitempty"`
	Type     FileType    `json:"type,omitempty"`
	File     *FileObject `json:"file,omitempty"`
	External *FileObject `json:"external,omitempty"`
}

// GetURL returns the external or internal URL depending on the image type.
func (i Image) GetURL() string {
	if i.File != nil {
		return i.File.URL
	}
	if i.External != nil {
		return i.External.URL
	}
	return ""
}

type AudioBlock struct {
	BasicBlock
	Audio Audio `json:"audio"`
}

type Audio struct {
	Caption  []RichText  `json:"caption,omitempty"`
	Type     FileType    `json:"type"`
	File     *FileObject `json:"file,omitempty"`
	External *FileObject `json:"external,omitempty"`
}

// GetURL returns the external or internal URL depending on the image type.
func (i Audio) GetURL() string {
	if i.File != nil {
		return i.File.URL
	}
	if i.External != nil {
		return i.External.URL
	}
	return ""
}

type CodeBlock struct {
	BasicBlock
	Code Code `json:"code"`
}

type Code struct {
	RichText []RichText `json:"rich_text"`
	Caption  []RichText `json:"caption,omitempty"`
	Language string     `json:"language"`
}

type VideoBlock struct {
	BasicBlock
	Video Video `json:"video"`
}

type Video struct {
	Caption  []RichText  `json:"caption,omitempty"`
	Type     FileType    `json:"type"`
	File     *FileObject `json:"file,omitempty"`
	External *FileObject `json:"external,omitempty"`
}

type FileBlock struct {
	BasicBlock
	File BlockFile `json:"file"`
}

type BlockFile struct {
	Caption  []RichText  `json:"caption,omitempty"`
	Type     FileType    `json:"type"`
	File     *FileObject `json:"file,omitempty"`
	External *FileObject `json:"external,omitempty"`
}

type PdfBlock struct {
	BasicBlock
	Pdf Pdf `json:"pdf"`
}

type Pdf struct {
	Caption  []RichText  `json:"caption,omitempty"`
	Type     FileType    `json:"type,omitempty"`
	File     *FileObject `json:"file,omitempty"`
	External *FileObject `json:"external,omitempty"`
}

type BookmarkBlock struct {
	BasicBlock
	Bookmark Bookmark `json:"bookmark"`
}

type Bookmark struct {
	Caption []RichText `json:"caption,omitempty"`
	URL     string     `json:"url"`
}

type ChildDatabaseBlock struct {
	BasicBlock
	ChildDatabase struct {
		Title string `json:"title"`
	} `json:"child_database"`
}

type TableOfContentsBlock struct {
	BasicBlock
	TableOfContents TableOfContents `json:"table_of_contents"`
}

type TableOfContents struct {
	// empty
	Color string `json:"color,omitempty"`
}

type DividerBlock struct {
	BasicBlock
	Divider Divider `json:"divider"`
}

type Divider struct {
	// empty
}

type EquationBlock struct {
	BasicBlock
	Equation Equation `json:"equation"`
}

type Equation struct {
	Expression string `json:"expression"`
}

type BreadcrumbBlock struct {
	BasicBlock
	Breadcrumb Breadcrumb `json:"breadcrumb"`
}

type Breadcrumb struct {
	// empty
}

type ColumnBlock struct {
	BasicBlock
	Column Column `json:"column"`
}

type Column struct {
	// Children should at least have 1 block when appending.
	Children Blocks `json:"children"`
}

type ColumnListBlock struct {
	BasicBlock
	ColumnList ColumnList `json:"column_list"`
}

type ColumnList struct {
	// Children can only contain column blocks
	// Children should have at least 2 blocks when appending.
	Children Blocks `json:"children"`
}

// NOTE: will only be returned by the API. Cannot be created by the API.
// https://developers.notion.com/reference/block#link-preview-blocks
type LinkPreviewBlock struct {
	BasicBlock
	LinkPreview LinkPreview `json:"link_preview"`
}

type LinkPreview struct {
	URL string `json:"url"`
}

type LinkToPageBlock struct {
	BasicBlock
	LinkToPage LinkToPage `json:"link_to_page"`
}

type LinkToPage struct {
	Type       BlockType  `json:"type"`
	PageID     PageID     `json:"page_id,omitempty"`
	DatabaseID DatabaseID `json:"database_id,omitempty"`
}

type TemplateBlock struct {
	BasicBlock
	Template Template `json:"template"`
}

type Template struct {
	RichText []RichText `json:"rich_text"`
	Children Blocks     `json:"children,omitempty"`
}

type SyncedBlock struct {
	BasicBlock
	SyncedBlock Synced `json:"synced_block"`
}

type Synced struct {
	// SyncedFrom is nil for the original block.
	SyncedFrom *SyncedFrom `json:"synced_from"`
	Children   Blocks      `json:"children,omitempty"`
}

type SyncedFrom struct {
	BlockID BlockID `json:"block_id"`
}

type UnsupportedBlock struct {
	BasicBlock
}

type AppendBlockChildrenResponse struct {
	Object  ObjectType `json:"object"`
	Results []Block    `json:"results"`
}

type appendBlockResponse struct {
	Object  ObjectType               `json:"object"`
	Results []map[string]interface{} `json:"results"`
}

func (r *AppendBlockChildrenResponse) UnmarshalJSON(data []byte) error {
	var raw appendBlockResponse
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	blocks := make([]Block, 0)
	for _, b := range raw.Results {
		block, err := decodeBlock(b)
		if err != nil {
			return err
		}
		blocks = append(blocks, block)
	}

	*r = AppendBlockChildrenResponse{
		Object:  raw.Object,
		Results: blocks,
	}
	return nil
}

func decodeBlock(raw map[string]interface{}) (Block, error) {
	var b Block
	switch BlockType(raw["type"].(string)) {
	case BlockTypeParagraph:
		b = &ParagraphBlock{}
	case BlockTypeHeading1:
		b = &Heading1Block{}
	case BlockTypeHeading2:
		b = &Heading2Block{}
	case BlockTypeHeading3:
		b = &Heading3Block{}
	case BlockCallout:
		b = &CalloutBlock{}
	case BlockQuote:
		b = &QuoteBlock{}
	case BlockTypeBulletedListItem:
		b = &BulletedListItemBlock{}
	case BlockTypeNumberedListItem:
		b = &NumberedListItemBlock{}
	case BlockTypeToDo:
		b = &ToDoBlock{}
	case BlockTypeCode:
		b = &CodeBlock{}
	case BlockTypeToggle:
		b = &ToggleBlock{}
	case BlockTypeChildPage:
		b = &ChildPageBlock{}
	case BlockTypeEmbed:
		b = &EmbedBlock{}
	case BlockTypeImage:
		b = &ImageBlock{}
	case BlockTypeVideo:
		b = &VideoBlock{}
	case BlockTypeFile:
		b = &FileBlock{}
	case BlockTypePdf:
		b = &PdfBlock{}
	case BlockTypeBookmark:
		b = &BookmarkBlock{}
	case BlockTypeChildDatabase:
		b = &ChildDatabaseBlock{}
	case BlockTypeTableOfContents:
		b = &TableOfContentsBlock{}
	case BlockTypeDivider:
		b = &DividerBlock{}
	case BlockTypeEquation:
		b = &EquationBlock{}
	case BlockTypeBreadcrumb:
		b = &BreadcrumbBlock{}
	case BlockTypeColumn:
		b = &ColumnBlock{}
	case BlockTypeColumnList:
		b = &ColumnListBlock{}
	case BlockTypeLinkPreview:
		b = &LinkPreviewBlock{}
	case BlockTypeLinkToPage:
		b = &LinkToPageBlock{}
	case BlockTypeTemplate:
		b = &TemplateBlock{}
	case BlockTypeSyncedBlock:
		b = &SyncedBlock{}
	case BlockTypeTableBlock:
		b = &TableBlock{}
	case BlockTypeTableRowBlock:
		b = &TableRowBlock{}

	case BlockTypeUnsupported:
		b = &UnsupportedBlock{}
	default:
		return &UnsupportedBlock{}, nil
	}
	j, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(j, b)
	return b, err
}
