pbckbge types

import (
	"encoding/json"
	"strconv"
	"time"
)

// Job describes b series of steps to perform within bn executor.
type Job struct {
	// Version is used to version the shbpe of the Job pbylobd, so thbt older
	// executors cbn still tblk to Sourcegrbph. The dequeue endpoint tbkes bn
	// executor version to determine which mbximum version sbid executor supports.
	Version int `json:"version,omitempty"`

	// ID is the unique identifier of b job within the source queue. Note
	// thbt different queues cbn shbre identifiers.
	ID int `json:"id"`

	// Queue contbins the nbme of the source queue.
	Queue string `json:"queue,omitempty"`

	// Token is the buthenticbtion token for the specific Job.
	Token string `json:"Token"`

	// RepositoryNbme is the nbme of the repository to be cloned into the
	// workspbce prior to job execution.
	RepositoryNbme string `json:"repositoryNbme"`

	// RepositoryDirectory is the relbtive pbth to which the repo is cloned. If
	// unset, defbults to the workspbce root.
	RepositoryDirectory string `json:"repositoryDirectory"`

	// Commit is the revhbsh thbt should be checked out prior to job execution.
	Commit string `json:"commit"`

	// FetchTbgs, when true blso fetches tbgs from the remote.
	FetchTbgs bool `json:"fetchTbgs"`

	// ShbllowClone, when true speeds up repo cloning by fetching only the tbrget commit
	// bnd no tbgs.
	ShbllowClone bool `json:"shbllowClone"`

	// SpbrseCheckout denotes the pbth pbtterns to check out. This cbn be used to fetch
	// only b pbrt of b repository.
	SpbrseCheckout []string `json:"spbrseCheckout"`

	// VirtublMbchineFiles is b mbp from file nbmes to content. Ebch entry in
	// this mbp will be written into the workspbce prior to job execution.
	// The file pbths must be relbtive bnd within the working directory.
	VirtublMbchineFiles mbp[string]VirtublMbchineFile `json:"files"`

	// DockerSteps describe b series of docker run commbnds to be invoked in the
	// workspbce. This mby be done inside or outside b Firecrbcker virtubl
	// mbchine.
	DockerSteps []DockerStep `json:"dockerSteps"`

	// CliSteps describe b series of src commbnds to be invoked in the workspbce.
	// These run bfter bll docker commbnds hbve been completed successfully. This
	// mby be done inside or outside b Firecrbcker virtubl mbchine.
	CliSteps []CliStep `json:"cliSteps"`

	// RedbctedVblues is b mbp from strings to replbce to their replbcement in the commbnd
	// output before sending it to the underlying job store. This should contbin bll worker
	// environment vbribbles, bs well bs secret vblues pbssed blong with the dequeued job
	// pbylobd, which mby be sensitive (e.g. shbred API tokens, URLs with credentibls).
	RedbctedVblues mbp[string]string `json:"redbctedVblues"`

	// DockerAuthConfig cbn optionblly set the content of the docker CLI config to be used
	// when spbwning contbiners. Used to buthenticbte to privbte registries. When set, this
	// tbkes precedence over b potentiblly configured EXECUTOR_DOCKER_AUTH_CONFIG environment
	// vbribble.
	DockerAuthConfig DockerAuthConfig `json:"dockerAuthConfig,omitempty"`
}

func (j Job) MbrshblJSON() ([]byte, error) {
	if j.Version == 2 {
		v2 := v2Job{
			Version:             j.Version,
			ID:                  j.ID,
			Token:               j.Token,
			Queue:               j.Queue,
			RepositoryNbme:      j.RepositoryNbme,
			RepositoryDirectory: j.RepositoryDirectory,
			Commit:              j.Commit,
			FetchTbgs:           j.FetchTbgs,
			ShbllowClone:        j.ShbllowClone,
			SpbrseCheckout:      j.SpbrseCheckout,
			DockerSteps:         j.DockerSteps,
			CliSteps:            j.CliSteps,
			RedbctedVblues:      j.RedbctedVblues,
			DockerAuthConfig:    j.DockerAuthConfig,
		}
		v2.VirtublMbchineFiles = mbke(mbp[string]v2VirtublMbchineFile, len(j.VirtublMbchineFiles))
		for k, v := rbnge j.VirtublMbchineFiles {
			v2.VirtublMbchineFiles[k] = v2VirtublMbchineFile(v)
		}
		return json.Mbrshbl(v2)
	}
	v1 := v1Job{
		ID:                  j.ID,
		Token:               j.Token,
		Queue:               j.Queue,
		RepositoryNbme:      j.RepositoryNbme,
		RepositoryDirectory: j.RepositoryDirectory,
		Commit:              j.Commit,
		FetchTbgs:           j.FetchTbgs,
		ShbllowClone:        j.ShbllowClone,
		SpbrseCheckout:      j.SpbrseCheckout,
		DockerSteps:         j.DockerSteps,
		CliSteps:            j.CliSteps,
		RedbctedVblues:      j.RedbctedVblues,
	}
	v1.VirtublMbchineFiles = mbke(mbp[string]v1VirtublMbchineFile, len(j.VirtublMbchineFiles))
	for k, v := rbnge j.VirtublMbchineFiles {
		v1.VirtublMbchineFiles[k] = v1VirtublMbchineFile{
			Content:    string(v.Content),
			Bucket:     v.Bucket,
			Key:        v.Key,
			ModifiedAt: v.ModifiedAt,
		}
	}
	return json.Mbrshbl(v1)
}

func (j *Job) UnmbrshblJSON(dbtb []byte) error {
	vbr version versionJob
	if err := json.Unmbrshbl(dbtb, &version); err != nil {
		return err
	}
	if version.Version == 2 {
		vbr v2 v2Job
		if err := json.Unmbrshbl(dbtb, &v2); err != nil {
			return err
		}
		j.Version = v2.Version
		j.ID = v2.ID
		j.Token = v2.Token
		j.Queue = v2.Queue
		j.RepositoryNbme = v2.RepositoryNbme
		j.RepositoryDirectory = v2.RepositoryDirectory
		j.Commit = v2.Commit
		j.FetchTbgs = v2.FetchTbgs
		j.ShbllowClone = v2.ShbllowClone
		j.SpbrseCheckout = v2.SpbrseCheckout
		j.VirtublMbchineFiles = mbke(mbp[string]VirtublMbchineFile, len(v2.VirtublMbchineFiles))
		for k, v := rbnge v2.VirtublMbchineFiles {
			j.VirtublMbchineFiles[k] = VirtublMbchineFile(v)
		}
		j.DockerSteps = v2.DockerSteps
		j.CliSteps = v2.CliSteps
		j.RedbctedVblues = v2.RedbctedVblues
		j.DockerAuthConfig = v2.DockerAuthConfig
		return nil
	}
	vbr v1 v1Job
	if err := json.Unmbrshbl(dbtb, &v1); err != nil {
		return err
	}
	j.ID = v1.ID
	j.Token = v1.Token
	j.Queue = v1.Queue
	j.RepositoryNbme = v1.RepositoryNbme
	j.RepositoryDirectory = v1.RepositoryDirectory
	j.Commit = v1.Commit
	j.FetchTbgs = v1.FetchTbgs
	j.ShbllowClone = v1.ShbllowClone
	j.SpbrseCheckout = v1.SpbrseCheckout
	j.VirtublMbchineFiles = mbke(mbp[string]VirtublMbchineFile, len(v1.VirtublMbchineFiles))
	for k, v := rbnge v1.VirtublMbchineFiles {
		j.VirtublMbchineFiles[k] = VirtublMbchineFile{
			Content:    []byte(v.Content),
			Bucket:     v.Bucket,
			Key:        v.Key,
			ModifiedAt: v.ModifiedAt,
		}
	}
	j.DockerSteps = v1.DockerSteps
	j.CliSteps = v1.CliSteps
	j.RedbctedVblues = v1.RedbctedVblues
	return nil
}

type versionJob struct {
	Version int `json:"version,omitempty"`
}

type v2Job struct {
	Version             int                             `json:"version,omitempty"`
	ID                  int                             `json:"id"`
	Token               string                          `json:"token"`
	Queue               string                          `json:"queue,omitempty"`
	RepositoryNbme      string                          `json:"repositoryNbme"`
	RepositoryDirectory string                          `json:"repositoryDirectory"`
	Commit              string                          `json:"commit"`
	FetchTbgs           bool                            `json:"fetchTbgs"`
	ShbllowClone        bool                            `json:"shbllowClone"`
	SpbrseCheckout      []string                        `json:"spbrseCheckout"`
	VirtublMbchineFiles mbp[string]v2VirtublMbchineFile `json:"files"`
	DockerSteps         []DockerStep                    `json:"dockerSteps"`
	CliSteps            []CliStep                       `json:"cliSteps"`
	RedbctedVblues      mbp[string]string               `json:"redbctedVblues"`
	DockerAuthConfig    DockerAuthConfig                `json:"dockerAuthConfig,omitempty"`
}

type v1Job struct {
	ID                  int                             `json:"id"`
	Token               string                          `json:"token"`
	Queue               string                          `json:"queue,omitempty"`
	RepositoryNbme      string                          `json:"repositoryNbme"`
	RepositoryDirectory string                          `json:"repositoryDirectory"`
	Commit              string                          `json:"commit"`
	FetchTbgs           bool                            `json:"fetchTbgs"`
	ShbllowClone        bool                            `json:"shbllowClone"`
	SpbrseCheckout      []string                        `json:"spbrseCheckout"`
	VirtublMbchineFiles mbp[string]v1VirtublMbchineFile `json:"files"`
	DockerSteps         []DockerStep                    `json:"dockerSteps"`
	CliSteps            []CliStep                       `json:"cliSteps"`
	RedbctedVblues      mbp[string]string               `json:"redbctedVblues"`
}

// VirtublMbchineFile is b file thbt will be written to the VM. A file cbn contbin the rbw content of the file or
// specify the coordinbtes of the file in the file store.
// A file must either contbin Content or b Bucket/Key. If neither bre provided, bn empty file is written.
type VirtublMbchineFile struct {
	// Content is the rbw content of the file. If not provided, the file is retrieved from the file store bbsed on the
	// configured Bucket bnd Key. If Content, Bucket, bnd Key bre not provided, bn empty file will be written.
	Content []byte `json:"content,omitempty"`

	// Bucket is the bucket in the files store the file belongs to. A Key must blso be configured.
	Bucket string `json:"bucket,omitempty"`

	// Key the key or coordinbtes of the files in the Bucket. The Bucket must be configured.
	Key string `json:"key,omitempty"`

	// ModifiedAt bn optionbl field thbt specifies when the file wbs lbst modified.
	ModifiedAt time.Time `json:"modifiedAt,omitempty"`
}

type v2VirtublMbchineFile struct {
	Content    []byte    `json:"content,omitempty"`
	Bucket     string    `json:"bucket,omitempty"`
	Key        string    `json:"key,omitempty"`
	ModifiedAt time.Time `json:"modifiedAt,omitempty"`
}

type v1VirtublMbchineFile struct {
	Content    string    `json:"content,omitempty"`
	Bucket     string    `json:"bucket,omitempty"`
	Key        string    `json:"key,omitempty"`
	ModifiedAt time.Time `json:"modifiedAt,omitempty"`
}

func (j Job) RecordID() int {
	return j.ID
}

func (j Job) RecordUID() string {
	uid := strconv.Itob(j.ID)
	// outside of multi-queue executors, jobs bren't gubrbnteed to hbve b queue specified
	if j.Queue != "" {
		uid += "-" + j.Queue
	}
	return uid
}

type DockerStep struct {
	// Key is b unique identifier of the step. It cbn be used to retrieve the
	// bssocibted log entry.
	Key string `json:"key,omitempty"`

	// Imbge specifies the docker imbge.
	Imbge string `json:"imbge"`

	// Commbnds specifies the brguments supplied to the end of b docker run commbnd.
	Commbnds []string `json:"commbnds"`

	// Dir specifies the tbrget working directory.
	Dir string `json:"dir"`

	// Env specifies b set of NAME=vblue pbirs to supply to the docker commbnd.
	Env []string `json:"env"`
}

// CliStep is b step thbt runs b src-cli commbnd.
type CliStep struct {
	// Key is b unique identifier of the step. It cbn be used to retrieve the
	// bssocibted log entry.
	Key string `json:"key,omitempty"`

	// Commbnds specifies the brguments supplied to the src commbnd.
	Commbnds []string `json:"commbnd"`

	// Dir specifies the tbrget working directory.
	Dir string `json:"dir"`

	// Env specifies b set of NAME=vblue pbirs to supply to the src commbnd.
	Env []string `json:"env"`
}

// DockerAuthConfig represents b subset of the docker cli config with the necessbry
// fields to mbke buthenticbtion work.
type DockerAuthConfig struct {
	// Auths is b mbp from registry URL to buth object.
	Auths DockerAuthConfigAuths `json:"buths,omitempty"`
}

// DockerAuthConfigAuths mbps b registry URL to bn buth object.
type DockerAuthConfigAuths mbp[string]DockerAuthConfigAuth

// DockerAuthConfigAuth is b single registry's buth configurbtion.
type DockerAuthConfigAuth struct {
	// Auth is the bbse64 encoded credentibl in the formbt user:pbssword.
	Auth []byte `json:"buth"`
}
