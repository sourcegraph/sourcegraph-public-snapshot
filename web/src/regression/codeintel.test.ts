import expect from 'expect'
import { describe, before, after, test } from 'mocha'
import * as path from 'path'
import { range } from 'lodash'
import { ElementHandle } from 'puppeteer'
import { map } from 'rxjs/operators'
import * as child_process from 'mz/child_process'
import { applyEdits } from '@sqs/jsonc-parser'
import { JSONPath } from '@sqs/jsonc-parser/lib/main'
import { setProperty } from '@sqs/jsonc-parser/lib/edit'
import { getTestTools } from './util/init'
import { GraphQLClient } from './util/GraphQLClient'
import { TestResourceManager } from './util/TestResourceManager'
import { ensureTestExternalService, getUser, setUserSiteAdmin } from './util/api'
import { ensureLoggedInOrCreateTestUser, getGlobalSettings } from './util/helpers'
import * as GQL from '../../../shared/src/graphql/schema'
import { Driver } from '../../../shared/src/e2e/driver'
import { Config, getConfig } from '../../../shared/src/e2e/config'
import { dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import { overwriteSettings } from '../../../shared/src/settings/edit'
import { saveScreenshotsUponFailures } from '../../../shared/src/e2e/screenshotReporter'
import { asError } from '../../../shared/src/util/errors'

describe('Code intelligence regression test suite', () => {
    const testUsername = 'test-sg-codeintel'
    const config = getConfig(
        'gitHubToken',
        'headless',
        'keepBrowser',
        'logBrowserConsole',
        'logStatusMessages',
        'noCleanup',
        'slowMo',
        'sourcegraphBaseUrl',
        'sudoToken',
        'sudoUsername',
        'testUserPassword'
    )
    const testExternalServiceInfo = {
        kind: GQL.ExternalServiceKind.GITHUB,
        uniqueDisplayName: '[TEST] GitHub (codeintel.test.ts)',
    }

    const testRepoSlugs = [
        'sourcegraph/sourcegraph',
        'sourcegraph-testing/prometheus-common',
        'sourcegraph-testing/prometheus-client-golang',
        'sourcegraph-testing/prometheus-redefinitions',
    ]

    const prometheusCommonHeadCommit = 'b5fe7d854c42dc7842e48d1ca58f60feae09d77b' // HEAD
    const prometheusCommonLSIFCommit = '287d3e634a1e550c9e463dd7e5a75a422c614505' // 2 behind HEAD
    const prometheusCommonFallbackCommit = 'e8215224146358493faab0295ce364cd386223b9' // 2 behind LSIF
    const prometheusClientHeadCommit = '333f01cef0d61f9ef05ada3d94e00e69c8d5cdda'
    const prometheusRedefinitionsHeadCommit = 'c68f0e063cf8a98e7ce3428cfd50588746010f1f'

    const prometheusCommonSamplePairLocations = [
        { path: 'model/value.go', line: 31, character: 19 },
        { path: 'model/value.go', line: 78, character: 6 },
        { path: 'model/value.go', line: 84, character: 9 },
        { path: 'model/value.go', line: 97, character: 10 },
        { path: 'model/value.go', line: 104, character: 10 },
        { path: 'model/value.go', line: 108, character: 9 },
        { path: 'model/value.go', line: 137, character: 43 },
        { path: 'model/value.go', line: 147, character: 10 },
        { path: 'model/value.go', line: 163, character: 10 },
        { path: 'model/value.go', line: 225, character: 11 },
        { path: 'model/value_test.go', line: 133, character: 9 },
        { path: 'model/value_test.go', line: 137, character: 11 },
        { path: 'model/value_test.go', line: 156, character: 10 },
    ]

    const prometheusClientSamplePairLocations = [
        { path: 'api/prometheus/v1/api.go', line: 41, character: 15 },
        { path: 'api/prometheus/v1/api.go', line: 70, character: 17 },
        { path: 'api/prometheus/v1/api_test.go', line: 1119, character: 18 },
        { path: 'api/prometheus/v1/api_test.go', line: 1123, character: 20 },
        { path: 'api/prometheus/v1/api_test.go', line: 1127, character: 20 },
        { path: 'api/prometheus/v1/api_test.go', line: 1131, character: 20 },
        { path: 'api/prometheus/v1/api_test.go', line: 1135, character: 20 },
        { path: 'api/prometheus/v1/api_test.go', line: 1139, character: 20 },
        { path: 'api/prometheus/v1/api_test.go', line: 1143, character: 20 },
        { path: 'api/prometheus/v1/api_test.go', line: 1147, character: 20 },
        { path: 'api/prometheus/v1/api_test.go', line: 1151, character: 20 },
        { path: 'api/prometheus/v1/api_test.go', line: 1155, character: 20 },
        { path: 'api/prometheus/v1/api_test.go', line: 1159, character: 20 },
        { path: 'api/prometheus/v1/api_test.go', line: 1163, character: 20 },
        { path: 'api/prometheus/v1/api_test.go', line: 1167, character: 20 },
        { path: 'api/prometheus/v1/api_test.go', line: 1171, character: 20 },
        { path: 'api/prometheus/v1/api_test.go', line: 1175, character: 20 },
        { path: 'api/prometheus/v1/api_test.go', line: 1179, character: 20 },
        { path: 'api/prometheus/v1/api_test.go', line: 1197, character: 17 },
        { path: 'api/prometheus/v1/api_bench_test.go', line: 34, character: 26 },
    ]

    const prometheusRedefinitionSamplePairLocations = [
        { path: 'sample.go', line: 7, character: 6 },
        { path: 'sample.go', line: 12, character: 10 },
        { path: 'sample.go', line: 16, character: 10 },
    ]

    let driver: Driver
    let gqlClient: GraphQLClient
    let outerResourceManager: TestResourceManager
    before(async function() {
        // sourcegraph/sourcegraph takes a while to clone
        this.timeout(30 * 1000)
        ;({ driver, gqlClient, resourceManager: outerResourceManager } = await getTestTools(config))
        outerResourceManager.add(
            'User',
            testUsername,
            await ensureLoggedInOrCreateTestUser(driver, gqlClient, {
                username: testUsername,
                deleteIfExists: true,
                ...config,
            })
        )
        outerResourceManager.add(
            'External service',
            testExternalServiceInfo.uniqueDisplayName,
            await ensureTestExternalService(
                gqlClient,
                {
                    ...testExternalServiceInfo,
                    config: {
                        url: 'https://github.com',
                        token: config.gitHubToken,
                        repos: testRepoSlugs,
                        repositoryQuery: ['none'],
                    },
                    waitForRepos: testRepoSlugs.map(r => `github.com/${r}`),
                },
                config
            )
        )

        const user = await getUser(gqlClient, testUsername)
        if (!user) {
            throw new Error(`test user ${testUsername} does not exist`)
        }
        await setUserSiteAdmin(gqlClient, user.id, true)

        outerResourceManager.add('Global setting', 'showBadgeAttachments', await enableBadgeAttachments(gqlClient))
        outerResourceManager.add('Global setting', 'codeIntel.includeForks', await setIncludeForks(gqlClient, true))
    })

    saveScreenshotsUponFailures(() => driver.page)

    after(async () => {
        if (!config.noCleanup) {
            await outerResourceManager.destroyAll()
        }
        if (driver) {
            await driver.close()
        }
    })

    describe('Basic code intelligence regression test suite', () => {
        const innerResourceManager = new TestResourceManager()
        before(async () => {
            innerResourceManager.add('Global setting', 'codeIntel.lsif', await setGlobalLSIFSetting(gqlClient, false))
        })
        after(async () => {
            if (!config.noCleanup) {
                await innerResourceManager.destroyAll()
            }
        })

        test('Definitions, references, and hovers', () =>
            testCodeNavigation(driver, config, {
                page: `/github.com/sourcegraph-testing/prometheus-client-golang@${prometheusClientHeadCommit}/-/blob/api/prometheus/v1/api.go`,
                line: 41,
                token: 'SamplePair',
                precise: false,
                expectedHoverContains: 'SamplePair pairs a SampleValue with a Timestamp.',
                expectedDefinition: [
                    {
                        url: `/github.com/sourcegraph-testing/prometheus-common@${prometheusCommonHeadCommit}/-/blob/model/value.go#L78:1`,
                        precise: false,
                    },
                    {
                        url: `/github.com/sourcegraph-testing/prometheus-redefinitions@${prometheusRedefinitionsHeadCommit}/-/blob/sample.go#L7:1`,
                        precise: false,
                    },
                ],
                expectedReferences: [],
            }))

        test('File sidebar, multiple levels of directories', async () => {
            await driver.page.goto(
                config.sourcegraphBaseUrl +
                    '/github.com/sourcegraph/sourcegraph@c543dfd3936019befe94b881ade89e637d1a3dc3'
            )
            for (const file of ['cmd', 'frontend', 'auth', 'providers', 'providers.go']) {
                await driver.findElementWithText(file, {
                    action: 'click',
                    selector: '.e2e-repo-rev-sidebar a',
                    wait: { timeout: 2 * 1000 },
                })
            }
            await driver.waitUntilURL(
                `${config.sourcegraphBaseUrl}/github.com/sourcegraph/sourcegraph@c543dfd3936019befe94b881ade89e637d1a3dc3/-/blob/cmd/frontend/auth/providers/providers.go`,
                { timeout: 2 * 1000 }
            )
        })

        test('Symbols sidebar', async () => {
            await driver.page.goto(
                config.sourcegraphBaseUrl +
                    '/github.com/sourcegraph/sourcegraph@c543dfd3936019befe94b881ade89e637d1a3dc3'
            )
            await driver.findElementWithText('SYMBOLS', {
                action: 'click',
                selector: '.e2e-repo-rev-sidebar button',
                wait: { timeout: 10 * 1000 },
            })
            await driver.findElementWithText('backgroundEntry', {
                action: 'click',
                selector: '.e2e-repo-rev-sidebar a span',
                wait: { timeout: 2 * 1000 },
            })
            await driver.replaceText({
                selector: 'input[placeholder="Search symbols..."]',
                newText: 'buildentry',
            })
            await driver.page.waitForFunction(
                () => {
                    const sidebar = document.querySelector<HTMLElement>('.e2e-repo-rev-sidebar')
                    return sidebar && !sidebar.innerText.includes('backgroundEntry')
                },
                {
                    timeout: 2 * 1000,
                }
            )
            await driver.findElementWithText('buildEntry', {
                action: 'click',
                selector: '.e2e-repo-rev-sidebar a span',
                wait: { timeout: 2 * 1000 },
            })
            await driver.waitUntilURL(
                `${config.sourcegraphBaseUrl}/github.com/sourcegraph/sourcegraph@c543dfd3936019befe94b881ade89e637d1a3dc3/-/blob/browser/config/webpack/base.config.ts#L6:7-6:17`,
                { timeout: 2 * 1000 }
            )
        })
    })

    describe('Precise code intelligence regression test suite', () => {
        const innerResourceManager = new TestResourceManager()
        before(async function() {
            this.timeout(30 * 1000)

            const repoCommits = [
                { repository: 'prometheus-common', commit: prometheusCommonLSIFCommit },
                { repository: 'prometheus-client-golang', commit: prometheusClientHeadCommit },
            ]

            for (const { repository } of repoCommits) {
                // First, remove all existing uploads for the repository
                await clearUploads(gqlClient, `github.com/sourcegraph-testing/${repository}`)
            }

            const uploadUrls = []
            for (const { repository, commit } of repoCommits) {
                // Upload each upload in parallel and get back the upload status URLs
                uploadUrls.push(
                    await performUpload(config, {
                        repository: `github.com/sourcegraph-testing/${repository}`,
                        commit,
                        root: '/',
                        filename: `lsif-data/github.com/sourcegraph-testing/${repository}@${commit.substring(
                            0,
                            12
                        )}.lsif`,
                    })
                )

                innerResourceManager.add('LSIF upload', `${repository} upload`, () =>
                    clearUploads(gqlClient, `github.com/sourcegraph-testing/${repository}`)
                )
            }

            for (const uploadUrl of uploadUrls) {
                // Check the upload status URLs to ensure that they succeed, then ensure
                // that they are all listed as one of the "active" uploads for that repo
                await ensureUpload(driver, uploadUrl)
            }

            await clearUploads(gqlClient, 'github.com/sourcegraph-testing/prometheus-redefinitions')
            innerResourceManager.add('Global setting', 'codeIntel.lsif', await setGlobalLSIFSetting(gqlClient, true))
        })
        after(async () => {
            if (!config.noCleanup) {
                await innerResourceManager.destroyAll()
            }
        })

        /**
         * Construct a code navigation test case that ensures that there are precise results from
         * the prometheus-common and prometheus-client-golang repositories and imprecise results from
         * the prometheus-redefinitions library.
         *
         * All of these tests deal with the same SamplePair struct.
         */
        const makeTestCase = (page: string, line: number): CodeNavigationTestCase => {
            const prometheusCommonPrefix = `/github.com/sourcegraph-testing/prometheus-common@${prometheusCommonLSIFCommit}/-/blob/`
            const prometheusCommonLocations = prometheusCommonSamplePairLocations.map(({ path, line, character }) => ({
                url: `${prometheusCommonPrefix}${path}#L${line}:${character}`,
                precise: true,
            }))

            const prometheusClientPrefix = `/github.com/sourcegraph-testing/prometheus-client-golang@${prometheusClientHeadCommit}/-/blob/`
            const prometheusClientLocations = prometheusClientSamplePairLocations.map(({ path, line, character }) => ({
                url: `${prometheusClientPrefix}${path}#L${line}:${character}`,
                precise: true,
            }))

            const prometheusRedefinitionPrefix = `/github.com/sourcegraph-testing/prometheus-redefinitions@${prometheusRedefinitionsHeadCommit}/-/blob/`
            const prometheusRedefinitionLocations = prometheusRedefinitionSamplePairLocations.map(
                ({ path, line, character }) => ({
                    url: `${prometheusRedefinitionPrefix}${path}#L${line}:${character}`,
                    precise: false,
                })
            )

            return {
                page,
                line,
                token: 'SamplePair',
                precise: true,
                expectedHoverContains: 'SamplePair pairs a SampleValue with a Timestamp.',
                expectedDefinition: {
                    url: `/github.com/sourcegraph-testing/prometheus-common@${prometheusCommonLSIFCommit}/-/blob/model/value.go#L78:6`,
                    precise: true,
                },
                expectedReferences: prometheusCommonLocations
                    .concat(prometheusClientLocations)
                    .concat(prometheusRedefinitionLocations),
            }
        }

        test('Cross-repository definitions, references, and hovers (from definition)', async () => {
            await testCodeNavigation(
                driver,
                config,
                makeTestCase(
                    `/github.com/sourcegraph-testing/prometheus-common@${prometheusCommonLSIFCommit}/-/blob/model/value.go`,
                    31
                )
            )
        })

        test('Cross-repository definitions, references, and hovers (from reference)', async () => {
            await testCodeNavigation(
                driver,
                config,
                makeTestCase(
                    `/github.com/sourcegraph-testing/prometheus-common@${prometheusCommonLSIFCommit}/-/blob/model/value_test.go`,
                    133
                )
            )
        })

        test('Cross-repository definitions, references, and hovers (from remote reference)', async () => {
            await testCodeNavigation(
                driver,
                config,
                makeTestCase(
                    `/github.com/sourcegraph-testing/prometheus-client-golang@${prometheusClientHeadCommit}/-/blob/api/prometheus/v1/api.go`,
                    41
                )
            )
        })

        test('Cross-repository definitions, references, and hovers (from old commit)', async () => {
            await testCodeNavigation(
                driver,
                config,
                makeTestCase(
                    `/github.com/sourcegraph-testing/prometheus-common@${prometheusCommonFallbackCommit}/-/blob/model/value.go`,
                    31
                )
            )
        })
    })
})

//
// Code navigation utilities

interface CodeNavigationTestCase {
    /**
     * The source page.
     */
    page: string

    /**
     * The source line.
     */
    line: number

    /**
     * The token to click. Should be unambiguous within this line for the test to succeed.
     */
    token: string

    /**
     * Whether or not definition/hover results are precise
     */
    precise: boolean

    /**
     * A substring of the expected hover text
     */
    expectedHoverContains: string

    /**
     * A locations (if unambiguous), or a subset of locations that must occur within the definitions panel.
     */
    expectedDefinition: TestLocation | TestLocation[]

    /**
     * A subset of locations that must occur within the references panel.
     */
    expectedReferences?: TestLocation[]
}

interface TestLocation {
    url: string

    /**
     * Whether or not this location should be accompanied by a UI badge indicating imprecise code intel. Precise = no badge.
     */
    precise: boolean
}

/**
 * Navigate to the given page and test the definitions, references, and hovers of the token
 * on the given line. Will ensure both hover and clicking the token produces the hover overlay.
 * Will check the precision indicator of the hoverlay and each file match in the definition
 * and reference panels. Will compare hover text. Will compare location of each file match or
 * the target of the page navigated to on jump-to-definition (in the case of a single definition).
 */
async function testCodeNavigation(
    driver: Driver,
    config: Pick<Config, 'sourcegraphBaseUrl'>,
    {
        page,
        line,
        token,
        precise,
        expectedHoverContains,
        expectedDefinition,
        expectedReferences,
    }: CodeNavigationTestCase
): Promise<void> {
    await driver.page.goto(config.sourcegraphBaseUrl + page)
    await driver.page.waitForSelector('.e2e-blob')
    const tokenEl = await findTokenElement(driver, line, token)

    // Check hover
    await tokenEl.hover()
    await waitForHover(driver, expectedHoverContains, precise)

    // Check click
    await clickOnEmptyPartOfCodeView(driver)
    await tokenEl.click()
    await waitForHover(driver, expectedHoverContains)

    // Find-references
    if (expectedReferences && expectedReferences.length > 0) {
        await clickOnEmptyPartOfCodeView(driver)
        await tokenEl.hover()
        await waitForHover(driver, expectedHoverContains)
        await (await driver.findElementWithText('Find references')).click()

        await driver.page.waitForSelector('.e2e-search-result')
        const refLinks = await collectLinks(driver)
        for (const expectedReference of expectedReferences) {
            expect(refLinks).toContainEqual(expectedReference)
        }
        await clickOnEmptyPartOfCodeView(driver)
    }

    // Go-to-definition
    await clickOnEmptyPartOfCodeView(driver)
    await tokenEl.hover()
    await waitForHover(driver, expectedHoverContains)
    await (await driver.findElementWithText('Go to definition')).click()

    if (Array.isArray(expectedDefinition)) {
        await driver.page.waitForSelector('.hierarchical-locations-view')
        const defLinks = await collectLinks(driver)
        for (const definition of expectedDefinition) {
            expect(defLinks).toContainEqual(definition)
        }
    } else {
        await driver.page.waitForFunction(
            defURL => document.location.href.endsWith(defURL),
            { timeout: 2000 },
            expectedDefinition.url
        )
        await driver.page.goBack()
    }

    await driver.page.keyboard.press('Escape')
}

/**
 * Return a list of locations (and their precision) that exist in the file list
 * panel. This will click on each repository and collect the visible links in a
 * sequence.
 */
async function collectLinks(driver: Driver): Promise<Set<TestLocation>> {
    await driver.page.waitForSelector('.e2e-loading-spinner', { hidden: true })

    const panelTabTitles = await getPanelTabTitles(driver)
    if (panelTabTitles.length === 0) {
        return new Set(await collectVisibleLinks(driver))
    }

    const links = new Set<TestLocation>()
    for (const title of panelTabTitles) {
        const tabElem = await driver.page.$$(`.e2e-hierarchical-locations-view-list span[title="${title}"]`)
        if (tabElem.length > 0) {
            await tabElem[0].click()
        }

        for (const link of await collectVisibleLinks(driver)) {
            links.add(link)
        }
    }

    return links
}

/**
 * Return the list of repository titles on the left-hand side of the definition or
 * reference result panel.
 */
async function getPanelTabTitles(driver: Driver): Promise<string[]> {
    return (
        await Promise.all(
            (await driver.page.$$('.hierarchical-locations-view > div:nth-child(1) span[title]')).map(e =>
                e.evaluate(e => e.getAttribute('title') || '')
            )
        )
    ).map(normalizeWhitespace)
}

/**
 * Return a list of locations (and their precision) that are current visible in a
 * file list panel. This may be definitions or references.
 */
function collectVisibleLinks(driver: Driver): Promise<TestLocation[]> {
    return driver.page.evaluate(() =>
        Array.from(document.querySelectorAll<HTMLElement>('.e2e-file-match-children-item-wrapper')).map(a => ({
            url: a.querySelector('.e2e-file-match-children-item')?.getAttribute('href') || '',
            precise: a.querySelector('.e2e-badge-row')?.childElementCount === 0,
        }))
    )
}

/**
 * Close any visible hover overlay.
 */
async function clickOnEmptyPartOfCodeView(driver: Driver): Promise<void> {
    await driver.page.click('.e2e-blob tr:nth-child(1) .line')
    await driver.page.waitForFunction(() => document.querySelectorAll('.e2e-tooltip-go-to-definition').length === 0)
}

/**
 * Find the element with the token text on the given line.
 *
 * Will close any toast so that the enture line is visible and will hover over the line
 * to ensure that the line is tokenized (as this is done on-demand).
 */
async function findTokenElement(driver: Driver, line: number, token: string): Promise<ElementHandle<Element>> {
    try {
        // If there's an open toast, close it. If the toast remains open and our target
        // identifier happens to be hidden by it, we won't be able to select the correct
        // token. This condition was reproducible in the code navigation test that searches
        // for the identifier `StdioLogger`.
        await driver.page.click('.e2e-close-toast')
    } catch (error) {
        // No toast open, this is fine
    }

    const selector = `.e2e-blob tr:nth-child(${line}) span`
    await driver.page.hover(selector)
    return driver.findElementWithText(token, { selector, fuzziness: 'exact' })
}

/**
 * Wait for the hover tooltip to become visible. Compare the visible text with the expected
 * contents (expected contents must be a substring of the visible contents). If precise is
 * supplied, ensure that the presence of the UI indicator matches this value.
 */
async function waitForHover(driver: Driver, expectedHoverContains: string, precise?: boolean): Promise<void> {
    await driver.page.waitForSelector('.e2e-tooltip-go-to-definition')
    await driver.page.waitForSelector('.e2e-tooltip-content')
    expect(normalizeWhitespace(await getTooltip(driver))).toContain(normalizeWhitespace(expectedHoverContains))

    if (precise !== undefined) {
        expect(
            await driver.page.evaluate(() => document.querySelectorAll<HTMLElement>('.e2e-hover-badge').length)
        ).toEqual(precise ? 0 : 1)
    }
}

/**
 * Return the currently visible hover text.
 */
async function getTooltip(driver: Driver): Promise<string> {
    return driver.page.evaluate(() => (document.querySelector('.e2e-tooltip-content') as HTMLElement).innerText)
}

/**
 * Collapse multiple spaces into one.
 */
function normalizeWhitespace(s: string): string {
    return s.replace(/\s+/g, ' ')
}

//
// LSIF utilities

/** Show badge attachments in the UI to distinguish precise and search-based results. */
async function enableBadgeAttachments(gqlClient: GraphQLClient): Promise<() => Promise<void>> {
    return writeSetting(gqlClient, ['experimentalFeatures', 'showBadgeAttachments'], true)
}

/** Replace the codeIntel.includeForks setting with the given value. */
async function setIncludeForks(gqlClient: GraphQLClient, enabled: boolean): Promise<() => Promise<void>> {
    return writeSetting(gqlClient, ['basicCodeIntel.includeForks'], enabled)
}
/** Replace the codeIntel.lsif setting with the given value. */
async function setGlobalLSIFSetting(gqlClient: GraphQLClient, enabled: boolean): Promise<() => Promise<void>> {
    return writeSetting(gqlClient, ['codeIntel.lsif'], enabled)
}

/**
 * Return a promise that updates the global settings to their original value. This return value
 * is suitable for use with the resource manager's destroy queue.
 */
async function writeSetting(gqlClient: GraphQLClient, path: JSONPath, value: unknown): Promise<() => Promise<void>> {
    const { subjectID, settingsID, contents: oldContents } = await getGlobalSettings(gqlClient)
    const newContents = applyEdits(
        oldContents,
        setProperty(oldContents, path, value, {
            eol: '\n',
            insertSpaces: true,
            tabSize: 2,
        })
    )

    await overwriteSettings(gqlClient, subjectID, settingsID, newContents)
    return async () => {
        const { subjectID: currentSubjectID, settingsID: currentSettingsID } = await getGlobalSettings(gqlClient)
        await overwriteSettings(gqlClient, currentSubjectID, currentSettingsID, oldContents)
    }
}

/**
 * Delete all LSIF uploads for a repository.
 */
async function clearUploads(gqlClient: GraphQLClient, repoName: string): Promise<void> {
    const { nodes, hasNextPage } = await gqlClient
        .queryGraphQL(
            gql`
                query ResolveRev($repoName: String!) {
                    repository(name: $repoName) {
                        lsifUploads {
                            nodes {
                                id
                            }

                            pageInfo {
                                hasNextPage
                            }
                        }
                    }
                }
            `,
            { repoName }
        )
        .pipe(
            map(dataOrThrowErrors),
            map(({ repository }) =>
                repository === null
                    ? { nodes: [], hasNextPage: false }
                    : { nodes: repository.lsifUploads.nodes, hasNextPage: repository.lsifUploads.pageInfo.hasNextPage }
            )
        )
        .toPromise()

    const indices = range(nodes.length)
    const args: { [k: string]: string } = {}
    for (const i of indices) {
        args[`upload${i}`] = nodes[i].id
    }

    await gqlClient
        .mutateGraphQL(
            gql`
                mutation(${indices.map(i => `$upload${i}: ID!`).join(', ')}) {
                    ${indices.map(i => gql`delete${i}: deleteLSIFUpload(id: $upload${i}) { alwaysNil }`).join('\n')}
                }
            `,
            args
        )
        .pipe(map(dataOrThrowErrors))
        .toPromise()

    if (hasNextPage) {
        // If we have more upload, clear the next page
        return clearUploads(gqlClient, repoName)
    }
}

/**
 * Untar the LSIF data from the lsif-data directory and perform an LSIF upload
 * via src command, which must be a binary available on the $PATH. Returns the
 * URI of the upload's details (in UI).
 */
async function performUpload(
    config: Pick<Config, 'sourcegraphBaseUrl'>,
    {
        repository,
        commit,
        root,
        filename,
    }: {
        repository: string
        commit: string
        root: string
        filename: string
    }
): Promise<string> {
    const cwd = path.join(__dirname, 'lsif-data', path.dirname(repository))

    try {
        // Untar the lsif data for this upload
        const tarCommand = ['tar', '-xzf', `${path.basename(filename)}.gz`].join(' ')
        await child_process.exec(tarCommand, { cwd })
    } catch (error) {
        throw new Error(`Failed to untar test data: ${asError(error).message}`)
    }

    let out!: string
    try {
        // Upload data
        const uploadCommand = [
            `src -endpoint ${config.sourcegraphBaseUrl}`,
            'lsif upload',
            `-repo ${repository}`,
            `-commit ${commit}`,
            `-root ${root}`,
            `-file ${path.basename(filename)}`,
        ].join(' ')
        ;[out] = await child_process.exec(uploadCommand, { cwd })
    } catch (error) {
        try {
            // See if the error is due to a missing utility
            await child_process.exec('which src')
        } catch (error) {
            throw new Error('src-cli is not available on PATH')
        }

        throw new Error(
            `Failed to upload LSIF data: ${(error.stderr as string) || (error.stdout as string) || '(no output)'}`
        )
    }

    // Extract the status URL
    const match = out.match(/View processing status at (.+).\n$/)
    if (!match) {
        throw new Error(`Unexpected output from Sourcegraph cli: ${out}`)
    }

    return match[1]
}

/**
 * Wait on the upload page until it has finished processing and ensure that it's
 * visible at the tip of the default branch.
 */
async function ensureUpload(driver: Driver, uploadUrl: string): Promise<void> {
    await driver.page.goto(uploadUrl)

    await driver.page.waitFor(
        () => document.querySelector('.e2e-upload-state')?.textContent === 'Upload processed successfully.'
    )

    const isLatestForRepoText = await (await driver.page.waitFor('.e2e-is-latest-for-repo')).evaluate(
        elem => elem.textContent
    )
    expect(isLatestForRepoText).toEqual('yes')
}
