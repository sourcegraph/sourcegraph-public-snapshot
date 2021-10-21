import { subDays } from 'date-fns'
import expect from 'expect'
import { range } from 'lodash'
import { test } from 'mocha'

import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { ISearchContext } from '@sourcegraph/shared/src/graphql/schema'
import { Driver, createDriverForTest } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { WebGraphQlOperations } from '../graphql-operations'

import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import { createRepositoryRedirectResult } from './graphQlResponseHelpers'
import { commonWebGraphQlResults } from './graphQlResults'
import { siteGQLID, siteID } from './jscontext'

const commonSearchGraphQLResults: Partial<WebGraphQlOperations & SharedGraphQlOperations> = {
    ...commonWebGraphQlResults,
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
        testContext.overrideSearchStreamEvents([{ type: 'done', data: {} }])
    })
    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    const getSearchFieldValue = (driver: Driver): Promise<string | undefined> =>
        driver.page.evaluate(() => document.querySelector<HTMLTextAreaElement>('#monaco-query-input textarea')?.value)

    const viewerSettingsWithSearchContexts: Partial<WebGraphQlOperations> = {
        ViewerSettings: () => ({
            viewerSettings: {
                __typename: 'SettingsCascade',
                subjects: [
                    {
                        __typename: 'DefaultSettings',
                        settingsURL: null,
                        viewerCanAdminister: false,
                        latestSettings: {
                            id: 0,
                            contents: JSON.stringify({
                                experimentalFeatures: {
                                    showSearchContext: true,
                                    showSearchContextManagement: true,
                                },
                            }),
                        },
                    },
                    {
                        __typename: 'Site',
                        id: siteGQLID,
                        siteID,
                        latestSettings: {
                            id: 470,
                            contents: JSON.stringify({
                                experimentalFeatures: {
                                    showSearchContext: true,
                                    showSearchContextManagement: true,
                                },
                            }),
                        },
                        settingsURL: '/site-admin/global-settings',
                        viewerCanAdminister: true,
                        allowSiteSettingsEdits: true,
                    },
                ],
                final: JSON.stringify({}),
            },
        }),
    }

    const testContextForSearchContexts: Partial<WebGraphQlOperations> = {
        ...commonSearchGraphQLResults,
        ...viewerSettingsWithSearchContexts,
        UserRepositories: () => ({
            node: {
                __typename: 'User',
                repositories: {
                    totalCount: 1,
                    nodes: [
                        {
                            id: '1',
                            name: 'repo',
                            viewerCanAdminister: false,
                            createdAt: '',
                            url: '',
                            isPrivate: false,
                            mirrorInfo: { cloned: true, cloneInProgress: false, updatedAt: null },
                            externalRepository: { serviceType: '', serviceID: '' },
                        },
                    ],
                    pageInfo: { hasNextPage: false },
                },
            },
        }),
    }

    const getSelectedSearchContextSpec = () =>
        driver.page.evaluate(() => document.querySelector('.test-selected-search-context-spec')?.textContent)

    const isSearchContextDropdownDisabled = () =>
        driver.page.evaluate(() => document.querySelector<HTMLButtonElement>('.test-search-context-dropdown')?.disabled)

    const clearLocalStorage = () => driver.page.evaluate(() => localStorage.clear())

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

    test('Missing context filter should default to global context', async () => {
        // Initialize localStorage to a valid context, that should not be used
        await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
        await driver.page.evaluate(() => localStorage.setItem('sg-last-search-context', '@test'))
        // Visit the search page with a query without a context filter
        await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
        await driver.page.waitForSelector('.test-selected-search-context-spec', { visible: true })
        expect(await getSelectedSearchContextSpec()).toStrictEqual('context:global')
        await clearLocalStorage()
    })

    test('Unavailable search context should remain in the query and disable the search context dropdown', async () => {
        await driver.page.goto(
            driver.sourcegraphBaseUrl + '/search?q=context:%40unavailableCtx+test&patternType=regexp',
            { waitUntil: 'networkidle0' }
        )
        await driver.page.waitForSelector('.test-selected-search-context-spec', { visible: true })
        await driver.page.waitForSelector('#monaco-query-input')
        expect(await getSearchFieldValue(driver)).toStrictEqual('context:@unavailableCtx test')
        expect(await isSearchContextDropdownDisabled()).toBeTruthy()
    })

    test('Reset unavailable search context from localStorage if query is not present', async () => {
        // First initialize localStorage on the page
        await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
        await driver.page.evaluate(() => localStorage.setItem('sg-last-search-context', 'doesnotexist'))
        // Visit the page again with localStorage initialized
        await driver.page.goto(driver.sourcegraphBaseUrl + '/search', {
            waitUntil: 'networkidle0',
        })
        await driver.page.waitForSelector('.test-selected-search-context-spec', { visible: true })
        expect(await getSelectedSearchContextSpec()).toStrictEqual('context:global')
        await clearLocalStorage()
    })

    test('Create search context', async () => {
        testContext.overrideGraphQL({
            ...testContextForSearchContexts,
            RepositoryRedirect: ({ repoName }) => createRepositoryRedirectResult(repoName),
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

        // Enter repositories
        const repositoriesConfig =
            '[{ "repository": "github.com/example/example", "revisions": ["main", "pr/feature1"] }]'
        await driver.page.waitForSelector('[data-testid="repositories-config-area"] .monaco-editor')
        await driver.replaceText({
            selector: '[data-testid="repositories-config-area"] .monaco-editor',
            newText: repositoriesConfig,
            selectMethod: 'keyboard',
            enterTextMethod: 'paste',
        })

        // Test configuration
        await driver.page.click('[data-testid="repositories-config-button"]')
        await driver.page.waitForSelector('[data-testid="repositories-config-button"] .text-success')

        // Click create
        await driver.page.click('[data-testid="search-context-submit-button"]')

        // Wait for submit request to finish and redirect to list page
        await driver.page.waitForSelector('.search-contexts-list-page')
    })

    test('Edit search context', async () => {
        testContext.overrideGraphQL({
            ...testContextForSearchContexts,
            RepositoryRedirect: ({ repoName }) => createRepositoryRedirectResult(repoName),
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
        const repositoriesConfig =
            '[{ "repository": "github.com/example/example", "revisions": ["main", "pr/feature1"] }]'
        await driver.page.waitForSelector('[data-testid="repositories-config-area"] .monaco-editor')
        await driver.replaceText({
            selector: '[data-testid="repositories-config-area"] .monaco-editor',
            newText: repositoriesConfig,
            selectMethod: 'keyboard',
            enterTextMethod: 'paste',
        })

        // Test configuration
        await driver.page.click('[data-testid="repositories-config-button"]')
        await driver.page.waitForSelector('[data-testid="repositories-config-button"] .text-success')

        // Click save
        await driver.page.click('[data-testid="search-context-submit-button"]')

        // Wait for submit request to finish and redirect to list page
        await driver.page.waitForSelector('.search-contexts-list-page')
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
                    repositories: [],
                },
            }),
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/contexts/context-1/edit')

        await driver.page.waitForSelector('.alert-danger')
        const errorText = await driver.page.evaluate(() => document.querySelector('.alert-danger')?.textContent)
        expect(errorText).toContain('You do not have sufficient permissions to edit this context.')
    })

    test('Delete search context', async () => {
        testContext.overrideGraphQL({
            ...testContextForSearchContexts,
            RepositoryRedirect: ({ repoName }) => createRepositoryRedirectResult(repoName),
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
        await driver.page.waitForSelector('.search-contexts-list-page')
    })

    test('Infinite scrolling in dropdown menu', async () => {
        // We're loading 15 search contexts per page, and we want to load 2 pages
        const searchContextsCount = 30

        testContext.overrideGraphQL({
            ...testContextForSearchContexts,
            AutoDefinedSearchContexts: () => ({
                autoDefinedSearchContexts: [],
            }),
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
                    description: '',
                    repositories: [],
                    updatedAt: subDays(new Date(), 1).toISOString(),
                })) as ISearchContext[]

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
        await driver.page.waitFor(
            searchContextsCount =>
                document.querySelectorAll('[data-testid="search-context-menu-item-name"]').length ===
                searchContextsCount,
            {},
            searchContextsCount
        )
    })
})
