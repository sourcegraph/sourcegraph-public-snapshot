import { Position } from '@sourcegraph/extension-api-classes'
import { OFFSET_TO_POSITION_TESTS, POSITION_TO_OFFSET_TESTS } from '../../client/types/textDocument.test'
import { ExtensionDocument, getEOL } from './textDocument'

describe('ExtensionDocument', () => {
    const textDocument = (text = 't'): ExtensionDocument => new ExtensionDocument({ uri: 'u', languageId: 'l', text })

    test('uri', () => expect(textDocument().uri).toBe('u'))
    test('languageId', () => expect(textDocument().languageId).toBe('l'))
    test('text', () => expect(textDocument().text).toBe('t'))

    describe('positionAt', () => {
        for (const [index, { text, ...testCase }] of OFFSET_TO_POSITION_TESTS.entries()) {
            test(index.toString(), () =>
                expect(textDocument(text).positionAt(testCase.offset)).toMatchObject(testCase.pos)
            )
        }
    })

    describe('offsetAt', () => {
        for (const [index, { text, ...testCase }] of POSITION_TO_OFFSET_TESTS.entries()) {
            test(index.toString(), () =>
                expect(textDocument(text).offsetAt(new Position(testCase.pos.line, testCase.pos.character))).toEqual(
                    testCase.offset
                )
            )
        }
    })

    test('getWordRangeAtPosition', () =>
        expect(textDocument('aa bb').getWordRangeAtPosition(new Position(0, 3))).toMatchObject({
            start: { line: 0, character: 3 },
            end: { line: 0, character: 5 },
        }))
})

describe('getEOL', () => {
    test('\\n', () => expect(getEOL('a\nb')).toBe('\n'))
    test('\\r\\n', () => expect(getEOL('a\r\nb')).toBe('\r\n'))
    test('\\r', () => expect(getEOL('a\rb')).toBe('\r'))
})
