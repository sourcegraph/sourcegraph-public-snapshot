import { OPENING_CODE_TAG, CLOSING_CODE_TAG, extractFromCodeBlock } from './text-processing'

describe('extractFromCodeBlock', () => {
    it('extracts value from code completion XML tags', () => {
        expect(extractFromCodeBlock(`hello world${CLOSING_CODE_TAG}`)).toBe('hello world')
        expect(extractFromCodeBlock(`<randomTag>hello world</randomTag>${CLOSING_CODE_TAG}`)).toBe(
            '<randomTag>hello world</randomTag>'
        )
        expect(extractFromCodeBlock(`const isEnabled = true${CLOSING_CODE_TAG}something else`)).toBe(
            'const isEnabled = true'
        )
    })

    it('returns the whole string if the closing tag is not found', () => {
        expect(extractFromCodeBlock('hello world')).toBe('hello world')
        expect(extractFromCodeBlock('<randomTag>hello world</randomTag>')).toBe('<randomTag>hello world</randomTag>')
        expect(extractFromCodeBlock('const isEnabled = true // something else')).toBe(
            'const isEnabled = true // something else'
        )
    })

    it('returns an empty string if the opening tag is found', () => {
        expect(extractFromCodeBlock(`${OPENING_CODE_TAG}hello world${CLOSING_CODE_TAG}`)).toBe('')
        expect(extractFromCodeBlock(`hello world${OPENING_CODE_TAG}`)).toBe('')
        expect(extractFromCodeBlock(OPENING_CODE_TAG)).toBe('')
    })
})
