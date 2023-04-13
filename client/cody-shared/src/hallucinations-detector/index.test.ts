import assert from 'assert'

import { highlightTokens } from '.'

const markdownText = `# Title

This is \`/some/hallucinated/file/path\`. Hosted on github.com/sourcegraph.

Quoted "file/path.js". Unquoted hallucinated file/path/Class.java file.

This is a cool/awesome test.

\`\`\`
/file/path.go -- should be ignored
\`\`\`
`

const expectedHighlightedTokensText = `# Title

This is  <span class="token-file token-hallucinated">\`/some/hallucinated/file/path\`</span> . Hosted on github.com/sourcegraph.

Quoted  <span class="token-file token-not-hallucinated">"file/path.js"</span> . Unquoted hallucinated <span class="token-file token-hallucinated">file/path/Class.java</span> file.

This is a cool/awesome test.

\`\`\`
/file/path.go -- should be ignored
\`\`\`
`

const validFilePaths = new Set(['file/path.js'])

describe('Hallucinations detector', () => {
    it('highlights hallucinated file paths', async () => {
        const { text } = await highlightTokens(markdownText, filePath => Promise.resolve(validFilePaths.has(filePath)))
        assert.deepStrictEqual(text, expectedHighlightedTokensText)
    })
})
