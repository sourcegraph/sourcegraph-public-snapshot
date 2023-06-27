import assert from 'assert'

import { findFilePaths, highlightTokens } from '.'

const markdownText = `# Title

This is a markdown [link](https://example.com)

This is \`/some/hallucinated/file/path\`. Hosted on github.com/sourcegraph.

Quoted "file/path.js". Unquoted hallucinated file/path/Class.java file.

This file is awesome cool/awesome. So is this this/is/a/directory and so/is/this and include the file "file/path.js".

The best part is that this files are  \`/some/hallucinated/file/path\`. And also have the files this/is/a/directory and so/is/this with "file/path.js".

This code client/cody-shared/test/ is usable.

\`\`\`
/file/path.go -- should be ignored
\`\`\`
`

const expectedHighlightedTokensText = `# Title

This is a markdown [link](https://example.com)

This is  <span class="token-file token-hallucinated" title="Hallucination detected: file does not exist">\`/some/hallucinated/file/path\`</span> . Hosted on github.com/sourcegraph.

Quoted  <span class="token-file token-not-hallucinated" title="No hallucination detected: file exists">"file/path.js"</span> . Unquoted hallucinated <span class="token-file token-hallucinated" title="Hallucination detected: file does not exist">file/path/Class.java</span> file.

This file is awesome cool/awesome. So is this <span class="token-file token-not-hallucinated" title="No hallucination detected: file exists">this/is/a/directory</span> and <span class="token-file token-not-hallucinated" title="No hallucination detected: file exists">so/is/this</span> and include the file  <span class="token-file token-not-hallucinated" title="No hallucination detected: file exists">"file/path.js"</span> .

The best part is that this files are   <span class="token-file token-hallucinated" title="Hallucination detected: file does not exist">\`/some/hallucinated/file/path\`</span> . And also have the files <span class="token-file token-not-hallucinated" title="No hallucination detected: file exists">this/is/a/directory</span> and <span class="token-file token-not-hallucinated" title="No hallucination detected: file exists">so/is/this</span> with  <span class="token-file token-not-hallucinated" title="No hallucination detected: file exists">"file/path.js"</span> .

This code <span class="token-file token-not-hallucinated" title="No hallucination detected: file exists">client/cody-shared/test/</span> is usable.

\`\`\`
/file/path.go -- should be ignored
\`\`\`
`

const validFilePaths = new Set(['file/path.js', 'this/is/a/directory', 'so/is/this', 'client/cody-shared/test/'])

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
            {
                input: '/items/:id',
                output: [],
            },
            {
                input: '/package.json',
                output: [],
            },
            {
                input: '/products/123',
                output: [{ fullMatch: '/products/123', pathMatch: '/products/123' }],
            },
            {
                input: '/src/components',
                output: [{ fullMatch: '/src/components', pathMatch: '/src/components' }],
            },
            {
                input: 'git/refs',
                output: [],
            },
            {
                input: 'git/refs/*',
                output: [],
            },
            {
                input: 'client/cody-shared/test/',
                output: [{ fullMatch: 'client/cody-shared/test/', pathMatch: 'client/cody-shared/test/' }],
            },
        ]
        for (const { input, output } of cases) {
            const actualOutput = findFilePaths(input)
            assert.deepStrictEqual(actualOutput, output)
        }
    })
})
