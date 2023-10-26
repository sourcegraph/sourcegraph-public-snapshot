import { faker } from '@faker-js/faker'
import { range } from 'lodash'

import { SymbolKind, type GitCommitFields, type HistoryResult, type SignatureFields } from '$lib/graphql-operations'
import type { HighlightedFileVariables, HighlightedFileResult } from '$lib/graphql-operations'
import type { CommitMatch, ContentMatch, PersonMatch, TeamMatch, PathMatch, SymbolMatch } from '$lib/shared'

/**
 * Initializes faker's randomness generator with a fixed seed, for
 * deterministic results.
 */
export function seed(seed?: number): number {
    return faker.seed(seed)
}

export function createSignature(): SignatureFields {
    const firstName = faker.person.firstName()
    const lastName = faker.person.lastName()
    const displayName = faker.internet.displayName({ firstName, lastName })

    return {
        person: {
            displayName,
            name: faker.person.fullName({ firstName, lastName }),
            avatarURL: faker.internet.avatar(),
            user: {
                displayName,
                id: faker.string.uuid(),
                url: faker.internet.url(),
                username: faker.internet.userName({ firstName, lastName }),
            },
        },
        date: faker.date.past().toISOString(),
    }
}

export function createGitCommit(initial?: Partial<GitCommitFields>): GitCommitFields {
    const oid = faker.git.commitSha()

    return {
        id: faker.string.uuid(),
        oid,
        abbreviatedOID: oid.slice(0, 7),
        subject: faker.git.commitMessage(),
        body: faker.lorem.paragraph(),
        author: createSignature(),
        committer: faker.helpers.maybe(createSignature) ?? null,
        parents: faker.helpers.multiple(
            () => {
                const oid = faker.git.commitSha()
                return {
                    oid,
                    abbreviatedOID: oid.slice(0, 7),
                    url: faker.internet.url(),
                }
            },
            { count: { min: 1, max: 2 } }
        ),
        url: faker.internet.url(),
        canonicalURL: faker.internet.url(),
        externalURLs: [],
        ...initial,
    }
}

export function createHistoryResults(count: number, pageSize: number): HistoryResult[] {
    return Array.from({ length: count }, (_, index) => ({
        nodes: faker.helpers.uniqueArray(createGitCommit, pageSize),
        pageInfo: {
            hasNextPage: index < count - 1,
            endCursor: String((index + 1) * pageSize),
        },
    }))
}

/**
 * Converts the input string to lower case and replaces all non-word characters with -
 */
function clean(str: string): string {
    return str.replaceAll(/\W+/g, '-').toLowerCase()
}

function createRepoName(): string {
    return `github.com/${clean(faker.company.name())}/${clean(faker.company.buzzNoun())}`
}

function createCommitURL(repoName: string, commitOID: string): string {
    return `${repoName}/-/commit/${commitOID}`
}

function createGitCommitMessage(): string {
    return faker.git.commitMessage() + '\n\n' + faker.lorem.paragraphs({ min: 0, max: 3 })
}

function createRepoStars(): number | undefined {
    return faker.helpers.maybe(() => faker.number.int({ max: 1000000 }))
}

function createUnifiedDiff(): string {
    const file = faker.system.filePath()
    return [
        `${file} ${file}`,
        ...faker.helpers.multiple(
            () => {
                const lineNew = faker.number.int({ min: 0, max: 1000 })
                const lineOld = faker.number.int({ min: lineNew, max: lineNew + 10 })

                return [
                    `@@ -${lineNew} +${lineOld} @@`,
                    ...faker.helpers.multiple(
                        () => `${faker.helpers.arrayElement([' ', '-', '+'])} ${loremLine(MAX_LINE_LENGTH)}`,
                        { count: { min: 3, max: 8 } }
                    ),
                ].join('\n')
            },
            { count: { min: 1, max: 3 } }
        ),
    ].join('\n')
}

export function createCommitMatch(
    type: 'diff' | 'commit' = faker.helpers.arrayElement(['diff', 'commit'])
): CommitMatch {
    const diff = type === 'diff'
    const repository = createRepoName()
    const oid = faker.git.commitSha()
    const message = createGitCommitMessage()
    const content = diff ? createUnifiedDiff() : message
    return {
        type: 'commit',
        oid,
        url: createCommitURL(repository, oid),
        ranges: (() => {
            const lines = content.split('\n')
            return faker.helpers
                .uniqueArray(
                    range(0, lines.length).filter(line => lines[line].length > 3),
                    3
                )
                .map(line => {
                    const start = faker.number.int({ max: lines[line].length - 3 })
                    const length = faker.number.int({
                        min: 3,
                        max: Math.min(MAX_HIGHLIGHT_LENGTH, lines[line].length - start),
                    })
                    return [line + 1, start, length]
                })
        })(),
        content: ['```', diff ? 'DIFF' : 'COMMIT', '\n', content, '\n```'].join(''),
        message,
        authorDate: faker.date.recent().toISOString(),
        authorName: faker.person.fullName(),
        repository,
        repoStars: createRepoStars(),
        committerDate: faker.date.recent().toISOString(),
        committerName: faker.person.fullName(),
    }
}

const MAX_LINE_LENGTH = 100
const MAX_HIGHLIGHT_LENGTH = 10

function createRange(
    line: number,
    maxLength: number = MAX_LINE_LENGTH
): {
    start: { line: number; column: number; offset: number }
    end: { line: number; column: number; offset: number }
} {
    const startColumn = faker.number.int({ max: maxLength - 1 })

    const start = {
        line,
        column: startColumn,
        offset: faker.number.int({ min: startColumn, max: maxLength - 1 }),
    }
    const end = {
        line,
        column: faker.number.int({ min: start.column + 1, max: maxLength }),
        offset: faker.number.int({ min: start.offset + 1, max: maxLength }),
    }

    return {
        start,
        end,
    }
}

export function createContentMatch(): ContentMatch {
    const repository = createRepoName()
    const path = faker.system.filePath()

    return {
        type: 'content',
        path,
        repository,
        repoStars: createRepoStars(),
        chunkMatches: faker.helpers.uniqueArray(range(1000, 20), faker.number.int({ min: 1, max: 10 })).map(line => {
            const content = faker.lorem.lines(5)
            const ranges = faker.helpers
                .uniqueArray(range(line, line + 3), faker.number.int({ min: 1, max: 5 }))
                .map(line => createRange(line))
                .sort((a, b) => a.start.line - b.start.line)
            return {
                content,
                ranges,
                contentStart: {
                    line,
                    column: 1,
                    offset: 1,
                },
            }
        }),
        pathMatches: faker.helpers.maybe(() =>
            faker.helpers.multiple(() => createRange(0, path.length), { count: { min: 0, max: 3 } })
        ),
    }
}

export function createPersonMatch(): PersonMatch {
    const username = faker.internet.userName()
    return {
        type: 'person',
        handle: faker.helpers.maybe(() => username),
        user: faker.helpers.maybe(() => ({
            username,
            avatarURL: faker.helpers.maybe(() => faker.internet.avatar()),
            displayName: faker.helpers.maybe(() =>
                faker.helpers.arrayElement([faker.person.fullName(), faker.internet.displayName()])
            ),
        })),
        email: faker.helpers.maybe(() => faker.internet.email()),
    }
}

export function createTeamMatch(): TeamMatch {
    const handle = faker.company.buzzNoun()
    return {
        type: 'team',
        name: handle + ' team',
        handle: faker.helpers.maybe(() => handle),
        email: faker.helpers.maybe(() => faker.internet.email()),
    }
}

export function createPathMatch(): PathMatch {
    const path = faker.system.filePath()
    return {
        type: 'path',
        repository: createRepoName(),
        path,
        repoStars: createRepoStars(),
        pathMatches: faker.helpers.maybe(() =>
            faker.helpers.multiple(() => createRange(0, path.length), { count: { min: 0, max: 3 } })
        ),
    }
}
export function createSymbolMatch(): SymbolMatch {
    const path = faker.system.filePath()
    return {
        type: 'symbol',
        repository: createRepoName(),
        path,
        repoStars: createRepoStars(),
        symbols: faker.helpers.multiple(
            () => ({
                line: faker.number.int({ min: 1, max: 1000 }),
                url: faker.internet.url(),
                kind: faker.helpers.enumValue(SymbolKind),
                name: faker.lorem.word(),
                containerName: faker.lorem.word(),
            }),
            { count: { min: 1, max: 5 } }
        ),
    }
}

function loremLine(minLength: number): string {
    let content = ''
    do {
        content += faker.lorem.sentence() + ' '
    } while (content.length < minLength)

    return content
}

function colorize(line: string): string {
    return line
        .split(' ')
        .map(word => faker.helpers.maybe(() => `<span style="color: ${faker.color.rgb()}">${word}</span>`) ?? word)
        .join(' ')
}

export function createHighlightedFileResult(ranges: HighlightedFileVariables['ranges']): HighlightedFileResult {
    return {
        repository: {
            id: faker.string.uuid(),
            commit: {
                id: faker.string.uuid(),
                blob: {
                    canonicalURL: faker.internet.url(),
                    isDirectory: false,
                    highlight: {
                        aborted: false,
                        lineRanges: ranges.map(({ startLine, endLine }) =>
                            range(startLine, endLine).map(
                                line =>
                                    `<tr><td class="line" data-line="${line}"></td><td class="code annotated-selection-match">${colorize(
                                        loremLine(MAX_LINE_LENGTH)
                                    )}</td></tr>`
                            )
                        ),
                    },
                },
            },
        },
    }
}
