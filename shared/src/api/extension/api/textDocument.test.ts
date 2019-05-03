import { Position } from '@sourcegraph/extension-api-classes'
import { OFFSET_TO_POSITION_TESTS, POSITION_TO_OFFSET_TESTS } from '../../client/types/textDocument.test'
import { ExtDocument, getEOL } from './textDocument'

describe('ExtDocument', () => {
    const doc = (text = 't') => new ExtDocument({ uri: 'u', languageId: 'l', text })

    test('uri', () => expect(doc().uri).toBe('u'))
    test('languageId', () => expect(doc().languageId).toBe('l'))
    test('text', () => expect(doc().text).toBe('t'))

    describe('positionAt', () => {
        for (const [i, { text, ...c }] of OFFSET_TO_POSITION_TESTS.entries()) {
            test(i.toString(), () => expect(doc(text).positionAt(c.offset)).toMatchObject(c.pos))
        }
    })

    describe('offsetAt', () => {
        for (const [i, { text, ...c }] of POSITION_TO_OFFSET_TESTS.entries()) {
            test(i.toString(), () =>
                expect(doc(text).offsetAt(new Position(c.pos.line, c.pos.character))).toEqual(c.offset)
            )
        }
    })

    test('getWordRangeAtPosition', () =>
        expect(doc('aa bb').getWordRangeAtPosition(new Position(0, 3))).toMatchObject({
            start: { line: 0, character: 3 },
            end: { line: 0, character: 5 },
        }))
})

describe('getEOL', () => {
    test('\\n', () => expect(getEOL('a\nb')).toBe('\n'))
    test('\\r\\n', () => expect(getEOL('a\r\nb')).toBe('\r\n'))
    test('\\r', () => expect(getEOL('a\rb')).toBe('\r'))
})
