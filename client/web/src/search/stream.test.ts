import { RepositoryMatch, toGQLRepositoryMatch, toMarkdownCodeHtml } from './stream'

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
            '{"__typename":"Markdown","text":"[github.com/save/the andimals](/github.com/save/the%20andimals)"}'
        )
    })
})

describe('markdown cleanup', () => {
    test('markdown cleaned up correctly for diff match', () => {
        const diffMatch = `\`\`\`diff
test
\`\`\`
test
\`\`\`diff
test
\`\`\``

        expect(toMarkdownCodeHtml(diffMatch)).toMatchInlineSnapshot(
            '{"__typename":"Markdown","html":"test\\n```\\ntest\\n```diff\\ntest\\n","text":"```diff\\ntest\\n```\\ntest\\n```diff\\ntest\\n```"}'
        )
    })

    test('markdown cleaned up correctly for commit match', () => {
        const diffMatch = `\`\`\`COMMIT_EDITMSG
test
\`\`\`
test
\`\`\`COMMIT_EDITMSG
test
\`\`\``

        expect(toMarkdownCodeHtml(diffMatch)).toMatchInlineSnapshot(
            '{"__typename":"Markdown","html":"test\\n```\\ntest\\n```COMMIT_EDITMSG\\ntest\\n","text":"```COMMIT_EDITMSG\\ntest\\n```\\ntest\\n```COMMIT_EDITMSG\\ntest\\n```"}'
        )
    })
})
