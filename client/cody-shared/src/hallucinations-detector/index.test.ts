import assert from 'assert'

import { findFilePaths, highlightTokens } from '.'

const markdownText = `# Title

This is \`/some/hallucinated/file/path\`. Hosted on github.com/sourcegraph.

Quoted "file/path.js". Unquoted hallucinated file/path/Class.java file.

This is a cool/awesome test. this/is/a/directory and so/is/this/ btw this/is/a/directory/ and so/is/this

The best file is file/path.js. The best directories are this/is/a/directory and this/is/a/directory/ and so/is/this. Another one: so/is/this/.

\`\`\`
/file/path.go -- should be ignored
\`\`\`
`

const expectedHighlightedTokensText = `# Title

This is  <span class="token-file token-hallucinated">\`/some/hallucinated/file/path\`</span> . Hosted on github.com/sourcegraph.

Quoted  <span class="token-file token-not-hallucinated">"file/path.js"</span> . Unquoted hallucinated <span class="token-file token-hallucinated">file/path/Class.java</span> file.

This is a cool/awesome test. <span class="token-file token-not-hallucinated">this/is/a/directory</span> and <span class="token-file token-not-hallucinated">so/is/this/</span> btw <span class="token-file token-not-hallucinated">this/is/a/directory/</span> and <span class="token-file token-not-hallucinated">so/is/this</span> 

The best file is file/path.js. The best directories are <span class="token-file token-not-hallucinated">this/is/a/directory</span> and <span class="token-file token-not-hallucinated">this/is/a/directory/</span> and <span class="token-file token-not-hallucinated">so/is/this</span> . Another one: <span class="token-file token-not-hallucinated">so/is/this</span> /.

\`\`\`
/file/path.go -- should be ignored
\`\`\`
`

const validFilePaths = new Set(['file/path.js', 'this/is/a/directory', 'so/is/this'])

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
