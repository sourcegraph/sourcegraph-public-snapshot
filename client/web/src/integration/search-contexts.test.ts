import { subDays } from 'date-fns'
import expect from 'expect'
import { range } from 'lodash'
import { afterEach, beforeEach, describe, test } from 'mocha'

import type { SharedGraphQlOperations, SearchContextMinimalFields } from '@sourcegraph/shared/src/graphql-operations'
import {
    highlightFileResult,
    mixedSearchStreamEvents,
} from '@sourcegraph/shared/src/search/integration/streaming-search-mocks'
import { accessibilityAudit } from '@sourcegraph/shared/src/testing/accessibility'
import { type Driver, createDriverForTest } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import type { WebGraphQlOperations } from '../graphql-operations'

import { type WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import { commonWebGraphQlResults, createViewerSettingsGraphQLOverride } from './graphQlResults'
import { createEditorAPI, getSearchQueryInputConfig, percySnapshotWithVariants } from './utils'

const commonSearchGraphQLResults: Partial<WebGraphQlOperations & SharedGraphQlOperations> = {
    ...commonWebGraphQlResults,
    ...highlightFileResult,
}

describe('Search contexts', () => {
    let driver: Driver
    before(async () => {
        driver = await createDriverForTest()
    })
    after(() => driver?.close())
    let testContext: WebIntegrationTestContext
    beforeEach(async function () {
        testContext = await createWebIntegrationTestContext({
            driver,
            currentTest: this.currentTest!,
            directory: __dirname,
        })
        testContext.overrideGraphQL(testContextForSearchContexts)
        testContext.overrideSearchStreamEvents(mixedSearchStreamEvents)
    })
    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    const testContextForSearchContexts: Partial<WebGraphQlOperations> = {
        ...commonSearchGraphQLResults,
        ...createViewerSettingsGraphQLOverride({
            user: {
                experimentalFeatures: {
                    showSearchContext: true,
                    searchQueryInput: 'v1',
                },
            },
        }),
    }

    const getSelectedSearchContextSpec = () =>
        driver.page.evaluate(() => document.querySelector('.test-selected-search-context-spec')?.textContent)

    const isSearchContextDropdownDisabled = () =>
        driver.page.evaluate(() => document.querySelector<HTMLButtonElement>('.test-search-context-dropdown')?.disabled)

    test('Search context selected based on URL', async () => {
        testContext.overrideGraphQL({
            ...testContextForSearchContexts,
            IsSearchContextAvailable: () => ({
                isSearchContextAvailable: true,
            }),
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=context:%40test+test&patternType=regexp', {
            waitUntil: 'networkidle0',
        })
        await driver.page.waitForSelector('.test-selected-search-context-spec', { visible: true })
        expect(await getSelectedSearchContextSpec()).toStrictEqual('context:@test')
    })

    test('Missing context filter should default to global context, even if another default is set', async () => {
        testContext.overrideGraphQL({
            ...testContextForSearchContexts,
            DefaultSearchContextSpec: () => ({
                defaultSearchContext: {
                    __typename: 'SearchContext',
                    spec: 'ctx-1',
                },
            }),
        })

        // Visit the search page with a query without a context filter
        await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test', {
            waitUntil: 'networkidle0',
        })
        await driver.page.waitForSelector('.test-selected-search-context-spec', { visible: true })
        expect(await getSelectedSearchContextSpec()).toStrictEqual('context:global')
    })

    test('Unavailable search context should remain in the query and disable the search context dropdown with default context', async () => {
        const { waitForInput, applySettings } = getSearchQueryInputConfig('codemirror6')

        testContext.overrideGraphQL({
            ...testContextForSearchContexts,
            ...createViewerSettingsGraphQLOverride({
                user: applySettings({
                    experimentalFeatures: {
                        showSearchContext: true,
                    },
                }),
            }),
            DefaultSearchContextSpec: () => ({
                defaultSearchContext: {
                    __typename: 'SearchContext',
                    spec: 'ctx-1',
                },
            }),
        })

        await driver.page.goto(
            driver.sourcegraphBaseUrl + '/search?q=context:%40unavailableCtx+test&patternType=regexp',
            { waitUntil: 'networkidle0' }
        )
        await driver.page.waitForSelector('.test-selected-search-context-spec', { visible: true })
        const editor = await waitForInput(driver, '[data-testid="searchbox"] .test-query-input')
        expect(await editor.getValue()).toStrictEqual('context:@unavailableCtx test')
        expect(await isSearchContextDropdownDisabled()).toBeTruthy()
        expect(await getSelectedSearchContextSpec()).toStrictEqual('context:ctx-1')
    })

    test('Reset unavailable search context from default if query is not present', async () => {
        testContext.overrideGraphQL({
            ...testContextForSearchContexts,
            DefaultSearchContextSpec: () => ({
                defaultSearchContext: null,
            }),
        })
        await driver.page.goto(driver.sourcegraphBaseUrl + '/search', {
            waitUntil: 'networkidle0',
        })
        await driver.page.waitForSelector('.test-selected-search-context-spec', { visible: true })
        expect(await getSelectedSearchContextSpec()).toStrictEqual('context:global')
    })

    test('Create static search context', async () => {
        testContext.overrideGraphQL({
            ...testContextForSearchContexts,
            RepositoriesByNames: ({ names }) => ({
                repositories: { nodes: names.map((name, index) => ({ id: `index-${index}`, name })) },
            }),
            CreateSearchContext: ({ searchContext, repositories }) => ({
                createSearchContext: {
                    __typename: 'SearchContext',
                    id: 'id1',
                    spec: searchContext.name,
                    name: searchContext.name,
                    namespace: null,
                    description: searchContext.description,
                    public: searchContext.public,
                    autoDefined: false,
                    updatedAt: '',
                    viewerCanManage: true,
                    viewerHasAsDefault: false,
                    viewerHasStarred: false,
                    query: searchContext.query,
                    repositories: repositories.map(repository => ({
                        __typename: 'SearchContextRepositoryRevisions',
                        revisions: repository.revisions,
                        repository: { name: repository.repositoryID },
                    })),
                },
            }),
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/contexts/new')

        await driver.replaceText({
            selector: '[data-testid="search-context-name-input"]',
            newText: 'new-context',
            enterTextMethod: 'type',
        })

        // Assert spec preview
        const specPreview = await driver.page.evaluate(
            () => document.querySelector('[data-testid="search-context-preview"]')?.textContent
        )
        expect(specPreview).toBe('context:@test/new-context')

        // Enter description
        await driver.replaceText({
            selector: '[data-testid="search-context-description-input"]',
            newText: 'Search context description',
            enterTextMethod: 'type',
        })

        // Select JSON config option
        await driver.page.click('#search-context-type-static')

        // Enter repositories
        const repositoriesConfig =
            '[{ "repository": "github.com/example/example", "revisions": ["main", "pr/feature1"] }]'
        {
            const editor = await createEditorAPI(driver, '[data-testid="repositories-config-area"] .test-editor')
            await editor.replace(repositoriesConfig, 'paste')
        }

        // Test configuration
        await driver.page.click('[data-testid="repositories-config-button"]')
        await driver.page.waitForSelector(
            '[data-testid="repositories-config-button"] [data-testid="repositories-config-success"]'
        )

        // Take Snapshot
        await percySnapshotWithVariants(driver.page, 'Create static search context page')
        await accessibilityAudit(driver.page)
        // Click create
        await driver.page.click('[data-testid="search-context-submit-button"]')

        // Wait for submit request to finish and redirect to list page
        await driver.page.waitForSelector('[data-testid="search-contexts-list-page"]')
    })

    test('Create dynamic query search context', async () => {
        testContext.overrideGraphQL({
            ...testContextForSearchContexts,
            CreateSearchContext: ({ searchContext, repositories }) => ({
                createSearchContext: {
                    __typename: 'SearchContext',
                    id: 'id1',
                    spec: searchContext.name,
                    name: searchContext.name,
                    namespace: null,
                    description: searchContext.description,
                    public: searchContext.public,
                    autoDefined: false,
                    updatedAt: '',
                    viewerCanManage: true,
                    viewerHasAsDefault: false,
                    viewerHasStarred: false,
                    query: searchContext.query,
                    repositories: repositories.map(repository => ({
                        __typename: 'SearchContextRepositoryRevisions',
                        revisions: repository.revisions,
                        repository: { name: repository.repositoryID },
                    })),
                },
            }),
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/contexts/new', {
            waitUntil: 'networkidle0',
        })

        await driver.replaceText({
            selector: '[data-testid="search-context-name-input"]',
            newText: 'new-context',
            enterTextMethod: 'type',
        })

        // Assert spec preview
        const specPreview = await driver.page.evaluate(
            () => document.querySelector('[data-testid="search-context-preview"]')?.textContent
        )
        expect(specPreview).toBe('context:@test/new-context')

        // Enter description
        await driver.replaceText({
            selector: '[data-testid="search-context-description-input"]',
            newText: 'Search context description',
            enterTextMethod: 'type',
        })

        // Select query option
        await driver.page.click('#search-context-type-dynamic')

        // Wait for search query input
        const editor = await createEditorAPI(driver, '[data-testid="search-context-dynamic-query"] .test-query-input')
        await editor.focus()

        // Take Snapshot
        await percySnapshotWithVariants(driver.page, 'Create dynamic query search context page')

        // Enter search query
        await editor.replace('repo:abc')

        await accessibilityAudit(driver.page)
        // Click create
        await driver.page.click('[data-testid="search-context-submit-button"]')

        // Wait for submit request to finish and redirect to list page
        await driver.page.waitForSelector('[data-testid="search-contexts-list-page"]')
    })

    test('Edit search context', async () => {
        testContext.overrideGraphQL({
            ...testContextForSearchContexts,
            RepositoriesByNames: ({ names }) => ({
                repositories: { nodes: names.map((name, index) => ({ id: `index-${index}`, name })) },
            }),
            UpdateSearchContext: ({ id, searchContext, repositories }) => ({
                updateSearchContext: {
                    __typename: 'SearchContext',
                    id,
                    spec: `@test/${searchContext.name}`,
                    name: searchContext.name,
                    namespace: {
                        __typename: 'User',
                        id: 'u1',
                        namespaceName: 'test',
                    },
                    description: searchContext.description,
                    public: searchContext.public,
                    autoDefined: false,
                    updatedAt: subDays(new Date(), 1).toISOString(),
                    viewerCanManage: true,
                    viewerHasAsDefault: false,
                    viewerHasStarred: false,
                    query: '',
                    repositories: repositories.map(repository => ({
                        __typename: 'SearchContextRepositoryRevisions',
                        revisions: repository.revisions,
                        repository: { name: repository.repositoryID },
                    })),
                },
            }),
            FetchSearchContextBySpec: ({ spec }) => ({
                searchContextBySpec: {
                    __typename: 'SearchContext',
                    id: spec,
                    spec,
                    name: 'context-1',
                    namespace: {
                        __typename: 'User',
                        id: 'u1',
                        namespaceName: 'test',
                    },
                    description: 'description',
                    public: true,
                    autoDefined: false,
                    updatedAt: subDays(new Date(), 1).toISOString(),
                    viewerCanManage: true,
                    viewerHasAsDefault: false,
                    viewerHasStarred: false,
                    query: '',
                    repositories: [
                        {
                            __typename: 'SearchContextRepositoryRevisions',
                            revisions: ['HEAD'],
                            repository: { name: 'github.com/example/example' },
                        },
                    ],
                },
            }),
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/contexts/context-1')

        await driver.page.waitForSelector('[data-testid="edit-search-context-link"]')
        await driver.page.click('[data-testid="edit-search-context-link"]')

        await driver.page.waitForSelector('[data-testid="search-context-name-input"]')
        await driver.replaceText({
            selector: '[data-testid="search-context-name-input"]',
            newText: 'new-context',
            enterTextMethod: 'type',
        })

        // Assert spec preview
        const specPreview = await driver.page.evaluate(
            () => document.querySelector('[data-testid="search-context-preview"]')?.textContent
        )
        expect(specPreview).toBe('context:@test/new-context')

        // Enter description
        await driver.replaceText({
            selector: '[data-testid="search-context-description-input"]',
            newText: 'Search context description',
            enterTextMethod: 'type',
        })

        // Enter repositories
        const repositoriesConfig = '[{ "repository": "github.com/example/example", "revisions": ["main"] }]'
        const editor = await createEditorAPI(driver, '[data-testid="repositories-config-area"] .test-editor')
        await editor.replace(repositoriesConfig, 'paste')

        // Test configuration
        await driver.page.click('[data-testid="repositories-config-button"]')
        await driver.page.waitForSelector(
            '[data-testid="repositories-config-button"] [data-testid="repositories-config-success"]'
        )

        // Click save
        await driver.page.click('[data-testid="search-context-submit-button"]')

        // Wait for submit request to finish and redirect to list page
        await driver.page.waitForSelector('[data-testid="search-contexts-list-page"]')
    })

    test('Cannot edit search context without necessary permissions', async () => {
        testContext.overrideGraphQL({
            ...testContextForSearchContexts,
            FetchSearchContextBySpec: ({ spec }) => ({
                searchContextBySpec: {
                    __typename: 'SearchContext',
                    id: spec,
                    spec,
                    name: 'context-1',
                    namespace: null,
                    description: 'description',
                    public: true,
                    autoDefined: false,
                    updatedAt: subDays(new Date(), 1).toISOString(),
                    viewerCanManage: false,
                    viewerHasAsDefault: false,
                    viewerHasStarred: false,
                    query: '',
                    repositories: [],
                },
            }),
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/contexts/context-1/edit')

        await driver.page.waitForSelector('[data-testid="search-contexts-alert-danger"]')
        const errorText = await driver.page.evaluate(
            () => document.querySelector('[data-testid="search-contexts-alert-danger"]')?.textContent
        )
        expect(errorText).toContain('You do not have sufficient permissions to edit this context.')
    })

    test('Delete search context', async () => {
        testContext.overrideGraphQL({
            ...testContextForSearchContexts,
            DeleteSearchContext: () => ({
                deleteSearchContext: {
                    alwaysNil: '',
                },
            }),
            FetchSearchContextBySpec: ({ spec }) => ({
                searchContextBySpec: {
                    __typename: 'SearchContext',
                    id: spec,
                    spec,
                    name: 'context-1',
                    namespace: {
                        __typename: 'User',
                        id: 'u1',
                        namespaceName: 'test',
                    },
                    description: 'description',
                    public: true,
                    autoDefined: false,
                    updatedAt: subDays(new Date(), 1).toISOString(),
                    viewerCanManage: true,
                    viewerHasAsDefault: false,
                    viewerHasStarred: false,
                    query: '',
                    repositories: [
                        {
                            __typename: 'SearchContextRepositoryRevisions',
                            revisions: ['HEAD'],
                            repository: { name: 'github.com/example/example' },
                        },
                    ],
                },
            }),
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/contexts/context-1/edit')

        // Click delete
        await driver.page.waitForSelector('[data-testid="search-context-delete-button"]')
        await driver.page.click('[data-testid="search-context-delete-button"]')

        // Wait for modal
        await driver.page.waitForSelector('[data-testid="delete-search-context-modal"]', { visible: true })
        await driver.page.click('[data-testid="confirm-delete-search-context"]')

        // Wait for delete request to finish and redirect to list page
        await driver.page.waitForSelector('[data-testid="search-contexts-list-page"]')
    })

    test('Infinite scrolling in dropdown menu', async () => {
        // We're loading 15 search contexts per page, and we want to load 2 pages
        const searchContextsCount = 30

        testContext.overrideGraphQL({
            ...testContextForSearchContexts,
            ListSearchContexts: ({ after }) => {
                const searchContexts = range(0, searchContextsCount).map(index => ({
                    __typename: 'SearchContext',
                    id: `id-${index}`,
                    spec: `ctx-${index}`,
                    name: `ctx-${index}`,
                    namespace: null,
                    public: true,
                    autoDefined: false,
                    viewerCanManage: false,
                    viewerHasAsDefault: false,
                    viewerHasStarred: false,
                    description: '',
                    repositories: [],
                    query: '',
                    updatedAt: subDays(new Date(), 1).toISOString(),
                })) as SearchContextMinimalFields[]

                if (after === null) {
                    return {
                        searchContexts: {
                            nodes: searchContexts.slice(0, searchContextsCount / 2),
                            totalCount: searchContexts.length,
                            pageInfo: {
                                hasNextPage: true,
                                endCursor: 'end-first-page',
                            },
                        },
                    }
                }

                return {
                    searchContexts: {
                        nodes: searchContexts.slice(searchContextsCount / 2),
                        totalCount: searchContexts.length,
                        pageInfo: {
                            hasNextPage: false,
                            endCursor: null,
                        },
                    },
                }
            },
        })

        // Go to search homepage and wait for context selector to load
        await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
        await driver.page.waitForSelector('.test-search-context-dropdown', { visible: true })

        // Open dropdown menu
        await driver.page.click('.test-search-context-dropdown')
        await driver.page.waitForSelector('[data-testid="search-context-menu-item"]', { visible: true })

        // Scroll to the bottom of the list
        await driver.page.evaluate(() => {
            const scrollableSection = document.querySelector<HTMLDivElement>('[data-testid="search-context-menu-list"]')
            if (scrollableSection) {
                scrollableSection.scrollTop = scrollableSection.offsetHeight
            }
        })

        // Wait for correct number of total elements to load
        await driver.page.waitForFunction(
            (searchContextsCount: number) =>
                document.querySelectorAll('[data-testid="search-context-menu-item-name"]').length ===
                searchContextsCount,
            {},
            searchContextsCount
        )

        await percySnapshotWithVariants(driver.page, 'Search contexts list page')
        await accessibilityAudit(driver.page)
    })

    test('Switching contexts with empty query', async () => {
        testContext.overrideGraphQL({
            ...testContextForSearchContexts,
            IsSearchContextAvailable: () => ({
                isSearchContextAvailable: true,
            }),
            ListSearchContexts: () => {
                const nodes = range(0, 2).map(index => ({
                    __typename: 'SearchContext',
                    id: `id-${index}`,
                    spec: `ctx-${index}`,
                    name: `ctx-${index}`,
                    namespace: null,
                    public: true,
                    autoDefined: false,
                    viewerCanManage: false,
                    viewerHasAsDefault: false,
                    viewerHasStarred: false,
                    description: '',
                    repositories: [],
                    query: '',
                    updatedAt: subDays(new Date(), 1).toISOString(),
                })) as SearchContextMinimalFields[]

                return {
                    searchContexts: {
                        nodes,
                        totalCount: nodes.length,
                        pageInfo: {
                            hasNextPage: false,
                            endCursor: null,
                        },
                    },
                }
            },
        })

        // Go to search results page with a single context filter in the query and wait for context selector to load
        await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=context:ctx-0&patternType=literal')
        await driver.page.waitForSelector('.test-search-context-dropdown', { visible: true })

        // Open dropdown menu
        await driver.page.click('.test-search-context-dropdown')
        await driver.page.waitForSelector('[data-testid="search-context-menu-item-name"]', { visible: true })

        await Promise.all([
            // A search will be submitted on context click, wait for the navigation
            driver.page.waitForNavigation({ waitUntil: 'networkidle0' }),
            // Click second context item in the dropdown
            driver.page.click(
                '[data-testid="search-context-menu-item"]:nth-child(2) [data-testid="search-context-menu-item-name"]'
            ),
        ])

        await driver.page.waitForSelector('.test-search-context-dropdown', { visible: true })
        // The context should have been switched
        expect(await getSelectedSearchContextSpec()).toStrictEqual('context:ctx-1')
    })
})
