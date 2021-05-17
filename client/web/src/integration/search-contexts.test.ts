import expect from 'expect'
import { test } from 'mocha'

import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { Driver, createDriverForTest } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import {
    RepoGroupsResult,
    SearchSuggestionsResult,
    WebGraphQlOperations,
    AutoDefinedSearchContextsResult,
    SearchResult,
} from '../graphql-operations'

import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import { commonWebGraphQlResults } from './graphQlResults'
import { createJsContext, siteGQLID, siteID } from './jscontext'
import { createRepositoryRedirectResult } from './graphQlResponseHelpers'

const commonSearchGraphQLResults: Partial<WebGraphQlOperations & SharedGraphQlOperations> = {
    ...commonWebGraphQlResults,
    SearchSuggestions: (): SearchSuggestionsResult => ({
        search: {
            suggestions: [],
        },
    }),
    Search: (): SearchResult => ({
        search: {
            results: {
                __typename: 'SearchResults',
                limitHit: true,
                matchCount: 30,
                approximateResultCount: '30+',
                missing: [],
                cloning: [],
                repositoriesCount: 372,
                timedout: [],
                indexUnavailable: false,
                dynamicFilters: [],
                results: [],
                alert: null,
                elapsedMilliseconds: 103,
            },
        },
    }),
    RepoGroups: (): RepoGroupsResult => ({
        repoGroups: [],
    }),
    AutoDefinedSearchContexts: (): AutoDefinedSearchContextsResult => ({
        autoDefinedSearchContexts: [],
    }),
    ConvertVersionContextToSearchContext: ({ name }) => ({
        convertVersionContextToSearchContext: { id: `id${name}`, spec: name },
    }),
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
        const context = createJsContext({ sourcegraphBaseUrl: driver.sourcegraphBaseUrl })
        testContext.overrideJsContext({
            ...context,
            experimentalFeatures: {
                versionContexts,
            },
        })
    })
    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(async () => {
        testContext?.dispose()
        await driver.page.evaluate(() => localStorage.clear())
    })

    const getSearchFieldValue = (driver: Driver): Promise<string | undefined> =>
        driver.page.evaluate(() => document.querySelector<HTMLTextAreaElement>('#monaco-query-input textarea')?.value)

    const viewerSettingsWithSearchContexts: Partial<WebGraphQlOperations> = {
        ViewerSettings: () => ({
            viewerSettings: {
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
                    },
                ],
                final: JSON.stringify({}),
            },
        }),
    }

    const testContextForSearchContexts: Partial<WebGraphQlOperations> = {
        ...commonSearchGraphQLResults,
        ...viewerSettingsWithSearchContexts,
        AutoDefinedSearchContexts: () => ({
            autoDefinedSearchContexts: [
                {
                    __typename: 'SearchContext',
                    id: '1',
                    spec: 'global',
                    description: '',
                    autoDefined: true,
                    public: true,
                    updatedAt: '2021-03-15T19:39:11Z',
                    repositories: [],
                },
                {
                    __typename: 'SearchContext',
                    id: '2',
                    spec: '@test',
                    description: '',
                    autoDefined: true,
                    public: true,
                    updatedAt: '2021-03-15T19:39:11Z',
                    repositories: [],
                },
            ],
        }),
        UserRepositories: () => ({
            node: {
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
    const versionContexts = [
        {
            name: 'version-context-1',
            description: 'v1',
            revisions: [],
        },
        {
            name: 'version-context-2',
            description: 'v2',
            revisions: [],
        },
    ]

    const getSelectedSearchContextSpec = () =>
        driver.page.evaluate(() => document.querySelector('.test-selected-search-context-spec')?.textContent)

    const isSearchContextDropdownVisible = () =>
        driver.page.evaluate(() => document.querySelector<HTMLButtonElement>('.test-search-context-dropdown') !== null)

    const isSearchContextHighlightTourStepVisible = () =>
        driver.page.evaluate(
            () =>
                document.querySelector<HTMLDivElement>('div[data-shepherd-step-id="search-contexts-start-tour"]') !==
                null
        )

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

    test('Missing context filter should default to global context', async () => {
        // Initialize localStorage to a valid context, that should not be used
        await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
        await driver.page.evaluate(() => localStorage.setItem('sg-last-search-context', '@test'))
        // Visit the search page with a query without a context filter
        await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
        await driver.page.waitForSelector('.test-selected-search-context-spec', { visible: true })
        expect(await getSelectedSearchContextSpec()).toStrictEqual('context:global')
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

    test('Search context dropdown should not be visible if user has no repositories', async () => {
        testContext.overrideGraphQL({
            ...testContextForSearchContexts,
            UserRepositories: () => ({
                node: {
                    repositories: {
                        totalCount: 0,
                        nodes: [],
                        pageInfo: { hasNextPage: false },
                    },
                },
            }),
        })
        await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
        await driver.page.waitForSelector('#monaco-query-input')
        expect(await isSearchContextDropdownVisible()).toBeFalsy()
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
    })

    test('Disable dropdown if version context is active', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=test&patternType=regexp&c=version-context-1')
        await driver.page.waitForSelector('.test-selected-search-context-spec', { visible: true })
        expect(await isSearchContextDropdownDisabled()).toBeTruthy()
    })

    test('Convert version context', async () => {
        testContext.overrideGraphQL({
            ...testContextForSearchContexts,
            IsSearchContextAvailable: () => ({
                isSearchContextAvailable: false,
            }),
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/contexts/convert-version-contexts')

        await driver.page.waitForSelector('.test-convert-version-context-btn', { visible: true })
        await driver.page.click('.test-convert-version-context-btn')

        await driver.page.waitForSelector('.convert-version-context-node .text-success')

        const successText = await driver.page.evaluate(
            () => document.querySelector('.convert-version-context-node .text-success')?.textContent
        )
        expect(successText).toBe('Version context successfully converted.')
    })

    test('Convert all version contexts', async () => {
        testContext.overrideGraphQL({
            ...testContextForSearchContexts,
            IsSearchContextAvailable: () => ({
                isSearchContextAvailable: false,
            }),
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/contexts/convert-version-contexts')

        // Wait for individual nodes to load
        await driver.page.waitForSelector('.test-convert-version-context-btn', { visible: true })
        await driver.page.waitForSelector('.test-convert-all-search-contexts-btn')
        await driver.page.click('.test-convert-all-search-contexts-btn')

        testContext.overrideGraphQL({
            ...testContextForSearchContexts,
            IsSearchContextAvailable: () => ({
                isSearchContextAvailable: true,
            }),
        })

        // Check that a success message appears with the correct number of converted contexts
        await driver.page.waitForSelector('.test-convert-all-search-contexts-success')
        const successText = await driver.page.evaluate(
            () => document.querySelector('.test-convert-all-search-contexts-success')?.textContent
        )
        expect(successText).toBe(
            `Sucessfully converted ${versionContexts.length} version contexts into search contexts.`
        )

        // Check that individual context nodes have 'Converted' text
        const convertedContexts = await driver.page.evaluate(
            () => document.querySelectorAll('.test-converted-context').length
        )
        expect(convertedContexts).toBe(versionContexts.length)
    })

    test('Highlight tour step should be visible with empty local storage', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=context:global+test&patternType=regexp')
        await driver.page.waitForSelector('.test-selected-search-context-spec', { visible: true })
        expect(await isSearchContextHighlightTourStepVisible()).toBeTruthy()
    })

    test('Highlight tour step should not be visible if already seen', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=context:global+test&patternType=regexp', {
            waitUntil: 'networkidle0',
        })
        await driver.page.evaluate(() =>
            localStorage.setItem('has-seen-search-contexts-dropdown-highlight-tour-step', 'true')
        )
        await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=context:global+test&patternType=regexp')
        await driver.page.waitForSelector('.test-selected-search-context-spec', { visible: true })
        expect(await isSearchContextHighlightTourStepVisible()).toBeFalsy()
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
                    description: searchContext.description,
                    public: searchContext.public,
                    autoDefined: false,
                    updatedAt: '',
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
            selector: '.test-search-context-name-input',
            newText: 'new-context',
            enterTextMethod: 'type',
        })

        // Assert spec preview
        const specPreview = await driver.page.evaluate(
            () => document.querySelector('.test-search-context-preview')?.textContent
        )
        expect(specPreview).toBe('context:@test/new-context')

        // Enter description
        await driver.replaceText({
            selector: '.test-search-context-description-input',
            newText: 'Search context description',
            enterTextMethod: 'type',
        })

        // Enter repositories
        const repositoriesConfig = `[{ "repository": "github.com/example/example", "revisions": ["main", "pr/feature1"] }]`
        await driver.page.waitForSelector('.test-repositories-config-input .monaco-editor')
        await driver.replaceText({
            selector: '.test-repositories-config-input .monaco-editor',
            newText: repositoriesConfig,
            selectMethod: 'keyboard',
            enterTextMethod: 'paste',
        })

        // Test configuration
        await driver.page.click('.test-repositories-config-button')
        await driver.page.waitForSelector('.test-repositories-config-button .text-success')

        // Click create
        await driver.page.click('.test-create-search-context-button')

        // Wait for submit request to finish and redirect to list page
        await driver.page.waitForSelector('.search-contexts-list-page')
    })
})
