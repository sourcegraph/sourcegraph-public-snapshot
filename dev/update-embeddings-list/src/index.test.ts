import { describe, expect, test } from '@jest/globals'

import { filter, sort, embeddedReposToMarkdown } from './index'

interface Repo {
    name: string
    url: string
}

interface Embedding {
    id: string
    state: string
    repo: Repo
}

describe('filter', () => {
    test('filters and sorts repo embedding jobs', () => {
        const input: Embedding[] = [
            {
                id: '1',
                state: 'COMPLETED',
                repo: {
                    name: 'b',
                    url: 'https://github.com/b',
                },
            },
            {
                id: '2',
                state: 'COMPLETED',
                repo: {
                    name: 'a',
                    url: 'https://github.com/a',
                },
            },
            {
                id: '3',
                state: 'PROCESSING',
                repo: {
                    name: 'c',
                    url: 'https://github.com/c',
                },
            },
        ]
        const expected = [
            {
                id: '1',
                state: 'COMPLETED',
                repo: {
                    name: 'b',
                    url: 'https://github.com/b',
                },
            },
            {
                id: '2',
                state: 'COMPLETED',
                repo: {
                    name: 'a',
                    url: 'https://github.com/a',
                },
            },
        ]
        expect(filter(input)).toEqual(expected)
    })
})

describe('sort', () => {
    test('sorts repos alphabetically', () => {
        const input: Repo[] = [
            {
                name: 'c/repo1',
                url: 'https://github.com/c/repo1',
            },
            {
                name: 'b/repo2',
                url: 'https://github.com/b/repo2',
            },
            {
                name: 'a/repo3',
                url: 'https://github.com/a/repo3',
            },
        ]
        const expected = [
            {
                name: 'a/repo3',
                url: 'https://github.com/a/repo3',
            },
            {
                name: 'b/repo2',
                url: 'https://github.com/b/repo2',
            },
            {
                name: 'c/repo1',
                url: 'https://github.com/c/repo1',
            },
        ]
        expect(sort(input)).toEqual(expected)
    })
})

describe('embeddedReposToMarkdown', () => {
    test('generates markdown for embedded repos', () => {
        const input: Embedding[] = [
            {
                id: '1',
                state: 'COMPLETED',
                repo: {
                    name: 'sourcegraph/sourcegraph',
                    url: 'https://github.com/sourcegraph/sourcegraph',
                },
            },
            {
                id: '2',
                state: 'COMPLETED',
                repo: {
                    name: 'golang/go',
                    url: 'https://github.com/golang/go',
                },
            },
        ]
        const expected = `# Embeddings for repositories with 5+ stars

Last updated: ${new Date().toLocaleString('en-US', {
            month: '2-digit',
            day: '2-digit',
            year: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
            timeZoneName: 'short',
        })}

1. [golang/go](https://github.com/golang/go)
1. [sourcegraph/sourcegraph](https://github.com/sourcegraph/sourcegraph)
`
        expect(embeddedReposToMarkdown(input)).toEqual(expected)
    })

    test('throws error if no repos', () => {
        expect(() => embeddedReposToMarkdown(undefined)).toThrowError('no embedded repos found!')
    })
})
