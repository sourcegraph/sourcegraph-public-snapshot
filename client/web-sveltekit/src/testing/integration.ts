import { test as base, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { readFileSync } from 'node:fs'
import { buildSchema } from 'graphql'
import type { TypeMocks, ObjectMock, UserMock } from './graphql-type-mocks'
import glob from 'glob'
import { faker } from '@faker-js/faker'
import { GraphQLMockServer } from './graphql-mocking'

export { expect, defineConfig } from '@playwright/test'

const defaultMocks: TypeMocks = {
    Query: () => ({
        // null means not signed in
        currentUser: null,
    }),
    SettingsCascade: () => ({
        // Ensure this is valid JSON
        final: '{}',
    }),
    TemporarySettings: () => ({
        // Ensure this is valid JSON
        contents: '{}',
    }),
    GitBlob: () => ({
        highlight: {
            lsif: '{}',
        },
    }),
    Signature: () => ({
        date: faker.date.past().toISOString(),
    }),
    GitObjectID: () => faker.git.commitSha(),
}

const SCHEMA_DIR = path.resolve(
    path.join(path.dirname(fileURLToPath(import.meta.url)), '../../../../cmd/frontend/graphqlbackend')
)
const typeDefs = glob
    .sync('**/*.graphql', { cwd: SCHEMA_DIR })
    .map(file => readFileSync(path.join(SCHEMA_DIR, file), 'utf8'))
    .join('\n')

class Sourcegraph {
    private debugMode = false
    constructor(private readonly page: Page, private readonly graphqlMock: GraphQLMockServer) {}

    async setup(): Promise<void> {
        await this.page.route(/\.api\/graphql/, route => {
            const { query, variables, operationName } = JSON.parse(route.request().postData() ?? '')
            const result = this.graphqlMock.query(query, variables, operationName)
            if (this.debugMode) {
                console.log('incoming graphql query', operationName, variables, result)
            }
            route.fulfill({ json: result })
        })
    }

    public debug() {
        this.debugMode = true
    }

    public mock(mocks: TypeMocks, operationName?: string): void {
        this.graphqlMock.addMocks(mocks, operationName)
    }

    public fixture(fixtures: (ObjectMock & { __typename: NonNullable<ObjectMock['__typename']> })[]): void {
        this.graphqlMock.addFixtures(fixtures)
    }

    public signIn(userMock: UserMock = {}): void {
        this.mock(
            {
                Query: () => ({
                    currentUser: {
                        avatarURL: null,
                        ...userMock,
                    },
                }),
            },
            'Init'
        )
    }

    public signOut(): void {
        this.mock(
            {
                Query: () => ({
                    currentUser: null,
                }),
            },
            'Init'
        )
    }

    public teardown(): void {
        this.graphqlMock.reset()
    }
}

export const test = base.extend<{ sg: Sourcegraph }, { graphqlMock: GraphQLMockServer }>({
    sg: [
        async ({ page, graphqlMock }, use) => {
            const sg = new Sourcegraph(page, graphqlMock)
            await sg.setup()
            await use(sg)
            sg.teardown()
        },
        { auto: true },
    ],
    graphqlMock: [
        async ({}, use) => {
            const graphqlMock = new GraphQLMockServer({
                schema: buildSchema(typeDefs),
                mocks: defaultMocks,
                typePolicies: {
                    GitBlob: {
                        keyField: 'canonicalURL',
                    },
                },
            })
            await use(graphqlMock)
        },
        { scope: 'worker' },
    ],
})
