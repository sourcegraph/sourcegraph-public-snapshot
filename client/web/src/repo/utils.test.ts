import { fail } from 'assert'

import { describe, expect, it } from 'vitest'

import { SymbolKind, SymbolNodeFields } from '../graphql-operations'
import { parseBrowserRepoURL } from '../util/url'

import { SymbolWithChildren, getInitialSearchTerm, hierarchyOf } from './utils'

describe('getInitialSearchTerm', () => {
    const tests: {
        name: string
        repo: string
        expected: string
    }[] = [
        {
            name: 'works with a github repo url',
            repo: 'github.com/sourcegraph/sourcegraph',
            expected: 'sourcegraph',
        },
        {
            name: 'works with a gitlab repo url',
            repo: 'gitlab.com/SourcegraphCody/jsonrpc2',
            expected: 'jsonrpc2',
        },
        {
            name: 'works with a perforce depot url',
            repo: 'public.perforce.com/sourcegraph/myp4depot',
            expected: 'myp4depot',
        },
        {
            name: 'works with a bitbucket repo name',
            repo: 'bitbucket.org/username/projectname/mybitbucketrepo',
            expected: 'mybitbucketrepo',
        },
        {
            name: 'works with a gerrit repo name',
            repo: 'mygerritserver.com/c/mygerritrepo',
            expected: 'mygerritrepo',
        },
        {
            name: 'works with an Azure DevOps repo name',
            repo: 'https://dev.azure.com/myADOorgname/myADOproject/_git/myADOrepo',
            expected: 'myADOrepo',
        },
        {
            name: 'works with a Plastic SCM repo name',
            repo: 'https://cloud.plasticscm.com/my-plastic-repo',
            expected: 'my-plastic-repo',
        },
    ]

    for (const t of tests) {
        it(t.name, () => {
            expect(getInitialSearchTerm(t.repo)).toBe(t.expected)
        })
    }
})

describe('hiararchyOf', () => {
    /** Recursively check that all symbols are ordered by line numbers in their URLs. */
    const expectOrdered = (symbols: SymbolWithChildren[]) => {
        let lastLine = 0
        let lastName: string | undefined

        for (const sym of symbols) {
            const url = parseBrowserRepoURL((sym as SymbolNodeFields).url)
            const thisLine = url.position?.line
            if (!thisLine) {
                continue
            }
            if (thisLine >= lastLine) {
                lastLine = thisLine
                lastName = sym.name
                if (sym.children && sym.children.length) {
                    expectOrdered(sym.children)
                }
                continue
            }
            const containerName = (sym as SymbolNodeFields).containerName ?? ''
            fail(
                `expected ascending line numbers:
                    (prev) ${containerName}.${lastName} -> L:${lastLine}
                    (curr) ${containerName}.${sym.name} -> L:${thisLine}`
            )
        }
    }

    const tests: {
        name: string
        symbols: SymbolNodeFields[]
        expectFunc: (res: SymbolWithChildren[]) => void
    }[] = [
        {
            name: 'structures deeply nested hierarchies',
            symbols: [
                {
                    __typename: 'Symbol',
                    name: 'StaticClass',
                    containerName: 'repo.ParentClass',
                    language: 'Java',
                    url: 'github.com/repo/ParentClass.java?L8:5-18:15',
                    kind: SymbolKind.CLASS,
                    location: {
                        __typename: 'Location',
                        resource: { path: 'ParentClass.java' },
                        range: { start: { line: 8, character: 5 }, end: { line: 8, character: 15 } },
                    },
                },
                {
                    __typename: 'Symbol',
                    name: 'PrivateClass',
                    containerName: 'repo',
                    language: 'Java',
                    url: 'github.com/repo/ParentClass.java?L18:2-18:10',
                    kind: SymbolKind.CLASS,
                    location: {
                        __typename: 'Location',
                        resource: { path: 'ParentClass.java' },
                        range: { start: { line: 18, character: 2 }, end: { line: 18, character: 10 } },
                    },
                },
                {
                    __typename: 'Symbol',
                    name: 'StaticClassProperty',
                    containerName: 'repo.ParentClass.StaticClass',
                    language: 'Java',
                    url: 'github.com/repo/ParentClass.java?L26:2-26:11',
                    kind: SymbolKind.PROPERTY,
                    location: {
                        __typename: 'Location',
                        resource: { path: 'ParentClass.java' },
                        range: { start: { line: 26, character: 2 }, end: { line: 26, character: 11 } },
                    },
                },

                {
                    __typename: 'Symbol',
                    name: 'ParentClass',
                    containerName: 'repo',
                    language: 'Java',
                    url: 'github.com/repo/ParentClass.java?L35:1-35:6',
                    kind: SymbolKind.CLASS,
                    location: {
                        __typename: 'Location',
                        resource: { path: 'ParentClass.java' },
                        range: { start: { line: 35, character: 1 }, end: { line: 35, character: 6 } },
                    },
                },
            ],
            expectFunc: (got: SymbolWithChildren[]) => {
                const topLevel: SymbolWithChildren[] = got[0].children
                let parent: SymbolWithChildren | undefined

                topLevel.forEach(sym => {
                    if (sym.children.length) {
                        parent = sym
                        return
                    }
                })
                expect(parent).not.toBeUndefined()
                expect(parent!.children).toHaveLength(1)

                const child: SymbolWithChildren = parent!.children[0]
                expect(child.children).toHaveLength(1)
            },
        },
        {
            name: 'handles orphaned symbols',
            symbols: [
                {
                    __typename: 'Symbol',
                    name: 'ChildNodeA',
                    containerName: 'Parent',
                    language: 'Python',
                    url: 'github.com/repo/file.py?L15:2-15:10',
                    kind: SymbolKind.CLASS,
                    location: {
                        __typename: 'Location',
                        resource: { path: 'file.py' },
                        range: { start: { line: 15, character: 2 }, end: { line: 15, character: 10 } },
                    },
                },
                {
                    __typename: 'Symbol',
                    name: 'ChildNodeB',
                    containerName: 'Parent',
                    language: 'Python',
                    url: 'github.com/repo/file.py?L18:2-18:10',
                    kind: SymbolKind.FIELD,
                    location: {
                        __typename: 'Location',
                        resource: { path: 'file.py' },
                        range: { start: { line: 18, character: 2 }, end: { line: 18, character: 10 } },
                    },
                },
            ],
            expectFunc: (got: SymbolWithChildren[]) => {
                const parent = got[0]
                expect(parent.__typename).toEqual('SymbolPlaceholder')
                expect(parent.children).toHaveLength(2)
            },
        },
        {
            name: 'handles variables with identical names in the same scope',
            symbols: [
                {
                    __typename: 'Symbol',
                    name: '_',
                    containerName: 'pkg',
                    language: 'Go',
                    url: 'github.com/repo/file.go?L5:1-5:2',
                    kind: SymbolKind.VARIABLE,
                    location: {
                        __typename: 'Location',
                        resource: { path: 'file.go' },
                        range: { start: { line: 1, character: 1 }, end: { line: 5, character: 2 } },
                    },
                },
                {
                    __typename: 'Symbol',
                    name: '_',
                    containerName: 'pkg',
                    language: 'Go',
                    url: 'github.com/repo/file.go?L1:1-1:2',
                    kind: SymbolKind.VARIABLE,
                    location: {
                        __typename: 'Location',
                        resource: { path: 'file.go' },
                        range: { start: { line: 1, character: 1 }, end: { line: 5, character: 2 } },
                    },
                },
            ],
            expectFunc: (got: SymbolWithChildren[]) => {
                const topLevel = got[0].children
                expect(topLevel).toHaveLength(2)

                const sym1 = topLevel[0] as SymbolNodeFields,
                    sym2 = topLevel[1] as SymbolNodeFields
                expect(sym1.url).not.toEqual(sym2.url)
            },
        },
    ]

    for (const t of tests) {
        it(t.name, () => {
            const got = hierarchyOf(t.symbols)

            expect(got).toHaveLength(1)
            expectOrdered(got)
            t.expectFunc(got)
        })
    }
})
