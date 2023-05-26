import * as assert from 'assert'

import { Term, regexForTerms, userQueryToKeywordQuery } from './local-keyword-context-fetcher'

describe('keyword context', () => {
    it('userQueryToKeywordQuery', () => {
        const cases: { query: string; expected: Term[] }[] = [
            {
                query: 'Where is auth in Sourcegraph?',
                expected: [
                    {
                        count: 1,
                        originals: ['Where', 'Where'],
                        prefix: 'where',
                        stem: 'where',
                    },
                    {
                        count: 1,
                        originals: ['auth', 'auth'],
                        prefix: 'auth',
                        stem: 'auth',
                    },
                    {
                        count: 1,
                        originals: ['Sourcegraph', 'Sourcegraph'],
                        prefix: 'sourcegraph',
                        stem: 'sourcegraph',
                    },
                ],
            },
            {
                query: `Explain the following code at a high level:
uint32_t PackUInt32(const Color& color) {
  uint32_t result = 0;
  result |= static_cast<uint32_t>(color.r * 255 + 0.5f) << 24;
  result |= static_cast<uint32_t>(color.g * 255 + 0.5f) << 16;
  result |= static_cast<uint32_t>(color.b * 255 + 0.5f) << 8;
  result |= static_cast<uint32_t>(color.a * 255 + 0.5f);
  return result;
}
`,
                expected: [
                    {
                        count: 1,
                        originals: ['Explain', 'Explain'],
                        prefix: 'explain',
                        stem: 'explain',
                    },
                    {
                        count: 1,
                        originals: ['following', 'following'],
                        prefix: 'follow',
                        stem: 'follow',
                    },
                    {
                        count: 1,
                        originals: ['code', 'code'],
                        prefix: 'code',
                        stem: 'code',
                    },
                    {
                        count: 1,
                        originals: ['high', 'high'],
                        prefix: 'high',
                        stem: 'high',
                    },
                    {
                        count: 1,
                        originals: ['level', 'level'],
                        prefix: 'level',
                        stem: 'level',
                    },
                    {
                        count: 6,
                        originals: ['uint32_t', 'uint32_t', 'uint32_t', 'uint32_t', 'uint32_t', 'uint32_t', 'uint32_t'],
                        prefix: 'uint',
                        stem: 'uinty2_t',
                    },
                    {
                        count: 1,
                        originals: ['PackUInt32', 'PackUInt32'],
                        prefix: 'packuint',
                        stem: 'packuinty2',
                    },
                    {
                        count: 1,
                        originals: ['const', 'const'],
                        prefix: 'const',
                        stem: 'const',
                    },
                    {
                        count: 6,
                        originals: ['Color', 'Color', 'color', 'color', 'color', 'color', 'color'],
                        prefix: 'color',
                        stem: 'color',
                    },
                    {
                        count: 6,
                        originals: ['result', 'result', 'result', 'result', 'result', 'result', 'result'],
                        prefix: 'result',
                        stem: 'result',
                    },
                    {
                        count: 4,
                        originals: ['static_cast', 'static_cast', 'static_cast', 'static_cast', 'static_cast'],
                        prefix: 'static_cast',
                        stem: 'static_cast',
                    },
                    {
                        count: 4,
                        originals: ['255', '255', '255', '255', '255'],
                        prefix: '255',
                        stem: '255',
                    },
                    {
                        count: 1,
                        originals: ['return', 'return'],
                        prefix: 'return',
                        stem: 'return',
                    },
                ],
            },
        ]
        for (const testcase of cases) {
            const actual = userQueryToKeywordQuery(testcase.query)
            assert.deepStrictEqual(actual, testcase.expected)
        }
    })
    it('query to regex', () => {
        const trials: {
            userQuery: string
            expRegex: string
        }[] = [
            {
                userQuery: 'Where is auth in Sourcegraph?',
                expRegex: '(?:where|auth|sourcegraph)',
            },
            {
                userQuery: 'saml auth handler',
                expRegex: '(?:saml|auth|handler)',
            },
            {
                userQuery: 'Where is the HTTP middleware defined in this codebase?',
                expRegex: '(?:where|http|middlewar|defin|codebas)',
            },
        ]
        for (const trial of trials) {
            const terms = userQueryToKeywordQuery(trial.userQuery)
            const regex = regexForTerms(...terms)
            expect(regex).toEqual(trial.expRegex)
        }
    })
})
