import assert from 'assert'

import { findFilePaths, highlightTokens } from '.'

const markdownText = `# Title

This is \`/some/hallucinated/file/path\`. Hosted on github.com/sourcegraph.

Quoted "file/path.js". Unquoted hallucinated file/path/Class.java file.

This is a cool/awesome test.

\`\`\`
/file/path.go -- should be ignored
\`\`\`
`

const expectedHighlightedTokensText = `# Title

This is \`<span class="token-file token-hallucinated">/some/hallucinated/file/path</span>\`. Hosted on github.com/sourcegraph.

Quoted "<span class="token-file token-not-hallucinated">file/path.js</span>". Unquoted hallucinated <span class="token-file token-hallucinated">file/path/Class.java</span> file.

This is a cool/awesome test.

\`\`\`
/file/path.go -- should be ignored
\`\`\`
`

const validFilePaths = new Set(['file/path.js'])

describe('Hallucinations detector', () => {
    it('highlights hallucinated file paths', async () => {
        const { text } = await highlightTokens(markdownText, filePaths => {
            const filePathExists: { [filePath: string]: boolean } = {}
            for (const filePath of filePaths) {
                filePathExists[filePath] = validFilePaths.has(filePath)
            }
            return Promise.resolve(filePathExists)
        })
        assert.deepStrictEqual(text, expectedHighlightedTokensText)
    })

    it('findFilePaths', () => {
        const cases: {
            input: string
            output: { pathMatch: string; fullMatch: string }[]
        }[] = [
            {
                input: 'foo/bar/baz',
                output: [{ fullMatch: 'foo/bar/baz', pathMatch: 'foo/bar/baz' }],
            },
            {
                input: 'use of this/that in a sentence',
                output: [],
            },
            {
                input: '`this/that`',
                output: [{ fullMatch: '`this/that`', pathMatch: 'this/that' }],
            },
            {
                input: 'pattern/foo/bar/*.ts',
                output: [],
            },
            {
                input: 'remix-run/react',
                output: [{ fullMatch: 'remix-run/react', pathMatch: 'remix-run/react' }],
            },
            {
                input: '`@remix-run/react`',
                output: [],
            },
            {
                input: '@remix-run/react',
                output: [],
            },
        ]
        for (const { input, output } of cases) {
            const actualOutput = findFilePaths(input)
            assert.deepStrictEqual(actualOutput, output)
        }
    })
})
