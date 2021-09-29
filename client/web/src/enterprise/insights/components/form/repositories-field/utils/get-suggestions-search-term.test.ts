import { getSuggestionsSearchTerm } from './get-suggestions-search-term'

describe('get-suggestions-search-term', () => {
    it('should return null term with null caret position', () => {
        expect(getSuggestionsSearchTerm({ value: '', caretPosition: null })).toStrictEqual({
            repositories: [],
            value: null,
            index: null,
        })
    })

    describe('should return correct term ', () => {
        it('with one repository in repositories input string', () => {
            // Caret position
            // ---------------------- ↓ -----------------
            const stringValue = 'github.com/example/about'

            expect(getSuggestionsSearchTerm({ value: stringValue, caretPosition: 5 })).toStrictEqual({
                repositories: ['github.com/example/about'],
                value: 'github.com/example/about',
                index: 0,
            })
        })

        it('with two repositories in repositories input string', () => {
            // Caret position
            // ------------------------------------------- ↓ -----------------
            const stringValue = 'github.com/example/about, github.com/example/another-repo'

            expect(getSuggestionsSearchTerm({ value: stringValue, caretPosition: 26 })).toStrictEqual({
                repositories: ['github.com/example/about', 'github.com/example/another-repo'],
                value: 'github.com/example/another-repo',
                index: 1,
            })
        })

        it('with caret at the end of the input string', () => {
            // Caret position
            // ------------------------------------------------------------------------- ↓ -
            const stringValue = 'github.com/example/about, github.com/example/another-repo'

            expect(getSuggestionsSearchTerm({ value: stringValue, caretPosition: stringValue.length })).toStrictEqual({
                repositories: ['github.com/example/about', 'github.com/example/another-repo'],
                value: 'github.com/example/another-repo',
                index: 1,
            })
        })

        it('with caret at the end of the input string with comma', () => {
            // Caret position
            // -------------------------------------------------------------------------- ↓ -
            const stringValue = 'github.com/example/about, github.com/example/another-repo,'

            expect(getSuggestionsSearchTerm({ value: stringValue, caretPosition: stringValue.length })).toStrictEqual({
                repositories: ['github.com/example/about', 'github.com/example/another-repo'],
                value: 'github.com/example/another-repo',
                index: 1,
            })
        })
    })
})
