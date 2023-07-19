import { faker } from '@faker-js/faker'

import type { GitCommitFields, HistoryResult, SignatureFields } from '$lib/graphql-operations'

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
