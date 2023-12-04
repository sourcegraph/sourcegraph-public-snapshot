import * as assert from 'assert'

import { describe, it } from 'vitest'

import { createStubTextDocument } from '@sourcegraph/extension-api-stubs'

import type * as sourcegraph from '../api'

import { definitionQuery, referencesQuery } from './queries'

describe('search requests', () => {
    it('makes correct search requests for goto definition', () => {
        interface DefinitionTest {
            doc: sourcegraph.TextDocument
            expectedSearchQueryTerms: string[]
        }
        const tests: DefinitionTest[] = [
            {
                doc: createStubTextDocument({
                    uri: 'git://github.com/foo/bar?rev#file.cpp',
                    languageId: 'cpp',
                    text: 'token',
                }),
                expectedSearchQueryTerms: [
                    '^token$',
                    'type:symbol',
                    'patternType:regexp',
                    'count:50',
                    'case:yes',
                    'file:\\.(cpp)$',
                ],
            },
        ]

        for (const test of tests) {
            assert.deepStrictEqual(
                definitionQuery({
                    searchToken: 'token',
                    doc: test.doc,
                    fileExts: ['cpp'],
                }),
                test.expectedSearchQueryTerms
            )
        }
    })

    it('makes correct search requests for references', () => {
        interface ReferencesTest {
            doc: sourcegraph.TextDocument
            expectedSearchQueryTerms: string[]
        }
        const tests: ReferencesTest[] = [
            {
                doc: createStubTextDocument({
                    uri: 'git://github.com/foo/bar?rev#file.cpp',
                    languageId: 'cpp',
                    text: 'token',
                }),
                expectedSearchQueryTerms: [
                    '\\btoken\\b',
                    'type:file',
                    'patternType:regexp',
                    'count:500',
                    'case:yes',
                    'file:\\.(cpp)$',
                ],
            },
        ]

        for (const test of tests) {
            assert.deepStrictEqual(
                referencesQuery({
                    searchToken: 'token',
                    doc: test.doc,
                    fileExts: ['cpp'],
                }),
                test.expectedSearchQueryTerms
            )
        }
    })
})
