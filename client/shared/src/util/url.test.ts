import { isExternalLink, parseRepoURI } from '@sourcegraph/common'

import { parseHash, withWorkspaceRootInputRevision } from './url'

describe('util/url', () => {
    const linePosition = { line: 1 }
    const lineCharPosition = { line: 1, character: 1 }
    const referenceMode = { ...lineCharPosition, viewState: 'references' }

    describe('parseHash', () => {
        test('parses empty hash', () => {
            expect(parseHash('')).toEqual({})
        })

        test('parses unexpectedly formatted hash', () => {
            expect(parseHash('L-53')).toEqual({})
            expect(parseHash('L53:')).toEqual({})
            expect(parseHash('L1:2-')).toEqual({})
            expect(parseHash('L1:2-3')).toEqual({})
            expect(parseHash('L1:2-3:')).toEqual({})
            expect(parseHash('L1:-3:')).toEqual({})
            expect(parseHash('L1:-3:4')).toEqual({})
            expect(parseHash('L1-2:3')).toEqual({})
            expect(parseHash('L1-2:')).toEqual({})
            expect(parseHash('L1:-2')).toEqual({})
            expect(parseHash('L1:2--3:4')).toEqual({})
            expect(parseHash('L53:a')).toEqual({})
        })

        test('parses hash with leading octothorpe', () => {
            expect(parseHash('#L1')).toEqual(linePosition)
        })

        test('parses hash with line', () => {
            expect(parseHash('L1')).toEqual(linePosition)
        })

        test('parses hash with line and character', () => {
            expect(parseHash('L1:1')).toEqual(lineCharPosition)
        })

        test('parses hash with range', () => {
            expect(parseHash('L1-2')).toEqual({ line: 1, endLine: 2 })
            expect(parseHash('L1:2-3:4')).toEqual({ line: 1, character: 2, endLine: 3, endCharacter: 4 })
            expect(parseHash('L47-L55')).toEqual({ line: 47, endLine: 55 })
            expect(parseHash('L34:2-L38:3')).toEqual({ line: 34, character: 2, endLine: 38, endCharacter: 3 })
        })

        test('parses hash with references', () => {
            expect(parseHash('$references')).toEqual({ viewState: 'references' })
            expect(parseHash('L1:1$references')).toEqual(referenceMode)
            expect(parseHash('L1:1$references')).toEqual(referenceMode)
        })
        test('parses modern hash with references', () => {
            expect(parseHash('tab=references')).toEqual({ viewState: 'references' })
            expect(parseHash('L1:1&tab=references')).toEqual(referenceMode)
            expect(parseHash('L1:1&tab=references')).toEqual(referenceMode)
        })
    })
})

describe('withWorkspaceRootInputRevision', () => {
    test('uses input revision for URI inside root with input revision', () =>
        expect(
            withWorkspaceRootInputRevision([{ uri: 'git://r?c', inputRevision: 'v' }], parseRepoURI('git://r?c#f'))
        ).toEqual(parseRepoURI('git://r?v#f')))

    test('does not change URI outside root (different repoName)', () =>
        expect(
            withWorkspaceRootInputRevision([{ uri: 'git://r?c', inputRevision: 'v' }], parseRepoURI('git://r2?c#f'))
        ).toEqual(parseRepoURI('git://r2?c#f')))

    test('does not change URI outside root (different revision)', () =>
        expect(
            withWorkspaceRootInputRevision([{ uri: 'git://r?c', inputRevision: 'v' }], parseRepoURI('git://r?c2#f'))
        ).toEqual(parseRepoURI('git://r?c2#f')))

    test('uses empty string input revision (treats differently from undefined)', () =>
        expect(
            withWorkspaceRootInputRevision([{ uri: 'git://r?c', inputRevision: '' }], parseRepoURI('git://r?c#f'))
        ).toEqual({ ...parseRepoURI('git://r?c#f'), revision: '' }))

    test('does not change URI if root has undefined input revision', () =>
        expect(
            withWorkspaceRootInputRevision(
                [{ uri: 'git://r?c', inputRevision: undefined }],
                parseRepoURI('git://r?c#f')
            )
        ).toEqual(parseRepoURI('git://r?c#f')))
})

describe('isExternalLink', () => {
    it('returns false for the same site', () => {
        jsdom.reconfigure({ url: 'https://github.com/here' })
        expect(isExternalLink('https://github.com/there')).toBe(false)
    })
    it('returns false for relative links', () => {
        jsdom.reconfigure({ url: 'https://github.com/here' })
        expect(isExternalLink('/there')).toBe(false)
    })
    it('returns false for invalid URLs', () => {
        jsdom.reconfigure({ url: 'https://github.com/here' })

        expect(isExternalLink(' ')).toBe(false)
    })
    it('returns true for a different site', () => {
        jsdom.reconfigure({ url: 'https://github.com/here' })
        expect(isExternalLink('https://sourcegraph.com/here')).toBe(true)
    })
})
