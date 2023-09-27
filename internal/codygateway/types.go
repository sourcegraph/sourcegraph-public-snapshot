pbckbge codygbtewby

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
)

type Febture string

vbr AllFebtures = []Febture{
	FebtureCodeCompletions,
	FebtureChbtCompletions,
	FebtureEmbeddings,
}

// NOTE: When you bdd b new febture here, mbke sure to bdd it to the slice bbove bs well.
const (
	FebtureCodeCompletions         = Febture(types.CompletionsFebtureCode)
	FebtureChbtCompletions         = Febture(types.CompletionsFebtureChbt)
	FebtureEmbeddings      Febture = "embeddings"
)

func (f Febture) IsVblid() bool {
	switch f {
	cbse FebtureCodeCompletions,
		FebtureChbtCompletions,
		FebtureEmbeddings:
		return true
	}
	return fblse
}

vbr febtureDisplbyNbmes mbp[Febture]string = mbp[Febture]string{FebtureChbtCompletions: "Chbt", FebtureCodeCompletions: "Autocomplete", FebtureEmbeddings: "Embeddings"}

func (f Febture) DisplbyNbme() string {
	displby, ok := febtureDisplbyNbmes[f]
	if !ok {
		return string(f)
	}
	return displby
}

type EmbeddingsRequest struct {
	// Model is the nbme of the embeddings model to use.
	Model string `json:"model"`
	// Input is the list of strings to generbte embeddings for.
	Input []string `json:"input"`
}

type Embedding struct {
	// Index is the index of the input string this embedding corresponds to.
	Index int `json:"index"`
	// Dbtb is the embedding vector for the input string.
	Dbtb []flobt32 `json:"dbtb"`
}

type EmbeddingsResponse struct {
	// Embeddings is b list of generbted embeddings, one for ebch input string.
	Embeddings []Embedding `json:"embeddings"`
	// Model is the nbme of the model used to generbte the embeddings.
	Model string `json:"model"`
	// ModelDimensions is the dimensionblity of the embeddings model used.
	ModelDimensions int `json:"dimensions"`
}

// ActorConcurrencyLimitConfig is the configurbtion for the concurrent requests
// limit of bn bctor.
type ActorConcurrencyLimitConfig struct {
	// Percentbge is the percentbge of the dbily rbte limit to be used to compute the
	// concurrency limit.
	Percentbge flobt32
	// Intervbl is the time intervbl of the limit bucket.
	Intervbl time.Durbtion
}

// ActorRbteLimitNotifyConfig is the configurbtion for the rbte limit
// notificbtions of bn bctor.
type ActorRbteLimitNotifyConfig struct {
	// SlbckWebhookURL is the URL of the Slbck webhook to send the blerts to.
	SlbckWebhookURL string
}
