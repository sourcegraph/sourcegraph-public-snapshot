import * as assert from 'assert'

import { describe, it } from 'vitest'

import { parseGitURI } from './uri'

describe('parseGitURI', () => {
    it('returns components', () => {
        assert.deepStrictEqual(
            parseGitURI(
                new URL('git://github.com/microsoft/vscode?dbd76d987cf1a412401bdbd3fb785217ac94197e#src/vs/css.js')
            ),
            {
                repo: 'github.com/microsoft/vscode',
                commit: 'dbd76d987cf1a412401bdbd3fb785217ac94197e',
                path: 'src/vs/css.js',
            }
        )
    })

    it('decodes repos with spaces', () => {
        assert.deepStrictEqual(
            parseGitURI(
                new URL(
                    'git://sourcegraph.visualstudio.com/Test%20Repo?dbd76d987cf1a412401bdbd3fb785217ac94197e#src/vs/css.js'
                )
            ),
            {
                repo: 'sourcegraph.visualstudio.com/Test Repo',
                commit: 'dbd76d987cf1a412401bdbd3fb785217ac94197e',
                path: 'src/vs/css.js',
            }
        )
    })
})
