import assert from 'assert'

import { findFilePaths, highlightTokens } from '.'

const markdownText = `# Title

This is \`/some/hallucinated/file/path\`. Hosted on github.com/sourcegraph.

Quoted "file/path.js". Unquoted hallucinated file/path/Class.java file.

This file is awesome cool/awesome. So is this this/is/a/directory and so/is/this and include the file "file/path.js".

The best part is that this files are  \`/some/hallucinated/file/path\`. And also have the files this/is/a/directory and so/is/this with "file/path.js".


\`\`\`
/file/path.go -- should be ignored
\`\`\`
`

const expectedHighlightedTokensText = `# Title

This is  <span class="token-file token-hallucinated" title="Hallucination detected: file does not exist">\`/some/hallucinated/file/path\`</span> . Hosted on github.com/sourcegraph.

Quoted  <span class="token-file token-not-hallucinated" title="No hallucination detected: file exists">"file/path.js"</span> . Unquoted hallucinated <span class="token-file token-hallucinated" title="Hallucination detected: file does not exist">file/path/Class.java</span> file.

This file is awesome cool/awesome. So is this <span class="token-file token-not-hallucinated" title="No hallucination detected: file exists">this/is/a/directory</span> and <span class="token-file token-not-hallucinated" title="No hallucination detected: file exists">so/is/this</span> and include the file  <span class="token-file token-not-hallucinated" title="No hallucination detected: file exists">"file/path.js"</span> .

The best part is that this files are   <span class="token-file token-hallucinated" title="Hallucination detected: file does not exist">\`/some/hallucinated/file/path\`</span> . And also have the files <span class="token-file token-not-hallucinated" title="No hallucination detected: file exists">this/is/a/directory</span> and <span class="token-file token-not-hallucinated" title="No hallucination detected: file exists">so/is/this</span> with  <span class="token-file token-not-hallucinated" title="No hallucination detected: file exists">"file/path.js"</span> .


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
