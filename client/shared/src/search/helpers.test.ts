import { toggleSubquery } from './helpers'

describe('search/helpers', () => {
    describe('toggleSearchFilter', () => {
        it('adds filter if it is not already in query', () => {
            expect(toggleSubquery('repo:test ', 'lang:c++')).toStrictEqual('repo:test lang:c++ ')
        })

        it('adds filter if it is not already in query, even if it matches substring for an existing filter', () => {
            expect(toggleSubquery('repo:test lang:c++ ', 'lang:c')).toStrictEqual('repo:test lang:c++ lang:c ')
        })

        it('removes filter from query it it exists', () => {
            expect(toggleSubquery('repo:test lang:c++ lang:c ', 'lang:c')).toStrictEqual('repo:test lang:c++')
        })
    })
})
