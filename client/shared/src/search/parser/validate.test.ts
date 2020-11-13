import { isComplex } from './validate'
import { parseSearchQuery, ParseSuccess } from './parser'

describe('isComplex()', () => {
    test('a complex query', () => {
        expect(
            isComplex((parseSearchQuery('repo:foo (Github case:yes) or (organisation case:no)') as ParseSuccess).nodes)
        ).toBeTruthy()
    })

    test('a simple query', () => {
        expect(isComplex((parseSearchQuery('repo:foo Github case:yes') as ParseSuccess).nodes)).toBeFalsy()
    })
})
