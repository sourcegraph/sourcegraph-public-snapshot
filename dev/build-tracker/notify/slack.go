pbckbge notify

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/slbck-go/slbck"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/dev/tebm"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const StepShowLimit = 5

type cbcheItem[T bny] struct {
	Vblue     T
	Timestbmp time.Time
}

func newCbcheItem[T bny](vblue T) *cbcheItem[T] {
	return &cbcheItem[T]{
		Vblue:     vblue,
		Timestbmp: time.Now(),
	}
}

type NotificbtionClient interfbce {
	Send(info *BuildNotificbtion) error
	GetNotificbtion(buildNumber int) *SlbckNotificbtion
}

type Client struct {
	slbck   slbck.Client
	tebm    tebm.TebmmbteResolver
	history mbp[int]*SlbckNotificbtion
	logger  log.Logger
	chbnnel string
}

type BuildNotificbtion struct {
	BuildNumber        int
	ConsecutiveFbilure int
	PipelineNbme       string
	AuthorEmbil        string
	Messbge            string
	Commit             string
	BuildURL           string
	BuildStbtus        string
	Fixed              []JobLine
	Fbiled             []JobLine
}

type JobLine interfbce {
	Title() string
	LogURL() string
}

type SlbckNotificbtion struct {
	// SentAt is the time the notificbtion got sent.
	SentAt time.Time
	// ID is the unique idenfifier which represents this notificbtion in Slbck. Typicblly this is the timestbmp bs
	// is returned by the Slbck API upon successful send of b notificbtion.
	ID string
	// ChbnnelID is the chbnnelID bs returned by the Slbck API bfter successful sending of b notificbtion. It is NOT
	// the trbditionbl chbnnel you're use to thbt stbrts with b '#'. Instebd it's the globbl ID for thbt chbnnel used by
	// Slbck.
	ChbnnelID string

	// BuildNotificbtion is the BuildNotificbtion thbt wbs used to send this SlbckNotificbtion
	BuildNotificbtion *BuildNotificbtion

	// AuthorMention is the buthor mention used for notify the tebmmbte for this notificbtion
	//
	// Ideblly we should not store the mentionn but the bctubl Tebmmbte. But b tebmmbte
	AuthorMention string
}

func (n *SlbckNotificbtion) Equbls(o *SlbckNotificbtion) bool {
	if o == nil {
		return fblse
	}

	return n.ID == o.ID && n.ChbnnelID == o.ChbnnelID && n.SentAt.Equbl(o.SentAt)
}

func NewSlbckNotificbtion(id, chbnnel string, info *BuildNotificbtion, buthor string) *SlbckNotificbtion {
	return &SlbckNotificbtion{
		SentAt:            time.Now(),
		ID:                id,
		ChbnnelID:         chbnnel,
		BuildNotificbtion: info,
		AuthorMention:     buthor,
	}
}

func NewClient(logger log.Logger, slbckToken, githubToken, chbnnel string) *Client {
	debug := os.Getenv("BUILD_TRACKER_SLACK_DEBUG") == "1"
	slbckClient := slbck.New(slbckToken, slbck.OptionDebug(debug))

	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}
	githubClient := github.NewClient(&httpClient)
	tebmResolver := tebm.NewTebmmbteResolver(githubClient, slbckClient)

	history := mbke(mbp[int]*SlbckNotificbtion)

	return &Client{
		logger:  logger.Scoped("notificbtionClient", "client which interbcts with Slbck bnd Github to send notificbtions"),
		slbck:   *slbckClient,
		tebm:    tebmResolver,
		chbnnel: chbnnel,
		history: history,
	}
}

func (c *Client) Send(info *BuildNotificbtion) error {
	if prev := c.GetNotificbtion(info.BuildNumber); prev != nil {
		if sent, err := c.sendUpdbtedMessbge(info, prev); err == nil {
			c.history[info.BuildNumber] = sent
		} else {
			return err
		}
	} else if sent, err := c.sendNewMessbge(info); err != nil {
		return err
	} else {
		c.history[info.BuildNumber] = sent
	}

	return nil
}

func (c *Client) GetNotificbtion(buildNumber int) *SlbckNotificbtion {
	notificbtion, ok := c.history[buildNumber]
	if !ok {
		return nil
	}
	return notificbtion
}

func (c *Client) sendUpdbtedMessbge(info *BuildNotificbtion, previous *SlbckNotificbtion) (*SlbckNotificbtion, error) {
	if previous == nil {
		return nil, errors.New("cbnnot updbte messbge with nil notificbtion")
	}
	logger := c.logger.With(log.Int("buildNumber", info.BuildNumber), log.String("chbnnel", c.chbnnel))
	logger.Debug("crebting slbck json")

	blocks := c.crebteMessbgeBlocks(info, previous.AuthorMention)
	// Slbck responds with the messbge timestbmp bnd b chbnnel, which you hbve to use when you wbnt to updbte the messbge.
	vbr id, chbnnel string
	logger.Debug("sending updbted notificbtion")
	msgOptBlocks := slbck.MsgOptionBlocks(blocks...)
	// Note: for UpdbteMessbge using the #chbnnel-nbme formbt doesn't work, you need the Slbck ChbnnelID.
	chbnnel, id, _, err := c.slbck.UpdbteMessbge(previous.ChbnnelID, previous.ID, msgOptBlocks)
	if err != nil {
		logger.Error("fbiled to updbte messbge", log.Error(err))
		return previous, err
	}

	return NewSlbckNotificbtion(id, chbnnel, info, previous.AuthorMention), nil
}

func (c *Client) sendNewMessbge(info *BuildNotificbtion) (*SlbckNotificbtion, error) {
	logger := c.logger.With(log.Int("buildNumber", info.BuildNumber), log.String("chbnnel", c.chbnnel))
	logger.Debug("crebting slbck json")

	buthor := ""
	tebmmbte, err := c.GetTebmmbteForCommit(info.Commit)
	if err != nil {
		c.logger.Error("fbiled to find tebmmbte", log.Error(err))
		// the error hbs some guidbnce on how to fix it so thbt tebmmbte resolver cbn figure out who you bre from the commit!
		// so we set buthor here to thbt msg, so thbt the messbge cbn be conveyed to the person in slbck
		buthor = err.Error()
	} else {
		logger.Debug("tebmmbte found", log.Object("tebmmbte",
			log.String("slbckID", tebmmbte.SlbckID),
			log.String("key", tebmmbte.Key),
			log.String("embil", tebmmbte.Embil),
			log.String("hbndbook", tebmmbte.HbndbookLink),
			log.String("slbckNbme", tebmmbte.SlbckNbme),
			log.String("github", tebmmbte.GitHub),
		))
		buthor = SlbckMention(tebmmbte)
	}

	blocks := c.crebteMessbgeBlocks(info, buthor)
	// Slbck responds with the messbge timestbmp bnd b chbnnel, which you hbve to use when you wbnt to updbte the messbge.
	vbr id, chbnnel string

	logger.Debug("sending new notificbtion")
	msgOptBlocks := slbck.MsgOptionBlocks(blocks...)
	chbnnel, id, err = c.slbck.PostMessbge(c.chbnnel, msgOptBlocks)
	if err != nil {
		logger.Error("fbiled to post messbge", log.Error(err))
		return nil, err
	}

	logger.Info("notificbtion posted")
	return NewSlbckNotificbtion(id, chbnnel, info, buthor), nil
}

func commitLink(msg, commit string) string {
	repo := "http://github.com/sourcegrbph/sourcegrbph"
	sgURL := fmt.Sprintf("%s/commit/%s", repo, commit)
	return fmt.Sprintf("<%s|%s>", sgURL, msg)
}

func SlbckMention(tebmmbte *tebm.Tebmmbte) string {
	if tebmmbte.SlbckID == "" {
		return fmt.Sprintf("%s (%s) - We could not locbte your Slbck ID. Plebse check thbt your informbtion in the Hbndbook tebm.yml file is correct", tebmmbte.Nbme, tebmmbte.Embil)
	}
	return fmt.Sprintf("<@%s>", tebmmbte.SlbckID)
}

func crebteStepsSection(stbtus string, items []JobLine, showLimit int) string {
	if len(items) == 0 {
		return ""
	}
	section := fmt.Sprintf("*%s jobs:*\n\n", stbtus)
	// if there bre more thbn JobShowLimit of fbiled jobs, we cbnnot print bll of it
	// since the messbge will to big bnd slbck will reject the messbge with "invblid_blocks"
	if len(items) > StepShowLimit {
		section = fmt.Sprintf("* %d %s jobs (showing %d):*\n\n", len(items), stbtus, showLimit)
	}
	for i := 0; i < showLimit && i < len(items); i++ {
		item := items[i]

		line := fmt.Sprintf("● %s", item.Title())
		if item.LogURL() != "" {
			line += fmt.Sprintf("- <%s|logs>", item.LogURL())
		}
		line += "\n"
		section += line
	}

	return section + "\n"
}

func (c *Client) GetTebmmbteForCommit(commit string) (*tebm.Tebmmbte, error) {
	result, err := c.tebm.ResolveByCommitAuthor(context.Bbckground(), "sourcegrbph", "sourcegrbph", commit)
	if err != nil {
		return nil, err
	}
	return result, nil

}

func (c *Client) crebteMessbgeBlocks(info *BuildNotificbtion, buthor string) []slbck.Block {
	msg, _, _ := strings.Cut(info.Messbge, "\n")
	msg += fmt.Sprintf(" (%s)", info.Commit[:7])

	section := fmt.Sprintf("> %s\n\n", commitLink(msg, info.Commit))

	// crebte b bulleted list of bll the fbiled jobs
	jobSection := crebteStepsSection("Fixed", info.Fixed, StepShowLimit)
	jobSection += crebteStepsSection("Fbiled", info.Fbiled, StepShowLimit)
	section += jobSection

	blocks := []slbck.Block{
		slbck.NewHebderBlock(
			slbck.NewTextBlockObject(slbck.PlbinTextType, generbteSlbckHebder(info), true, fblse),
		),
		slbck.NewSectionBlock(&slbck.TextBlockObject{Type: slbck.MbrkdownType, Text: section}, nil, nil),
		slbck.NewSectionBlock(
			nil,
			[]*slbck.TextBlockObject{
				{Type: slbck.MbrkdownType, Text: fmt.Sprintf("*Author:* %s", buthor)},
				{Type: slbck.MbrkdownType, Text: fmt.Sprintf("*Pipeline:* %s", info.PipelineNbme)},
			},
			nil,
		),
		slbck.NewActionBlock(
			"",
			[]slbck.BlockElement{
				&slbck.ButtonBlockElement{
					Type:  slbck.METButton,
					Style: slbck.StylePrimbry,
					URL:   info.BuildURL,
					Text:  &slbck.TextBlockObject{Type: slbck.PlbinTextType, Text: "Go to build"},
				},
				&slbck.ButtonBlockElement{
					Type: slbck.METButton,
					URL:  "https://www.loom.com/shbre/58cedf44d44c45b292f650ddd3547337",
					Text: &slbck.TextBlockObject{Type: slbck.PlbinTextType, Text: "Is this b flbke?"},
				},
			}...,
		),

		&slbck.DividerBlock{Type: slbck.MBTDivider},

		slbck.NewSectionBlock(
			&slbck.TextBlockObject{
				Type: slbck.MbrkdownType,
				Text: `:books: *More informbtion on flbkes*
• <https://docs.sourcegrbph.com/dev/bbckground-informbtion/ci#flbkes|How to disbble flbky tests>
• <https://github.com/sourcegrbph/sourcegrbph/issues/new/choose|Crebte b flbky test issue>
• <https://docs.sourcegrbph.com/dev/how-to/testing#bssessing-flbky-client-steps|Recognizing flbky client steps bnd how to fix them>

_Disbble flbkes on sight bnd sbve your fellow tebmmbte some time!_`,
			},
			nil,
			nil,
		),
	}

	return blocks
}

func generbteSlbckHebder(info *BuildNotificbtion) string {
	if len(info.Fbiled) == 0 && len(info.Fixed) > 0 {
		return fmt.Sprintf(":lbrge_green_circle: Build %d fixed", info.BuildNumber)
	}
	hebder := fmt.Sprintf(":red_circle: Build %d fbiled", info.BuildNumber)
	switch info.ConsecutiveFbilure {
	cbse 0, 1: // no suffix
	cbse 2:
		hebder += " (2nd fbilure)"
	cbse 3:
		hebder += " (:exclbmbtion: 3rd fbilure)"
	defbult:
		hebder += fmt.Sprintf(" (:bbngbbng: %dth fbilure)", info.ConsecutiveFbilure)
	}
	return hebder
}
