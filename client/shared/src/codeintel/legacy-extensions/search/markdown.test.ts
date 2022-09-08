import * as assert from 'assert'

import { wrapIndentationInCodeBlocks } from './markdown'

describe('wrapping indentation in code blocks', () => {
    it('wraps indentation in code blocks', () => {
        assert.deepStrictEqual(
            wrapIndentationInCodeBlocks(
                'java',
                `
prose

  code

prose
        `
            ),
            `
prose

\`\`\`java
  code
\`\`\`

prose
        `
        )
    })

    it('wraps indentation in code blocks', () => {
        assert.deepStrictEqual(
            wrapIndentationInCodeBlocks(
                'java',
                `
prose

  code

  code2
        `
            ),
            `
prose

\`\`\`java
  code

  code2
\`\`\`
        `
        )
    })
})
