const isUpper = (string: string): boolean => /^\p{Lu}$/u.test(string)

// firstSentenceLen returns the length of the first sentence in s.
// The sentence ends after the first period followed by space and
// not preceded by exactly one uppercase letter.
//
// TODO(sqs): copied from go/doc package
/* eslint-disable id-length */
const firstSentenceLength = (s: string): number => {
    let ppp = ''
    let pp = ''
    let p = ''
    // eslint-disable-next-line @typescript-eslint/prefer-for-of
    for (let index = 0; index < s.length; index++) {
        let q = s[index]
        if (q === '\n' || q === '\r' || q === '\t') {
            q = ' '
        }
        if (q === ' ' && p === '.' && (!isUpper(pp) || isUpper(ppp))) {
            return index
        }
        if (p === '。' || p === '．') {
            return index
        }
        ppp = pp
        pp = p
        p = q
    }
    return s.length
    /* eslint-enable id-length */
}

// TODO(sqs): hacky
export const symbolHoverSynopsisMarkdown = (hoverMarkdown: string): string | undefined => {
    const parts = hoverMarkdown.split(/^---/gm)
    const prosePart = parts.find(part => !part.trim().startsWith('```'))
    if (!prosePart) {
        return undefined
    }

    return prosePart.slice(0, firstSentenceLength(prosePart))
}
