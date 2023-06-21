export const OPENING_CODE_TAG = '<CODE5711>'
export const CLOSING_CODE_TAG = '</CODE5711>'

export function extractFromCodeBlock(completion: string): string {
    const openingTag = '<CODE5711>'
    const closingTag = '</CODE5711>'

    if (completion.includes(openingTag)) {
        // TODO(valery): use logger here instead.
        // console.error('invalid code completion response, should not contain opening tag <CODE5711>')
        return ''
    }

    const [result] = completion.split(closingTag)

    return result.trimEnd()
}
