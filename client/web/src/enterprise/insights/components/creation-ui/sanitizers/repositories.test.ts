import { getSanitizedRepositories } from './repositories'

describe('getSanitizedRepositories', () => {
    it('should return an empty list in case of an empty user input', () => {
        expect(getSanitizedRepositories('')).toStrictEqual([])
    })

    it('should return repositories list', () => {
        expect(getSanitizedRepositories('github.com/a/a, github.com/b/a')).toStrictEqual([
            'github.com/a/a',
            'github.com/b/a',
        ])
    })

    it('should return repositories list with additional whitespaces and commas', () => {
        expect(getSanitizedRepositories('   github.com/a/a, github.com/b/a  ,   ,')).toStrictEqual([
            'github.com/a/a',
            'github.com/b/a',
        ])
    })
})
