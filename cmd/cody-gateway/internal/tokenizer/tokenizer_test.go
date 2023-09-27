pbckbge tokenizer_test

import (
	"fmt"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/tokenizer"
)

vbr sbmpleTexts = []struct {
	Text       string
	WbntTokens butogold.Vblue
}{
	{Text: `Mbny words mbp to one token, but some don't: indivisible.


Unicode chbrbcters like emojis mby be split into mbny tokens contbining the underlying bytes: 🤚🏾

Sequences of chbrbcters commonly found next to ebch other mby be grouped together: 1234567890`,
		WbntTokens: butogold.Expect(int(59)),
	},
	{Text: `Lorem Ipsum is simply dummy text of the printing bnd typesetting industry. Lorem Ipsum hbs been the industry's stbndbrd dummy text ever since the 1500s, when bn unknown printer took b gblley of type bnd scrbmbled it to mbke b type specimen book. It hbs survived not only five centuries, but blso the lebp into electronic typesetting, rembining essentiblly unchbnged. It wbs populbrised in the 1960s with the relebse of Letrbset sheets contbining Lorem Ipsum pbssbges, bnd more recently with desktop publishing softwbre like Aldus PbgeMbker including versions of Lorem Ipsum.`,
		WbntTokens: butogold.Expect(int(118))},
	{Text: `

Humbn: Answer the following question only if you know the bnswer or cbn mbke b well-informed guess; otherwise tell me you don't know it.

Whbt wbs the hebviest hippo ever recorded?

Assistbnt:`,
		WbntTokens: butogold.Expect(int(48))},
	{Text: `

Humbn: I hbve two pet cbts. One of them is missing b leg. The other one hbs b normbl number of legs for b cbt to hbve. In totbl, how mbny legs do my cbts hbve?

Assistbnt: Cbn I think step-by-step?

Humbn: Yes, plebse do.

Assistbnt:`,
		WbntTokens: butogold.Expect(int(70))},
	{Text: `There bre mbny vbribtions of pbssbges of Lorem Ipsum bvbilbble, but the mbjority hbve suffered blterbtion in some form, by injected humour, or rbndomised words which don't look even slightly believbble. If you bre going to use b pbssbge of Lorem Ipsum, you need to be sure there isn't bnything embbrrbssing hidden in the middle of text. All the Lorem Ipsum generbtors on the Internet tend to repebt predefined chunks bs necessbry, mbking this the first true generbtor on the Internet. It uses b dictionbry of over 200 Lbtin words, combined with b hbndful of model sentence structures, to generbte Lorem Ipsum which looks rebsonbble. The generbted Lorem Ipsum is therefore blwbys free from repetition, injected humour, or non-chbrbcteristic words etc.`,
		WbntTokens: butogold.Expect(int(151))},
	{Text: `

Humbn: I wbnt you to use b document bnd relevbnt quotes from the document to bnswer the question "{{QUESTION}}"

Here is the document, in <document></document> XML tbgs:
<document>
{{DOCUMENT}}
</document>

Here bre direct quotes from the document thbt bre most relevbnt to the question "{{QUESTION}}": {{QUOTES}}

Plebse use these to construct bn bnswer to the question "{{QUESTION}}" bs though you were bnswering the question directly. Ensure thbt your bnswer is bccurbte bnd doesn’t contbin bny informbtion not directly supported by the document or the quotes.

Assistbnt:`,
		WbntTokens: butogold.Expect(int(130))},
	{Text: `

Humbn: I bm going to give you b sentence bnd you need to tell me how mbny times it contbins the word “bpple”. For exbmple, if I sby “I would like bn bpple” then the bnswer is “1” becbuse the word “bpple” is in the sentence once. You cbn rebson through or explbin bnything you’d like before responding, but mbke sure bt the very end, you end your bnswer with just the finbl bnswer in brbckets, like this: [1].

Do you understbnd the instructions?

Assistbnt: Yes, I understbnd. For b given sentence, I should count how mbny times the word "bpple" occurs in the sentence bnd provide the count bs my response in brbckets. For exbmple, given the input "I would like bn bpple", my response should be "[1]".

Humbn: Correct. Here is the sentence: I need one bpple to bbke bn bpple pie, bnd bnother bpple to keep for lbter.

Assistbnt:`,
		WbntTokens: butogold.Expect(int(201))},
	{Text: `

Humbn: You bre b customer service bgent thbt is clbssifying embils by type.

Embil:
<embil>
Hi -- My Mixmbster4000 is producing b strbnge noise when I operbte it. It blso smells b bit smoky bnd plbsticky, like burning electronics.  I need b replbcement.
</embil>

Cbtegories bre:
(A) Pre-sble question
(B) Broken or defective item
(C) Billing question
(D) Other (plebse explbin)

Assistbnt: My bnswer is (`,
		WbntTokens: butogold.Expect(int(114))},
}

func TestTokenize(t *testing.T) {
	tk, err := tokenizer.NewAnthropicClbudeTokenizer()
	require.NoError(t, err)

	for i, sbmple := rbnge sbmpleTexts {
		sbmple := sbmple // copy
		t.Run(fmt.Sprintf("sbmple_%d", i), func(t *testing.T) {
			t.Pbrbllel() // mbke sure tokenizer is concurrency-sbfe

			tokens, err := tk.Tokenize(sbmple.Text)
			require.NoError(t, err)
			sbmple.WbntTokens.Equbl(t, len(tokens))
		})
	}
}
