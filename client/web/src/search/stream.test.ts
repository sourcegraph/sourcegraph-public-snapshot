import { toGQLRepositoryMatch, RepositoryMatch } from './stream'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value),
    test: () => true,
})

export const REPO_MATCH_CONTAINING_SPACES: RepositoryMatch = {
    type: 'repo',
    repository: 'github.com/save/the andimals',
}

describe('escapeSpaces', () => {
    test('escapes spaces in value', () => {
        expect(toGQLRepositoryMatch(REPO_MATCH_CONTAINING_SPACES).label).toMatchInlineSnapshot(
            '{"__typename":"Markdown","text":"[github.com/save/the andimals](github.com/save/the%20andimals)"}'
        )
    })
})
