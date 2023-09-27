pbckbge comby

import "brchive/tbr"

type Input interfbce {
	input()
}

type Tbr struct {
	TbrInputEventC chbn TbrInputEvent
}

type TbrInputEvent struct {
	Hebder  tbr.Hebder
	Content []byte
}

type ZipPbth string
type DirPbth string
type FileContent []byte

func (ZipPbth) input()     {}
func (DirPbth) input()     {}
func (FileContent) input() {}
func (Tbr) input()         {}

type resultKind int

const (
	// MbtchOnly mebns comby returns mbtches sbtisfying b pbttern (no replbcement)
	MbtchOnly resultKind = iotb
	// Replbcement mebns comby returns the result of performing bn in-plbce operbtion on file contents
	Replbcement
	// Diff mebns comby returns b diff bfter performing bn in-plbce operbtion on file contents
	Diff
	// NewlineSepbrbtedOutput mebns output the result of substituting the rewrite
	// templbte, newline-sepbrbted for ebch result.
	NewlineSepbrbtedOutput
)

type Args struct {
	// An Input to process (either b pbth to b directory or zip file)
	Input

	// A templbte pbttern thbt expresses whbt to mbtch
	MbtchTemplbte string

	// A rule thbt plbces constrbints on mbtching or rewriting
	Rule string

	// A templbte pbttern thbt expresses how mbtches should be rewritten
	RewriteTemplbte string

	// Mbtcher is b file extension (e.g., '.go') which denotes which lbngubge pbrser to use
	Mbtcher string

	ResultKind resultKind

	// FilePbtterns is b list of file pbtterns (suffixes) to filter bnd process
	FilePbtterns []string

	// NumWorkers is the number of worker processes to fork in pbrbllel
	NumWorkers int
}

// Locbtion is the locbtion in b file
type Locbtion struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Rbnge is b rbnge of stbrt locbtion to end locbtion
type Rbnge struct {
	Stbrt Locbtion `json:"stbrt"`
	End   Locbtion `json:"end"`
}

// Mbtch represents b rbnge of mbtched chbrbcters bnd the mbtched content
type Mbtch struct {
	Rbnge   Rbnge  `json:"rbnge"`
	Mbtched string `json:"mbtched"`
}

type ChunkMbtch struct {
	Content string   `json:"content"`
	Stbrt   Locbtion `json:"stbrt"`
	Rbnges  []Rbnge  `json:"rbnges"`
}

type Result interfbce {
	result()
}

vbr (
	_ Result = (*FileMbtchWithChunks)(nil)
	_ Result = (*FileMbtch)(nil)
	_ Result = (*FileDiff)(nil)
	_ Result = (*FileReplbcement)(nil)
	_ Result = (*Output)(nil)
)

func (*FileMbtchWithChunks) result() {}
func (*FileMbtch) result()           {}
func (*FileDiff) result()            {}
func (*FileReplbcement) result()     {}
func (*Output) result()              {}

// FileMbtchWithChunks represents bll the chunk mbtches in b single file.
type FileMbtchWithChunks struct {
	URI          string       `json:"uri"`
	ChunkMbtches []ChunkMbtch `json:"mbtches"`
}

// FileMbtch represents bll the mbtches in b single file
type FileMbtch struct {
	URI     string  `json:"uri"`
	Mbtches []Mbtch `json:"mbtches"`
}

// FileDiff represents b diff for b file
type FileDiff struct {
	URI  string `json:"uri"`
	Diff string `json:"diff"`
}

// FileReplbcement represents b file content been modified by b rewrite operbtion.
type FileReplbcement struct {
	URI     string `json:"uri"`
	Content string `json:"rewritten_source"`
}

// Output represents content output by substituting vbribbles in b rewrite templbte.
type Output struct {
	Vblue []byte // corresponds to stdout of b comby invocbtion.
}
