import { describe, test, before, beforeEach, after, afterEach } from 'mocha'
import { saveScreenshotsUponFailures } from '../../../shared/src/e2e/screenshotReporter'
import { retry } from '../../../shared/src/e2e/e2e-test-utils'
import { createDriverForTest, Driver, percySnapshot } from '../../../shared/src/e2e/driver'
import got from 'got'
import { gql } from '../../../shared/src/graphql/graphql'
import { random, sortBy } from 'lodash'
import MockDate from 'mockdate'
import { ExternalServiceKind } from '../../../shared/src/graphql/schema'
import { getConfig } from '../../../shared/src/e2e/config'
import assert from 'assert'
import expect from 'expect'
import { asError } from '../../../shared/src/util/errors'
import { Settings } from '../schema/settings.schema'

const { gitHubToken, sourcegraphBaseUrl } = getConfig('gitHubToken', 'sourcegraphBaseUrl')

describe('e2e test suite', () => {
    let driver: Driver

    before(async function() {
        // Cloning the repositories takes ~1 minute, so give initialization 2
        // minutes instead of 1 (which would be inherited from
        // `jest.setTimeout(1 * 60 * 1000)` above).
        this.timeout(2 * 60 * 1000)

        // Reset date mocking
        MockDate.reset()

        const config = getConfig('headless', 'slowMo', 'testUserPassword')

        // Start browser
        driver = await createDriverForTest({
            sourcegraphBaseUrl,
            logBrowserConsole: true,
            ...config,
        })
        const repoSlugs = [
            'sourcegraph/java-langserver',
            'gorilla/mux',
            'sourcegraphtest/AlwaysCloningTest',
            'sourcegraph/jsonrpc2',
            'sourcegraph/go-diff',
            'sourcegraph/appdash',
            'sourcegraph/sourcegraph-typescript',
            'sourcegraph-testing/automation-e2e-test',
        ]
        await driver.ensureLoggedIn({ username: 'test', password: config.testUserPassword, email: 'test@test.com' })
        await driver.resetUserSettings()
        await driver.ensureHasExternalService({
            kind: ExternalServiceKind.GITHUB,
            displayName: 'e2e-test-github',
            config: JSON.stringify({
                url: 'https://github.com',
                token: gitHubToken,
                repos: repoSlugs,
            }),
            ensureRepos: repoSlugs.map(slug => `github.com/${slug}`),
        })
    })

    // Close browser.
    after('Close browser', async () => {
        if (driver) {
            await driver.close()
        }
    })

    // Take a screenshot when a test fails.
    saveScreenshotsUponFailures(() => driver.page)

    beforeEach(async () => {
        if (driver) {
            // Clear local storage to reset sidebar selection (files or tabs) for each test
            await driver.page.evaluate(() => {
                localStorage.setItem('repo-rev-sidebar-last-tab', 'files')
            })

            await driver.resetUserSettings()
        }
    })

    describe('Core functionality', () => {
        test('Check settings are saved and applied', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/users/test/settings')
            await driver.page.waitForSelector('.e2e-settings-file .monaco-editor')

            const message = 'A wild notice appears!'
            await driver.replaceText({
                selector: '.e2e-settings-file .monaco-editor',
                newText: JSON.stringify({
                    notices: [
                        {
                            dismissable: false,
                            location: 'top',
                            message,
                        },
                    ],
                }),
                selectMethod: 'keyboard',
            })
            await driver.page.click('.e2e-settings-file .e2e-save-toolbar-save')
            await driver.page.waitForSelector('.e2e-global-alert .notices .global-alerts__alert', { visible: true })
            await driver.page.evaluate(message => {
                const elem = document.querySelector('.e2e-global-alert .notices .global-alerts__alert')
                if (!elem) {
                    throw new Error('No .e2e-global-alert .notices .global-alerts__alert element found')
                }
                if (!(elem as HTMLElement).innerText.includes(message)) {
                    throw new Error('Expected "' + message + '" message, but didn\'t find it')
                }
            }, message)
        })

        test('Check access tokens work (create, use and delete)', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/users/test/settings/tokens/new')
            await driver.page.waitForSelector('.e2e-create-access-token-description')

            const name = 'E2E Test ' + new Date().toISOString() + ' ' + random(1, 1e7)

            await driver.replaceText({
                selector: '.e2e-create-access-token-description',
                newText: name,
                selectMethod: 'keyboard',
            })

            await driver.page.click('.e2e-create-access-token-submit')
            const token: string = await (
                await driver.page.waitForFunction(() => {
                    const elem = document.querySelector<HTMLInputElement>('.e2e-access-token input[type=text]')
                    return elem?.value
                })
            ).jsonValue()

            const resp = await got.post('/.api/graphql', {
                prefixUrl: sourcegraphBaseUrl,
                headers: {
                    Authorization: 'token ' + token,
                },
                body: JSON.stringify({
                    query: gql`
                        query {
                            currentUser {
                                username
                            }
                        }
                    `,
                    variables: {},
                }),
            })

            const username = JSON.parse(resp.body).data.currentUser.username
            expect(username).toBe('test')

            await Promise.all([
                driver.acceptNextDialog(),
                (
                    await driver.page.waitForSelector(
                        `[data-e2e-access-token-description="${name}"] .e2e-access-token-delete`,
                        { visible: true }
                    )
                ).click(),
            ])

            await driver.page.waitFor(
                (name: string) => !document.querySelector(`[data-e2e-access-token-description="${name}"]`),
                {},
                name
            )
        })

        test('Check allowed usernames', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/users/test/settings/profile')
            await driver.page.waitForSelector('.e2e-user-settings-profile-page-username')

            const name = 'alice.bob-chris-'

            await driver.replaceText({
                selector: '.e2e-user-settings-profile-page-username',
                newText: name,
                selectMethod: 'selectall',
            })

            await driver.page.click('.e2e-user-settings-profile-page-update-profile')
            await driver.page.waitForSelector('.e2e-user-settings-profile-page-alert-success', { visible: true })

            await driver.page.goto(sourcegraphBaseUrl + `/users/${name}/settings/profile`)
            await driver.replaceText({
                selector: '.e2e-user-settings-profile-page-username',
                newText: 'test',
                selectMethod: 'selectall',
            })

            await driver.page.click('.e2e-user-settings-profile-page-update-profile')
            await driver.page.waitForSelector('.e2e-user-settings-profile-page-alert-success', { visible: true })
        })
    })

    describe('External services', () => {
        test('External service add, edit, delete', async () => {
            const displayName = 'e2e-github-test-2'
            await driver.ensureHasExternalService({
                kind: ExternalServiceKind.GITHUB,
                displayName,
                config:
                    '{"url": "https://github.myenterprise.com", "token": "initial-token", "repositoryQuery": ["none"]}',
            })
            await driver.page.goto(sourcegraphBaseUrl + '/site-admin/external-services')
            await (
                await driver.page.waitForSelector(
                    `[data-e2e-external-service-name="${displayName}"] .e2e-edit-external-service-button`
                )
            ).click()

            // Type in a new external service configuration.
            await driver.replaceText({
                selector: '.view-line',
                newText:
                    '{"url": "https://github.myenterprise.com", "token": "second-token", "repositoryQuery": ["none"]}',
                selectMethod: 'keyboard',
            })
            await driver.page.click('.e2e-update-external-service-button')
            // Must wait for the operation to complete, or else a "Discard changes?" dialog will pop up
            await driver.page.waitForSelector('.e2e-update-external-service-button:not([disabled])', { visible: true })

            await (
                await driver.page.waitForSelector('.list-group-item[href="/site-admin/external-services"]', {
                    visible: true,
                })
            ).click()

            await Promise.all([
                driver.acceptNextDialog(),
                (
                    await driver.page.waitForSelector(
                        '[data-e2e-external-service-name="e2e-github-test-2"] .e2e-delete-external-service-button',
                        { visible: true }
                    )
                ).click(),
            ])

            await driver.page.waitFor(
                () => !document.querySelector('[data-e2e-external-service-name="e2e-github-test-2"]')
            )
        })

        test('External service repositoryPathPattern', async () => {
            const repo = 'sourcegraph/go-blame' // Tiny repo, fast to clone
            const repositoryPathPattern = 'foobar/{host}/{nameWithOwner}'
            const slug = `github.com/${repo}`
            const pathPatternSlug = `foobar/github.com/${repo}`

            const config = {
                kind: ExternalServiceKind.GITHUB,
                displayName: 'e2e-test-github-repoPathPattern',
                config: JSON.stringify({
                    url: 'https://github.com',
                    token: gitHubToken,
                    repos: [repo],
                    repositoryPathPattern,
                }),
                // Make sure repository is named according to path pattern
                ensureRepos: [pathPatternSlug],
            }
            await driver.ensureHasExternalService(config)

            // Make sure repository slug without path pattern redirects to path pattern
            await driver.page.goto(sourcegraphBaseUrl + '/' + slug)
            await driver.assertWindowLocationPrefix('/' + pathPatternSlug)
        })

        const awsAccessKeyID = process.env.AWS_ACCESS_KEY_ID
        const awsSecretAccessKey = process.env.AWS_SECRET_ACCESS_KEY
        const awsCodeCommitUsername = process.env.AWS_CODE_COMMIT_GIT_USERNAME
        const awsCodeCommitPassword = process.env.AWS_CODE_COMMIT_GIT_PASSWORD

        const testIfAwsCredentialsSet =
            awsSecretAccessKey && awsAccessKeyID && awsCodeCommitUsername && awsCodeCommitPassword
                ? test
                : test.skip.bind(test)

        testIfAwsCredentialsSet('AWS CodeCommit', async () => {
            await driver.ensureHasExternalService({
                kind: ExternalServiceKind.AWSCODECOMMIT,
                displayName: 'e2e-aws-code-commit',
                config: JSON.stringify({
                    region: 'us-west-1',
                    accessKeyID: awsAccessKeyID,
                    secretAccessKey: awsSecretAccessKey,
                    repositoryPathPattern: 'aws/{name}',
                    gitCredentials: {
                        username: awsCodeCommitUsername,
                        password: awsCodeCommitPassword,
                    },
                }),
                ensureRepos: ['aws/test'],
            })
            await driver.page.goto(sourcegraphBaseUrl + '/aws/test/-/blob/README')
            const blob: string = await (
                await driver.page.waitFor(() => {
                    const elem = document.querySelector<HTMLElement>('.e2e-repo-blob')
                    return elem?.textContent
                })
            ).jsonValue()

            expect(blob).toBe('README\n\nchange')
        })

        const bbsURL = process.env.BITBUCKET_SERVER_URL
        const bbsToken = process.env.BITBUCKET_SERVER_TOKEN
        const bbsUsername = process.env.BITBUCKET_SERVER_USERNAME

        const testIfBBSCredentialsSet = bbsURL && bbsToken && bbsUsername ? test : test.skip.bind(test)

        testIfBBSCredentialsSet('Bitbucket Server', async () => {
            await driver.ensureHasExternalService({
                kind: ExternalServiceKind.BITBUCKETSERVER,
                displayName: 'e2e-bitbucket-server',
                config: JSON.stringify({
                    url: bbsURL,
                    token: bbsToken,
                    username: bbsUsername,
                    repos: ['SOURCEGRAPH/jsonrpc2'],
                    repositoryPathPattern: 'bbs/{projectKey}/{repositorySlug}',
                }),
                ensureRepos: ['bbs/SOURCEGRAPH/jsonrpc2'],
            })
            await driver.page.goto(sourcegraphBaseUrl + '/bbs/SOURCEGRAPH/jsonrpc2/-/blob/.travis.yml')
            const blob: string = await (
                await driver.page.waitFor(() => {
                    const elem = document.querySelector<HTMLElement>('.e2e-repo-blob')
                    return elem?.textContent
                })
            ).jsonValue()

            expect(blob).toBe('language: go\ngo: \n - 1.x\n\nscript:\n - go test -race -v ./...')
        })
    })

    describe('Visual tests', () => {
        test('Repositories list', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/site-admin/repositories?query=gorilla%2Fmux')
            await driver.page.waitForSelector('a[href="/github.com/gorilla/mux"]', { visible: true })
            await percySnapshot(driver.page, 'Repositories list')
        })

        test('Search results repo', async () => {
            await driver.page.goto(
                sourcegraphBaseUrl + '/search?q=repo:%5Egithub.com/gorilla/mux%24&patternType=regexp'
            )
            await driver.page.waitForSelector('a[href="/github.com/gorilla/mux"]', { visible: true })
            // Flaky https://github.com/sourcegraph/sourcegraph/issues/2704
            // await percySnapshot(page, 'Search results repo')
        })

        test('Search results file', async () => {
            await driver.page.goto(
                sourcegraphBaseUrl + '/search?q=repo:%5Egithub.com/gorilla/mux%24+file:%5Emux.go%24&patternType=regexp'
            )
            await driver.page.waitForSelector('a[href="/github.com/gorilla/mux"]', { visible: true })
            // Flaky https://github.com/sourcegraph/sourcegraph/issues/2704
            // await percySnapshot(page, 'Search results file')
        })

        test('Search results code', async () => {
            await driver.page.goto(
                sourcegraphBaseUrl +
                    '/search?q=repo:^github.com/gorilla/mux$&patternType=regexp file:mux.go "func NewRouter"'
            )
            await driver.page.waitForSelector('a[href="/github.com/gorilla/mux"]', { visible: true })
            // Flaky https://github.com/sourcegraph/sourcegraph/issues/2704
            // await percySnapshot(page, 'Search results code')
        })

        test('Site admin overview', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/site-admin')
            await driver.page.waitForSelector('.e2e-site-admin-overview-menu', { visible: true })
            await percySnapshot(driver.page, 'Site admin overview')
        })
    })

    describe('Theme switcher', () => {
        test('changes the theme', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/github.com/gorilla/mux/-/blob/mux.go')
            await driver.page.waitForSelector('.theme.theme-dark, .theme.theme-light', { visible: true })

            const getActiveThemeClasses = (): Promise<string[]> =>
                driver.page.evaluate(() =>
                    Array.from(document.querySelector('.theme')!.classList).filter(c => c.startsWith('theme-'))
                )

            expect(await getActiveThemeClasses()).toHaveLength(1)
            await driver.page.waitForSelector('.e2e-user-nav-item-toggle')
            await driver.page.click('.e2e-user-nav-item-toggle')

            // Switch to dark
            await driver.page.select('.e2e-theme-toggle', 'dark')
            expect(await getActiveThemeClasses()).toEqual(['theme-dark'])

            // Switch to light
            await driver.page.select('.e2e-theme-toggle', 'light')
            expect(await getActiveThemeClasses()).toEqual(['theme-light'])
        })
    })

    describe('Repository component', () => {
        const blobTableSelector = '.e2e-blob > table'
        /**
         * @param line 1-indexed line number
         * @param spanOffset 1-indexed index of the span that's to be clicked
         */
        const clickToken = async (line: number, spanOffset: number): Promise<void> => {
            const selector = `${blobTableSelector} tr:nth-child(${line}) > td.code > div:nth-child(1) > span:nth-child(${spanOffset})`
            await driver.page.waitForSelector(selector, { visible: true })
            await driver.page.click(selector)
        }

        // expectedCount defaults to one because of we haven't specified, we just want to ensure it exists at all
        const getHoverContents = async (expectedCount = 1): Promise<string[]> => {
            const selector =
                expectedCount > 1 ? `.e2e-tooltip-content:nth-child(${expectedCount})` : '.e2e-tooltip-content'
            await driver.page.waitForSelector(selector, { visible: true })
            return driver.page.evaluate(() =>
                // You can't reference hoverContentSelector in puppeteer's driver.page.evaluate
                Array.from(document.querySelectorAll('.e2e-tooltip-content')).map(t => t.textContent || '')
            )
        }
        const assertHoverContentContains = async (val: string, count?: number): Promise<void> => {
            expect(await getHoverContents(count)).toEqual(expect.arrayContaining([expect.stringContaining(val)]))
        }

        const clickHoverJ2D = async (): Promise<void> => {
            const selector = '.e2e-tooltip-go-to-definition'
            await driver.page.waitForSelector(selector, { visible: true })
            await driver.page.click(selector)
        }
        const clickHoverFindRefs = async (): Promise<void> => {
            const selector = '.e2e-tooltip-find-references'
            await driver.page.waitForSelector(selector, { visible: true })
            await driver.page.click(selector)
        }

        describe('file tree', () => {
            test('does navigation on file click', async () => {
                await driver.page.goto(
                    sourcegraphBaseUrl + '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d'
                )
                await (
                    await driver.page.waitForSelector('[data-tree-path="async.go"]', {
                        visible: true,
                    })
                ).click()
                await driver.assertWindowLocation(
                    '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d/-/blob/async.go'
                )
            })

            test('expands directory on row click (no navigation)', async () => {
                await driver.page.goto(
                    sourcegraphBaseUrl + '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d'
                )
                await driver.page.waitForSelector('.tree__row-icon', { visible: true })
                await driver.page.click('.tree__row-icon')
                await driver.page.waitForSelector('.tree__row--selected [data-tree-path="websocket"]', {
                    visible: true,
                })
                await driver.page.waitForSelector('.tree__row--expanded [data-tree-path="websocket"]', {
                    visible: true,
                })
                await driver.assertWindowLocation(
                    '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d'
                )
            })

            test('does navigation on directory row click', async () => {
                await driver.page.goto(
                    sourcegraphBaseUrl + '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d'
                )
                await driver.page.waitForSelector('.tree__row-label', { visible: true })
                await driver.page.click('.tree__row-label')
                await driver.page.waitForSelector('.tree__row--selected [data-tree-path="websocket"]', {
                    visible: true,
                })
                await driver.page.waitForSelector('.tree__row--expanded [data-tree-path="websocket"]', {
                    visible: true,
                })
                await driver.assertWindowLocation(
                    '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d/-/tree/websocket'
                )
            })

            test('selects the current file', async () => {
                await driver.page.goto(
                    sourcegraphBaseUrl +
                        '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d/-/blob/async.go'
                )
                await driver.page.waitForSelector('.tree__row--active [data-tree-path="async.go"]', {
                    visible: true,
                })
            })

            test('shows partial tree when opening directory', async () => {
                await driver.page.goto(
                    sourcegraphBaseUrl +
                        '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d/-/tree/websocket'
                )
                await driver.page.waitForSelector('.tree__row', { visible: true })
                expect(await driver.page.evaluate(() => document.querySelectorAll('.tree__row').length)).toEqual(1)
            })

            test('responds to keyboard shortcuts', async () => {
                const assertNumRowsExpanded = async (expectedCount: number): Promise<void> => {
                    expect(
                        await driver.page.evaluate(() => document.querySelectorAll('.tree__row--expanded').length)
                    ).toEqual(expectedCount)
                }

                await driver.page.goto(
                    sourcegraphBaseUrl +
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/.travis.yml'
                )
                await driver.page.waitForSelector('.tree__row', { visible: true }) // waitForSelector for tree to render

                await driver.page.click('.tree')
                await driver.page.keyboard.press('ArrowUp') // arrow up to 'diff' directory
                await driver.page.waitForSelector('.tree__row--selected [data-tree-path="diff"]', { visible: true })
                await driver.page.keyboard.press('ArrowRight') // arrow right (expand 'diff' directory)
                await driver.page.waitForSelector('.tree__row--selected [data-tree-path="diff"]', { visible: true })
                await driver.page.waitForSelector('.tree__row--expanded [data-tree-path="diff"]', { visible: true })
                await driver.page.waitForSelector('.tree__row [data-tree-path="diff/testdata"]', { visible: true })
                await driver.page.keyboard.press('ArrowRight') // arrow right (move to nested 'diff/testdata' directory)
                await driver.page.waitForSelector('.tree__row--selected [data-tree-path="diff/testdata"]', {
                    visible: true,
                })
                await assertNumRowsExpanded(1) // only `diff` directory is expanded, though `diff/testdata` is expanded

                await driver.page.keyboard.press('ArrowRight') // arrow right (expand 'diff/testdata' directory)
                await driver.page.waitForSelector('.tree__row--selected [data-tree-path="diff/testdata"]', {
                    visible: true,
                })
                await driver.page.waitForSelector('.tree__row--expanded [data-tree-path="diff/testdata"]', {
                    visible: true,
                })
                await assertNumRowsExpanded(2) // `diff` and `diff/testdata` directories expanded

                await driver.page.waitForSelector('.tree__row [data-tree-path="diff/testdata/empty.diff"]', {
                    visible: true,
                })
                // select some file nested under `diff/testdata`
                await driver.page.keyboard.press('ArrowDown') // arrow down
                await driver.page.keyboard.press('ArrowDown') // arrow down
                await driver.page.keyboard.press('ArrowDown') // arrow down
                await driver.page.keyboard.press('ArrowDown') // arrow down
                await driver.page.waitForSelector(
                    '.tree__row--selected [data-tree-path="diff/testdata/empty_orig.diff"]',
                    {
                        visible: true,
                    }
                )

                await driver.page.keyboard.press('ArrowLeft') // arrow left (navigate immediately up to parent directory `diff/testdata`)
                await driver.page.waitForSelector('.tree__row--selected [data-tree-path="diff/testdata"]', {
                    visible: true,
                })
                await assertNumRowsExpanded(2) // `diff` and `diff/testdata` directories expanded

                await driver.page.keyboard.press('ArrowLeft') // arrow left
                await driver.page.waitForSelector('.tree__row--selected [data-tree-path="diff/testdata"]', {
                    visible: true,
                }) // `diff/testdata` still selected
                await assertNumRowsExpanded(1) // only `diff` directory expanded
            })
        })
        describe('symbol sidebar', () => {
            const listSymbolsTests = [
                {
                    name: 'lists symbols in file for Go',
                    filePath:
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/cmd/go-diff/go-diff.go',
                    symbolNames: ['main', 'stdin', 'diffPath', 'fileIdx', 'main'],
                    symbolTypes: ['package', 'constant', 'variable', 'variable', 'function'],
                },
                {
                    name: 'lists symbols in another file for Go',
                    filePath:
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/diff.go',
                    symbolNames: [
                        'diff',
                        'Stat',
                        'Stat',
                        'hunkPrefix',
                        'hunkHeader',
                        'diffTimeParseLayout',
                        'diffTimeFormatLayout',
                        'add',
                    ],
                    symbolTypes: [
                        'package',
                        'function',
                        'function',
                        'variable',
                        'constant',
                        'constant',
                        'constant',
                        'function',
                    ],
                },
                {
                    name: 'lists symbols in file for Python',
                    filePath:
                        '/github.com/sourcegraph/appdash@ebfcffb1b5c00031ce797183546746715a3cfe87/-/blob/python/appdash/sockcollector.py',
                    symbolNames: [
                        'RemoteCollector',
                        'sock',
                        '_debug',
                        '__init__',
                        '_log',
                        'connect',
                        'collect',
                        'close',
                    ],
                    symbolTypes: ['class', 'variable', 'variable', 'field', 'field', 'field', 'field', 'field'],
                },
                {
                    name: 'lists symbols in file for TypeScript',
                    filePath:
                        '/github.com/sourcegraph/sourcegraph-typescript@a7b7a61e31af76dad3543adec359fa68737a58a1/-/blob/server/src/cancellation.ts',
                    symbolNames: [
                        'createAbortError',
                        'isAbortError',
                        'throwIfCancelled',
                        'tryCancel',
                        'toAxiosCancelToken',
                        'source',
                    ],
                    symbolTypes: ['constant', 'constant', 'function', 'function', 'function', 'constant'],
                },
                {
                    name: 'lists symbols in file for Java',
                    filePath:
                        '/github.com/sourcegraph/java-langserver@03efbe9558acc532e88f5288b4e6cfa155c6f2dc/-/blob/src/main/java/com/sourcegraph/common/Config.java',
                    symbolNames: [
                        'com.sourcegraph.common',
                        'Config',
                        'LIGHTSTEP_INCLUDE_SENSITIVE',
                        'LIGHTSTEP_PROJECT',
                        'LIGHTSTEP_TOKEN',
                        'ANDROID_JAR_PATH',
                        'IGNORE_DEPENDENCY_RESOLUTION_CACHE',
                        'LSP_TIMEOUT',
                        'LANGSERVER_ROOT',
                        'LOCAL_REPOSITORY',
                        'EXECUTE_GRADLE_ORIGINAL_ROOT_PATHS',
                        'shouldExecuteGradle',
                        'PRIVATE_REPO_ID',
                        'PRIVATE_REPO_URL',
                        'PRIVATE_REPO_USERNAME',
                        'PRIVATE_REPO_PASSWORD',
                        'log',
                        'checkEnv',
                        'ConfigException',
                    ],
                    symbolTypes: [
                        'package',
                        'class',
                        'field',
                        'field',
                        'field',
                        'field',
                        'field',
                        'field',
                        'field',
                        'field',
                        'field',
                        'method',
                        'field',
                        'field',
                        'field',
                        'field',
                        'field',
                        'method',
                        'class',
                    ],
                },
            ]

            for (const symbolTest of listSymbolsTests) {
                test(symbolTest.name, async () => {
                    await driver.page.goto(sourcegraphBaseUrl + symbolTest.filePath)

                    await (await driver.page.waitForSelector('[data-e2e-tab="symbols"]')).click()

                    await driver.page.waitForSelector('.e2e-symbol-name', { visible: true })

                    const symbolNames = await driver.page.evaluate(() =>
                        Array.from(document.querySelectorAll('.e2e-symbol-name')).map(t => t.textContent || '')
                    )
                    const symbolTypes = await driver.page.evaluate(() =>
                        Array.from(document.querySelectorAll('.e2e-symbol-icon')).map(
                            t => t.getAttribute('data-tooltip') || ''
                        )
                    )

                    expect(sortBy(symbolNames)).toEqual(sortBy(symbolTest.symbolNames))
                    expect(sortBy(symbolTypes)).toEqual(sortBy(symbolTest.symbolTypes))
                })
            }

            const navigateToSymbolTests = [
                {
                    name: 'navigates to file on symbol click for Go',
                    repoPath: '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d',
                    filePath: '/tree/cmd',
                    symbolPath: '/blob/cmd/go-diff/go-diff.go#L19:2-19:10',
                },
                {
                    name: 'navigates to file on symbol click for Java',
                    repoPath: '/github.com/sourcegraph/java-langserver@03efbe9558acc532e88f5288b4e6cfa155c6f2dc',
                    filePath: '/tree/src/main/java/com/sourcegraph/common',
                    symbolPath: '/blob/src/main/java/com/sourcegraph/common/Config.java#L14:20-14:26',
                    skip: true,
                },
                {
                    name:
                        'displays valid symbols at different file depths for Go (./examples/cmd/webapp-opentracing/main.go.go)',
                    repoPath: '/github.com/sourcegraph/appdash@ebfcffb1b5c00031ce797183546746715a3cfe87',
                    filePath: '/tree/examples',
                    symbolPath: '/blob/examples/cmd/webapp-opentracing/main.go#L26:6-26:10',
                    skip: true,
                },
                {
                    name: 'displays valid symbols at different file depths for Go (./sqltrace/sql.go)',
                    repoPath: '/github.com/sourcegraph/appdash@ebfcffb1b5c00031ce797183546746715a3cfe87',
                    filePath: '/tree/sqltrace',
                    symbolPath: '/blob/sqltrace/sql.go#L14:2-14:5',
                    skip: true,
                },
            ]

            for (const navigationTest of navigateToSymbolTests) {
                const testFunc = navigationTest.skip ? test.skip : test
                testFunc(navigationTest.name, async () => {
                    const repoBaseURL = sourcegraphBaseUrl + navigationTest.repoPath + '/-'

                    await driver.page.goto(repoBaseURL + navigationTest.filePath)

                    await (await driver.page.waitForSelector('[data-e2e-tab="symbols"]')).click()

                    await driver.page.waitForSelector('.e2e-symbol-name', { visible: true })

                    await (
                        await driver.page.waitForSelector(`.e2e-symbol-link[href*="${navigationTest.symbolPath}"]`, {
                            visible: true,
                        })
                    ).click()
                    await driver.assertWindowLocation(repoBaseURL + navigationTest.symbolPath, true)
                })
            }

            const highlightSymbolTests = [
                {
                    name: 'highlights correct line for Go',
                    filePath:
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/diff.go',
                    index: 5,
                    line: 65,
                },
                {
                    name: 'highlights correct line for Typescript',
                    filePath:
                        '/github.com/sourcegraph/sourcegraph-typescript@a7b7a61e31af76dad3543adec359fa68737a58a1/-/blob/server/src/cancellation.ts',
                    index: 2,
                    line: 17,
                },
            ]

            for (const { name, filePath, index, line } of highlightSymbolTests) {
                test(name, async () => {
                    await driver.page.goto(sourcegraphBaseUrl + filePath)
                    await driver.page.waitForSelector('[data-e2e-tab="symbols"]')
                    await driver.page.click('[data-e2e-tab="symbols"]')
                    await driver.page.waitForSelector('.e2e-symbol-name', { visible: true })
                    await driver.page.click(`.filtered-connection__nodes li:nth-child(${index + 1}) a`)

                    await driver.page.waitForSelector('.e2e-blob .selected .line')
                    const selectedLineNumber = await driver.page.evaluate(() => {
                        const elem = document.querySelector<HTMLElement>('.e2e-blob .selected .line')
                        return elem?.dataset.line && parseInt(elem.dataset.line, 10)
                    })

                    expect(selectedLineNumber).toEqual(line)
                })
            }
        })

        describe('directory page', () => {
            // TODO(slimsag:discussions): temporarily disabled because the discussions feature flag removes this component.
            /*
            it('shows a row for each file in the directory', async () => {
                await driver.page.goto(sourcegraphBaseUrl + '/github.com/gorilla/securecookie@e59506cc896acb7f7bf732d4fdf5e25f7ccd8983')
                await enableOrAddRepositoryIfNeeded()
                await driver.page.waitForSelector('.tree-page__entries-directories', { visible: true })
                await retry(async () =>
                    assert.equal(
                        await driver.page.evaluate(
                            () => document.querySelectorAll('.tree-page__entries-directories .tree-entry').length
                        ),
                        1
                    )
                )
                await retry(async () =>
                    assert.equal(
                        await driver.page.evaluate(
                            () => document.querySelectorAll('.tree-page__entries-files .tree-entry').length
                        ),
                        7
                    )
                )
            })
            */

            test('shows commit information on a row', async () => {
                await driver.page.goto(
                    sourcegraphBaseUrl + '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d',
                    {
                        waitUntil: 'domcontentloaded',
                    }
                )
                await driver.page.waitForSelector('.e2e-tree-page-no-recent-commits')
                await driver.page.click('.e2e-tree-page-show-all-commits')
                await driver.page.waitForSelector('.git-commit-node__message', { visible: true })
                await retry(async () =>
                    expect(
                        await driver.page.evaluate(
                            () => document.querySelectorAll('.git-commit-node__message')[3].textContent
                        )
                    ).toContain('Add support for new/removed binary files.')
                )
                await retry(async () =>
                    expect(
                        await driver.page.evaluate(() =>
                            document.querySelectorAll('.git-commit-node-byline')[3].textContent!.trim()
                        )
                    ).toContain('Dmitri Shuralyov')
                )
                await retry(async () =>
                    expect(
                        await driver.page.evaluate(
                            () => document.querySelectorAll('.git-commit-node__oid')[3].textContent
                        )
                    ).toEqual('2083912')
                )
            })

            // TODO(slimsag:discussions): temporarily disabled because the discussions feature flag removes this component.
            /*
            it('navigates when clicking on a row', async () => {
                await driver.page.goto(sourcegraphBaseUrl + '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d')
                await enableOrAddRepositoryIfNeeded()
                // click on directory
                await driver.page.waitForSelector('.tree-entry', { visible: true })
                await driver.page.click('.tree-entry')
                await assertWindowLocation(
                    '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d/-/tree/websocket'
                )
            })
            */
        })

        describe('rev resolution', () => {
            test('shows clone in progress interstitial page', async () => {
                await driver.page.goto(sourcegraphBaseUrl + '/github.com/sourcegraphtest/AlwaysCloningTest')
                await driver.page.waitForSelector('.hero-page__subtitle', { visible: true })
                await retry(async () =>
                    expect(
                        await driver.page.evaluate(() => document.querySelector('.hero-page__subtitle')!.textContent)
                    ).toEqual('Cloning in progress')
                )
            })

            test('resolves default branch when unspecified', async () => {
                await driver.page.goto(sourcegraphBaseUrl + '/github.com/sourcegraph/go-diff/-/blob/diff/diff.go')
                await driver.page.waitForSelector('#repo-rev-popover', { visible: true })
                await retry(async () => {
                    expect(
                        await driver.page.evaluate(() => document.querySelector('.e2e-revision')!.textContent!.trim())
                    ).toEqual('master')
                })
                // Verify file contents are loaded.
                await driver.page.waitForSelector(blobTableSelector)
            })

            test('updates rev with switcher', async () => {
                await driver.page.goto(sourcegraphBaseUrl + '/github.com/sourcegraph/go-diff/-/blob/diff/diff.go')
                // Open rev switcher
                await driver.page.waitForSelector('#repo-rev-popover', { visible: true })
                await driver.page.click('#repo-rev-popover')
                // Click "Tags" tab
                await driver.page.click('.revisions-popover .tab-bar__tab:nth-child(2)')
                await driver.page.waitForSelector('a.git-ref-node[href*="0.5.0"]', { visible: true })
                await driver.page.click('a.git-ref-node[href*="0.5.0"]')
                await driver.assertWindowLocation('/github.com/sourcegraph/go-diff@v0.5.0/-/blob/diff/diff.go')
            })
        })

        describe('hovers', () => {
            describe('Blob', () => {
                test('gets displayed and updates URL when clicking on a token', async () => {
                    await driver.page.goto(
                        sourcegraphBaseUrl +
                            '/github.com/gorilla/mux@15a353a636720571d19e37b34a14499c3afa9991/-/blob/mux.go'
                    )
                    await driver.page.waitForSelector(blobTableSelector)
                    await clickToken(24, 5)
                    await driver.assertWindowLocation(
                        '/github.com/gorilla/mux@15a353a636720571d19e37b34a14499c3afa9991/-/blob/mux.go#L24:19'
                    )
                    await getHoverContents() // verify there is a hover
                    await percySnapshot(driver.page, 'Code intel hover tooltip')
                })

                test('gets displayed when navigating to a URL with a token position', async () => {
                    await driver.page.goto(
                        sourcegraphBaseUrl +
                            '/github.com/gorilla/mux@15a353a636720571d19e37b34a14499c3afa9991/-/blob/mux.go#L151:23'
                    )
                    await assertHoverContentContains(
                        'ErrMethodMismatch is returned when the method in the request does not match'
                    )
                })

                describe('jump to definition', () => {
                    test('noops when on the definition', async () => {
                        await driver.page.goto(
                            sourcegraphBaseUrl +
                                '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L29:6'
                        )
                        await clickHoverJ2D()
                        await driver.assertWindowLocation(
                            '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L29:6'
                        )
                    })

                    test('does navigation (same repo, same file)', async () => {
                        await driver.page.goto(
                            sourcegraphBaseUrl +
                                '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L25:10'
                        )
                        await clickHoverJ2D()
                        await driver.assertWindowLocation(
                            '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L29:6'
                        )
                    })

                    test('does navigation (same repo, different file)', async () => {
                        await driver.page.goto(
                            sourcegraphBaseUrl +
                                '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/print.go#L13:31'
                        )
                        await clickHoverJ2D()
                        await driver.assertWindowLocation(
                            '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/diff.pb.go#L38:6'
                        )
                        // Verify file tree is highlighting the new path.
                        await driver.page.waitForSelector('.tree__row--active [data-tree-path="diff/diff.pb.go"]', {
                            visible: true,
                        })
                    })

                    // basic code intel doesn't support cross-repo jump-to-definition yet.
                    // If this test gets re-enabled `sourcegraph/vcsstore` and
                    // `sourcegraph/go-vcs` need to be cloned.
                    test.skip('does navigation (external repo)', async () => {
                        await driver.page.goto(
                            sourcegraphBaseUrl +
                                '/github.com/sourcegraph/vcsstore@267289226b15e5b03adedc9746317455be96e44c/-/blob/server/diff.go#L27:30'
                        )
                        await clickHoverJ2D()
                        await driver.assertWindowLocation(
                            '/github.com/sourcegraph/go-vcs@aa7c38442c17a3387b8a21f566788d8555afedd0/-/blob/vcs/repository.go#L103:6'
                        )
                    })
                })

                describe('find references', () => {
                    test('opens widget and fetches local references', async function() {
                        this.timeout(120000)

                        await driver.page.goto(
                            sourcegraphBaseUrl +
                                '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L29:6'
                        )
                        await clickHoverFindRefs()
                        await driver.assertWindowLocation(
                            '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L29:6&tab=references'
                        )

                        await driver.assertNonemptyLocalRefs()

                        // verify the appropriate # of references are fetched
                        await driver.page.waitForSelector('.panel__tabs-content .file-match-children', {
                            visible: true,
                        })
                        await retry(async () =>
                            expect(
                                await driver.page.evaluate(
                                    () =>
                                        document.querySelectorAll('.panel__tabs-content .file-match-children__item')
                                            .length
                                )
                            ).toEqual(
                                // Basic code intel finds 8 references with some overlapping context, resulting in 4 hunks.
                                4
                            )
                        )

                        // verify all the matches highlight a `MultiFileDiffReader` token
                        await driver.assertAllHighlightedTokens('MultiFileDiffReader')
                    })

                    // TODO unskip this once basic-code-intel looks for external
                    // references even when local references are found.
                    test.skip('opens widget and fetches external references', async () => {
                        await driver.page.goto(
                            sourcegraphBaseUrl +
                                '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L32:16&tab=references'
                        )

                        // verify some external refs are fetched (we cannot assert how many, but we can check that the matched results
                        // look like they're for the appropriate token)
                        await driver.assertNonemptyExternalRefs()

                        // verify all the matches highlight a `Reader` token
                        await driver.assertAllHighlightedTokens('Reader')
                    })
                })
            })
        })

        describe.skip('godoc.org "Uses" links', () => {
            test('resolves standard library function', async () => {
                // https://godoc.org/bytes#Compare
                await driver.page.goto(sourcegraphBaseUrl + '/-/godoc/refs?def=Compare&pkg=bytes&repo=')
                await driver.assertWindowLocationPrefix('/github.com/golang/go/-/blob/src/bytes/bytes_decl.go')
                await driver.assertStickyHighlightedToken('Compare')
                await driver.assertNonemptyLocalRefs()
                await driver.assertAllHighlightedTokens('Compare')
            })

            test('resolves standard library function (from stdlib repo)', async () => {
                // https://godoc.org/github.com/golang/go/src/bytes#Compare
                await driver.page.goto(
                    sourcegraphBaseUrl +
                        '/-/godoc/refs?def=Compare&pkg=github.com%2Fgolang%2Fgo%2Fsrc%2Fbytes&repo=github.com%2Fgolang%2Fgo'
                )
                await driver.assertWindowLocationPrefix('/github.com/golang/go/-/blob/src/bytes/bytes_decl.go')
                await driver.assertStickyHighlightedToken('Compare')
                await driver.assertNonemptyLocalRefs()
                await driver.assertAllHighlightedTokens('Compare')
            })

            test('resolves external package function (from gorilla/mux)', async () => {
                // https://godoc.org/github.com/gorilla/mux#Router
                await driver.page.goto(
                    sourcegraphBaseUrl +
                        '/-/godoc/refs?def=Router&pkg=github.com%2Fgorilla%2Fmux&repo=github.com%2Fgorilla%2Fmux'
                )
                await driver.assertWindowLocationPrefix('/github.com/gorilla/mux/-/blob/mux.go')
                await driver.assertStickyHighlightedToken('Router')
                await driver.assertNonemptyLocalRefs()
                await driver.assertAllHighlightedTokens('Router')
            })
        })

        describe('external code host links', () => {
            test('on repo navbar ("View on GitHub")', async () => {
                await driver.page.goto(
                    sourcegraphBaseUrl +
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L19',
                    { waitUntil: 'domcontentloaded' }
                )
                await driver.page.waitForSelector('.nav-link[href*="https://github"]', { visible: true })
                await retry(async () =>
                    expect(
                        await driver.page.evaluate(
                            () =>
                                (document.querySelector('.nav-link[href*="https://github"]') as HTMLAnchorElement).href
                        )
                    ).toEqual(
                        'https://github.com/sourcegraph/go-diff/blob/3f415a150aec0685cb81b73cc201e762e075006d/diff/parse.go#L19'
                    )
                )
            })
        })
    })

    describe('Search component', () => {
        test('can execute search with search operators', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/github.com/sourcegraph/go-diff')

            const operators: { [key: string]: string } = {
                repo: '^github.com/sourcegraph/go-diff$',
                count: '1000',
                type: 'file',
                file: '.go',
                '-file': '.md',
            }

            const operatorsQuery = Object.keys(operators)
                .map(op => `${op}:${operators[op]}`)
                .join('+')

            await driver.page.goto(`${sourcegraphBaseUrl}/search?q=diff+${operatorsQuery}&patternType=regexp`)
            await driver.page.waitForSelector('.e2e-search-results-stats', { visible: true })
            await retry(async () => {
                const label = await driver.page.evaluate(
                    () => document.querySelector('.e2e-search-results-stats')!.textContent || ''
                )
                expect(label.includes('results')).toEqual(true)
            })
            await driver.page.waitForSelector('.e2e-file-match-children-item', { visible: true })
        })

        test('renders results for sourcegraph/go-diff (no search group)', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/github.com/sourcegraph/go-diff')
            await driver.page.goto(
                sourcegraphBaseUrl +
                    '/search?q=diff+repo:sourcegraph/go-diff%403f415a150aec0685cb81b73cc201e762e075006d+type:file&patternType=regexp'
            )
            await driver.page.waitForSelector('.e2e-search-results-stats', { visible: true })
            await retry(async () => {
                const label = await driver.page.evaluate(
                    () => document.querySelector('.e2e-search-results-stats')!.textContent || ''
                )
                expect(label.includes('results')).toEqual(true)
            })

            const firstFileMatchHref = await driver.page.$eval(
                '.e2e-file-match-children-item',
                a => (a as HTMLAnchorElement).href
            )

            // navigate to result on click
            await driver.page.click('.e2e-file-match-children-item')

            await retry(async () => {
                expect(await driver.page.evaluate(() => window.location.href)).toEqual(firstFileMatchHref)
            })
        })

        describe('multiple revisions per repository', () => {
            let previousExperimentalFeatures: any
            before(async () => {
                await driver.setConfig(['experimentalFeatures'], prev => {
                    previousExperimentalFeatures = prev?.value
                    return { searchMultipleRevisionsPerRepository: true }
                })
                // Wait for configuration to be applied.
                await new Promise(resolve => setTimeout(resolve, 6000))
            })
            after(async () => {
                await driver.setConfig(['experimentalFeatures'], () => previousExperimentalFeatures)
            })

            test('searches', async () => {
                await driver.page.goto(
                    sourcegraphBaseUrl +
                        '/search?q=repo:sourcegraph/go-diff%24%40master:print-options:*refs/heads/+func+NewHunksReader&patternType=regexp'
                )
                await driver.page.waitForSelector('.e2e-search-results-stats', { visible: true })
                await retry(async () => {
                    const label = await driver.page.evaluate(
                        () => document.querySelector('.e2e-search-results-stats')!.textContent || ''
                    )
                    expect(label.includes('results')).toEqual(true)
                })

                const fileMatchHrefs = (
                    await driver.page.$$eval('.e2e-file-match-children-item', as =>
                        as.map(a => (a as HTMLAnchorElement).pathname)
                    )
                ).sort()

                // Only check for specific branches, so that the test doesn't break when new
                // branches are added (it's an active repository).
                const checkBranches = [
                    'master',
                    'print-options',

                    // These next 2 branches are included because of the *refs/heads/ in the query.
                    // If they are ever deleted from the actual live repository, replace them with
                    // any other branches that still exist.
                    'test-already-exist-pr',
                    'bug-fix-wip',
                ].sort()
                expect(
                    fileMatchHrefs.filter(href => checkBranches.some(branch => href.includes(`@${branch}/`)))
                ).toEqual(checkBranches.map(branch => `/github.com/sourcegraph/go-diff@${branch}/-/blob/diff/parse.go`))
            })
        })

        test('accepts query for sourcegraph/jsonrpc2', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/search')

            // Update the input value
            await driver.page.waitForSelector('.e2e-query-input', { visible: true })
            await driver.page.keyboard.type('test repo:sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d')

            // TODO: test search scopes

            // Submit the search
            await driver.page.click('.search-button')

            await driver.page.waitForSelector('.e2e-search-results-stats', { visible: true })
            await retry(async () => {
                const label = await driver.page.evaluate(
                    () => document.querySelector('.e2e-search-results-stats')!.textContent || ''
                )
                const match = /(\d+) results?/.exec(label)
                if (!match) {
                    throw new Error(
                        `.e2e-search-results-stats textContent did not match regex '(\\d+) results': '${label}'`
                    )
                }
                const numberOfResults = parseInt(match[1], 10)
                expect(numberOfResults).toBeGreaterThan(0)
            })
        })

        test('redirects to a URL with &patternType=regexp if no patternType in URL', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/search?q=test')
            await driver.assertWindowLocation('/search?q=test&patternType=regexp')
        })

        test('regexp toggle appears and updates patternType query parameter when clicked', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/search?q=test&patternType=literal')
            await driver.page.waitForSelector('.e2e-query-input')
            await driver.page.waitForSelector('.e2e-regexp-toggle')
            await driver.page.click('.e2e-regexp-toggle')
            await driver.page.goto(sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
            await driver.page.waitForSelector('.e2e-regexp-toggle')
            await driver.page.click('.e2e-regexp-toggle')
            await driver.page.goto(sourcegraphBaseUrl + '/search?q=test&patternType=literal')
        })
    })

    describe('Search pattern type setting', () => {
        test('Search pattern type setting correctly sets default pattern type', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/users/test/settings')
            await driver.replaceText({
                selector: '.e2e-settings-file .monaco-editor',
                newText: JSON.stringify({
                    'search.defaultPatternType': 'regexp',
                }),
                selectMethod: 'keyboard',
            })
            await driver.page.click('.e2e-settings-file .e2e-save-toolbar-save')

            await driver.page.goto(sourcegraphBaseUrl + '/search')
            await driver.page.waitForSelector('.e2e-query-input', { visible: true })
            await driver.page.waitForSelector('.e2e-regexp-toggle', { visible: true })

            const activeToggle = await driver.page.evaluate(
                () => document.querySelectorAll('.e2e-regexp-toggle--active').length
            )
            expect(activeToggle).toEqual(1)
            await driver.page.keyboard.type('test')
            await driver.page.click('.search-button')
            await driver.assertWindowLocation('/search?q=test&patternType=regexp')
        })
    })

    describe('Search result type tabs', () => {
        test('Search results type tabs appear', async () => {
            await driver.page.goto(
                sourcegraphBaseUrl + '/search?q=repo:%5Egithub.com/gorilla/mux%24&patternType=regexp'
            )
            await driver.page.waitForSelector('.e2e-search-result-type-tabs', { visible: true })
            await driver.page.waitForSelector('.e2e-search-result-tab--active', { visible: true })
            const tabs = await driver.page.evaluate(() =>
                Array.from(document.querySelectorAll('.e2e-search-result-tab'), tab => tab.textContent)
            )

            expect(tabs.length).toEqual(6)
            expect(tabs).toStrictEqual(['Code', 'Diffs', 'Commits', 'Symbols', 'Repositories', 'Filenames'])

            const activeTab = await driver.page.evaluate(
                () => document.querySelectorAll('.e2e-search-result-tab--active').length
            )
            expect(activeTab).toEqual(1)

            const label = await driver.page.evaluate(
                () => document.querySelector('.e2e-search-result-tab--active')!.textContent || ''
            )
            expect(label).toEqual('Code')
        })

        test.skip('Clicking search results tabs updates query and URL', async () => {
            for (const searchType of ['diff', 'commit', 'symbol', 'repo']) {
                await driver.page.waitForSelector(`.e2e-search-result-tab-${searchType}`)
                await driver.page.click(`.e2e-search-result-tab-${searchType}`)
                await driver.assertWindowLocation(
                    `/search?q=repo:%5Egithub.com/gorilla/mux%24+type:${searchType}&patternType=regexp`
                )
            }
        })
    })

    describe('Saved searches', () => {
        test('Save search from search results page', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/search?q=test')
            await driver.page.waitForSelector('.e2e-save-search-link', { visible: true })
            await driver.page.click('.e2e-save-search-link')
            await driver.page.waitForSelector('.e2e-saved-search-modal')
            await driver.page.waitForSelector('.e2e-saved-search-modal-save-button')
            await driver.page.click('.e2e-saved-search-modal-save-button')
            await driver.assertWindowLocation('/users/test/searches/add?query=test&patternType=regexp')

            await driver.page.waitForSelector('.e2e-saved-search-form-input-description', { visible: true })
            await driver.page.click('.e2e-saved-search-form-input-description')
            await driver.page.keyboard.type('test query')
            await driver.page.waitForSelector('.e2e-saved-search-form-submit-button', { visible: true })
            await driver.page.click('.e2e-saved-search-form-submit-button')
            await driver.assertWindowLocation('/users/test/searches')

            const nodes = await driver.page.evaluate(
                () => document.querySelectorAll('.e2e-saved-search-list-page-row').length
            )
            expect(nodes).toEqual(1)

            expect(
                await driver.page.evaluate(
                    () => document.querySelector('.e2e-saved-search-list-page-row-title')!.textContent
                )
            ).toEqual('test query')
        })
        test('Delete saved search', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/users/test/searches')
            await driver.page.waitForSelector('.e2e-delete-saved-search-button', { visible: true })
            driver.page.on('dialog', async dialog => {
                await dialog.accept()
            })
            await driver.page.click('.e2e-delete-saved-search-button')
            await driver.page.waitFor(() => !document.querySelector('.e2e-saved-search-list-page-row'))
            const nodes = await driver.page.evaluate(
                () => document.querySelectorAll('.e2e-saved-search-list-page-row').length
            )
            expect(nodes).toEqual(0)
        })
        test('Save search from saved searches page', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/users/test/searches')
            await driver.page.waitForSelector('.e2e-add-saved-search-button', { visible: true })
            await driver.page.click('.e2e-add-saved-search-button')
            await driver.assertWindowLocation('/users/test/searches/add')

            await driver.page.waitForSelector('.e2e-saved-search-form-input-description', { visible: true })
            await driver.page.click('.e2e-saved-search-form-input-description')
            await driver.page.keyboard.type('test query 2')

            await driver.page.waitForSelector('.e2e-saved-search-form-input-query', { visible: true })
            await driver.page.click('.e2e-saved-search-form-input-query')
            await driver.page.keyboard.type('test patternType:literal')

            await driver.page.waitForSelector('.e2e-saved-search-form-submit-button', { visible: true })
            await driver.page.click('.e2e-saved-search-form-submit-button')
            await driver.assertWindowLocation('/users/test/searches')

            const nodes = await driver.page.evaluate(
                () => document.querySelectorAll('.e2e-saved-search-list-page-row').length
            )
            expect(nodes).toEqual(1)

            expect(
                await driver.page.evaluate(
                    () => document.querySelector('.e2e-saved-search-list-page-row-title')!.textContent
                )
            ).toEqual('test query 2')
        })
        test('Edit saved search', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/users/test/searches')
            await driver.page.waitForSelector('.e2e-edit-saved-search-button', { visible: true })
            await driver.page.click('.e2e-edit-saved-search-button')

            await driver.page.waitForSelector('.e2e-saved-search-form-input-description', { visible: true })
            await driver.page.click('.e2e-saved-search-form-input-description')
            await driver.page.keyboard.type(' edited')

            await driver.page.waitForSelector('.e2e-saved-search-form-submit-button', { visible: true })
            await driver.page.click('.e2e-saved-search-form-submit-button')
            await driver.page.goto(sourcegraphBaseUrl + '/users/test/searches')
            await driver.page.waitForSelector('.e2e-saved-search-list-page-row-title')

            expect(
                await driver.page.evaluate(
                    () => document.querySelector('.e2e-saved-search-list-page-row-title')!.textContent
                )
            ).toEqual('test query 2 edited')
        })
    })

    describe('Search statistics', () => {
        beforeEach(async () => {
            await driver.setUserSettings<Settings>({ experimentalFeatures: { searchStats: true } })
        })
        afterEach(async () => {
            await driver.resetUserSettings()
        })

        // This is a substring that appears in the sourcegraph/go-diff repository, which is present
        // in the external service added for the e2e test. It is OK if it starts to appear in other
        // repositories (such as sourcegraph/sourcegraph now that it's mentioned here); the test
        // just checks that it is found in at least 1 Go file.
        const uniqueString = 'Incomplete-'
        const uniqueStringPostfix = 'Lines'

        test('button on search results page', async () => {
            await driver.page.goto(`${sourcegraphBaseUrl}/search?q=${uniqueString}`)
            await driver.page.waitForSelector(`a[href="/stats?q=${uniqueString}"]`)
        })

        test('page', async () => {
            await driver.page.goto(`${sourcegraphBaseUrl}/stats?q=${uniqueString}`)

            // Ensure the global navbar hides the search input (to avoid confusion with the one on
            // the stats page).
            await driver.page.waitForSelector('.global-navbar a.nav-link[href="/search"]')
            assert.strictEqual(
                await driver.page.evaluate(() => document.querySelectorAll('.e2e-query-input').length),
                0
            )

            const queryInputValue = () =>
                driver.page.evaluate(() => {
                    const input = document.querySelector<HTMLInputElement>('.e2e-stats-query')
                    return input ? input.value : null
                })

            // Check for a Go result (the sample repositories have Go files).
            await driver.page.waitForSelector(`a[href*="${uniqueString}+lang:go"]`)
            assert.strictEqual(await queryInputValue(), uniqueString)
            await percySnapshot(driver.page, 'Search stats')

            // Update the query and rerun the computation.
            await driver.page.type('.e2e-stats-query', uniqueStringPostfix) // the uniqueString is followed by 'Incomplete-Lines' in go-diff
            const wantQuery = `${uniqueString}${uniqueStringPostfix}`
            assert.strictEqual(await queryInputValue(), wantQuery)
            await driver.page.click('.e2e-stats-query-update')
            await driver.page.waitForSelector(`a[href*="${wantQuery}+lang:go"]`)
            assert.ok(driver.page.url().endsWith(`/stats?q=${wantQuery}`))
        })
    })

    describe('Campaigns', () => {
        let previousExperimentalFeatures: any
        before(async () => {
            await driver.setConfig(['experimentalFeatures'], prev => {
                previousExperimentalFeatures = prev?.value
                return { automation: 'enabled' }
            })
            // wait for configuration to be applied
            await retry(async () => {
                await driver.page.goto(sourcegraphBaseUrl + '/campaigns/new')
                try {
                    assert.notStrictEqual(
                        await driver.page.evaluate(() => document.querySelectorAll('.e2e-campaign-nav-entry').length),
                        0
                    )
                } catch (error) {
                    await new Promise(resolve => setTimeout(resolve, 1000))
                    throw asError(error)
                }
            })
        })
        after(async () => {
            await driver.setConfig(['experimentalFeatures'], () => previousExperimentalFeatures)
        })
        async function testCampaignPreview({
            previewURL,
            changesetCount,
            snapshotName,
        }: {
            previewURL: string
            changesetCount: number
            snapshotName: string
        }): Promise<void> {
            await driver.page.goto(previewURL.replace('127.0.0.1', 'localhost'))
            await driver.page.waitForSelector('.e2e-campaign-form')

            // fill campaign preview form
            await driver.page.type('.e2e-campaign-title', 'E2E campaign')

            await driver.page.waitForSelector('.e2e-changeset-node')
            // check there were exactly as expected diffs generated
            const generatedChangesetCount = await driver.page.evaluate(
                () => document.querySelectorAll('.e2e-changeset-node').length
            )
            expect(generatedChangesetCount).toEqual(changesetCount)
            await percySnapshot(driver.page, snapshotName)
        }
        test('View campaign preview for plan', async () => {
            const repo = await driver.getRepository('github.com/sourcegraph-testing/automation-e2e-test')
            const { previewURL } = await driver.createCampaignPlanFromPatches([
                {
                    repository: repo.id,
                    baseRevision: '339d09ae1ce5907e0678ae5f1f91d9ad38db6107',
                    baseRef: 'refs/heads/master',
                    patch: `diff --unified file1.txt file1.txt
--- file1.txt 2020-01-01 01:02:03 -0700
+++ file1.txt 2020-01-01 03:04:05 -0700
@@ -1 +1,2 @@
 this is file 1
+hello
`,
                },
            ])
            await testCampaignPreview({
                previewURL,
                changesetCount: 1,
                snapshotName: 'Campaign preview page',
            })
        })
    })

    describe('Interactive search mode (feature flagged)', () => {
        let previousExperimentalFeatures: any

        before(async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/site-admin/global-settings')
            await driver.page.waitForSelector('.e2e-settings-file .monaco-editor')

            await driver.replaceText({
                selector: '.e2e-settings-file .monaco-editor',
                newText: JSON.stringify({
                    experimentalFeatures: {
                        splitSearchModes: true,
                    },
                }),
                selectMethod: 'keyboard',
            })

            await driver.page.click('.e2e-settings-file .e2e-save-toolbar-save')
            // wait for configuration to be applied
            await retry(
                async () => {
                    await driver.page.goto(sourcegraphBaseUrl + '/search')
                    await driver.page.waitForSelector('.e2e-search-mode-toggle')
                    assert.notStrictEqual(
                        await driver.page.evaluate(() => document.querySelectorAll('.e2e-search-mode-toggle').length),
                        0,
                        'Expected search mode toggle to appear, but was not able to find it on the page.'
                    )
                },
                { minTimeout: 1000 }
            )
        })

        after(async () => {
            await driver.setConfig(['experimentalFeatures'], () => previousExperimentalFeatures)
        })

        test('Interactive search mode component appears', async () => {
            await driver.page.waitForSelector('.e2e-search-mode-toggle')
            expect(
                await driver.page.evaluate(() => {
                    const toggles = document.querySelectorAll('.e2e-search-mode-toggle')
                    return toggles.length
                })
            ).toBe(1)
        })

        test('Interactive search mode filter buttons', async () => {
            await driver.page.waitForSelector('.e2e-search-mode-toggle', { visible: true })
            await driver.page.click('.e2e-search-mode-toggle')
            await driver.page.click('.e2e-search-mode-toggle__interactive-mode')

            // Wait for the input component to appear
            await driver.page.waitForSelector('.e2e-interactive-mode-input', { visible: true })
            // Wait for the add filter row to appear.
            await driver.page.waitForSelector('.e2e-add-filter-row', { visible: true })
            // Wait for the default add filter buttons appear
            await driver.page.waitForSelector('.e2e-add-filter-button-repo', { visible: true })
            await driver.page.waitForSelector('.e2e-add-filter-button-file', { visible: true })

            // Add a repo filter
            await driver.page.waitForSelector('.e2e-add-filter-button-repo')
            await driver.page.click('.e2e-add-filter-button-repo')

            // FilterInput is autofocused
            await driver.page.waitForSelector('.filter-input')
            // Search for repo:gorilla in the repo filter chip input
            await driver.page.keyboard.type('gorilla')
            await driver.page.keyboard.press('Enter')
            await driver.assertWindowLocation('/search?q=repo:gorilla&patternType=literal')

            // Edit the filter
            await driver.page.waitForSelector('.filter-input')
            await driver.page.click('.filter-input')
            await driver.page.waitForSelector('.filter-input__input-field')
            await driver.page.keyboard.type('/mux')
            // Press enter to lock in filter
            await driver.page.keyboard.press('Enter')
            // The main query input should be autofocused, so hit enter again to submit
            await driver.assertWindowLocation('/search?q=repo:gorilla/mux&patternType=literal')

            // Add a file filter from search results page
            await driver.page.waitForSelector('.e2e-add-filter-button-file', { visible: true })
            await driver.page.click('.e2e-add-filter-button-file')
            await driver.page.waitForSelector('.filter-input__input-field', { visible: true })
            await driver.page.keyboard.type('README')
            await driver.page.keyboard.press('Enter')
            await driver.page.keyboard.press('Enter')
            await driver.assertWindowLocation('/search?q=repo:gorilla/mux+file:README&patternType=literal')

            // Delete filter
            await driver.page.goto(sourcegraphBaseUrl + '/search?q=repo:gorilla/mux&patternType=literal')
            await driver.page.waitForSelector('.e2e-filter-input__delete-button', { visible: true })
            await driver.page.click('.e2e-filter-input__delete-button')
            await driver.assertWindowLocation('/search?q=&patternType=literal')

            // Test suggestions
            await driver.page.goto(sourcegraphBaseUrl + '/search')
            await driver.page.waitForSelector('.e2e-add-filter-button-repo', { visible: true })
            await driver.page.click('.e2e-add-filter-button-repo')
            await driver.page.waitForSelector('.filter-input', { visible: true })
            await driver.page.waitForSelector('.filter-input__input-field')
            await driver.page.keyboard.type('gorilla')
            await driver.page.waitForSelector('.e2e-filter-input__suggestions')
            await driver.page.waitForSelector('.e2e-suggestion-item')
            await driver.page.keyboard.press('ArrowDown')
            await driver.page.keyboard.press('Enter')
            await driver.page.keyboard.press('Enter')
            await driver.assertWindowLocation('/search?q=repo:%5Egithub%5C.com/gorilla/mux%24&patternType=literal')

            // Test cancelling editing an input with escape key
            await driver.page.click('.filter-input__button-text')
            await driver.page.waitForSelector('.filter-input__input-field')
            await driver.page.keyboard.type('/mux')
            await driver.page.keyboard.press('Escape')
            await driver.page.click('.e2e-search-button')
            await driver.assertWindowLocation('/search?q=repo:%5Egithub%5C.com/gorilla/mux%24&patternType=literal')

            // Test cancelling editing an input by clicking outside close button
            await driver.page.click('.filter-input__button-text')
            await driver.page.waitForSelector('.filter-input__input-field')
            await driver.page.keyboard.type('/mux')
            await driver.page.click('.e2e-search-button')
            await driver.assertWindowLocation('/search?q=repo:%5Egithub%5C.com/gorilla/mux%24&patternType=literal')
        })

        test('Interactive search mode updates query when on searching from directory page', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/github.com/sourcegraph/jsonrpc2')
            await driver.page.waitForSelector('.tree-page__section-search .e2e-query-input')
            await driver.page.click('.tree-page__section-search .e2e-query-input')
            await driver.page.type('.tree-page__section-search .e2e-query-input', 'test')
            await driver.page.click('.tree-page__section-search .e2e-search-button')
            await driver.assertWindowLocation(
                '/search?q=repo:%5Egithub%5C.com/sourcegraph/jsonrpc2%24+test&patternType=literal'
            )
            await driver.page.waitForSelector('.e2e-query-input')
            const queryInputValue = () =>
                driver.page.evaluate(() => {
                    const input = document.querySelector<HTMLInputElement>('.e2e-query-input')
                    return input ? input.value : null
                })
            assert.strictEqual(await queryInputValue(), 'test')
            await driver.page.waitForSelector('.filter-input')
            const filterInputValue = () =>
                driver.page.evaluate(() => {
                    const filterInput = document.querySelector<HTMLButtonElement>('.filter-input__button-text')
                    return filterInput ? filterInput.textContent : null
                })
            assert.strictEqual(await filterInputValue(), 'repo:^github\\.com/sourcegraph/jsonrpc2$')
        })

        test('Interactive search mode filter dropdown and finite-option filter inputs', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/search')
            await driver.page.waitForSelector('.e2e-query-input', { visible: true })
            await driver.page.waitForSelector('.e2e-filter-dropdown')
            await driver.page.type('.e2e-query-input', 'test')
            await driver.page.click('.e2e-filter-dropdown')
            await driver.page.select('.e2e-filter-dropdown', 'fork')
            await driver.page.waitForSelector('.e2e-filter-input-finite-form')
            await driver.page.waitForSelector('.e2e-filter-input-radio-button-no')
            await driver.page.click('.e2e-filter-input-radio-button-no')
            await driver.page.click('.e2e-confirm-filter-button')
            await driver.assertWindowLocation('/search?q=test+fork:no&patternType=literal')
            // Edit filter
            await driver.page.waitForSelector('.filter-input')
            await driver.page.waitForSelector('.e2e-filter-input__button-text-fork')
            await driver.page.click('.e2e-filter-input__button-text-fork')
            await driver.page.waitForSelector('.e2e-filter-input-radio-button-only')
            await driver.page.click('.e2e-filter-input-radio-button-only')
            await driver.page.click('.e2e-confirm-filter-button')
            await driver.assertWindowLocation('/search?q=test+fork:only&patternType=literal')
            // Edit filter by clicking dropdown menu
            await driver.page.waitForSelector('.e2e-filter-dropdown')
            await driver.page.click('.e2e-filter-dropdown')
            await driver.page.select('.e2e-filter-dropdown', 'fork')
            await driver.page.waitForSelector('.e2e-filter-input-finite-form')
            await driver.page.waitForSelector('.e2e-filter-input-radio-button-no')
            await driver.page.click('.e2e-filter-input-radio-button-no')
            await driver.page.click('.e2e-confirm-filter-button')
            await driver.assertWindowLocation('/search?q=test+fork:no&patternType=literal')
        })
    })

    describe('Case sensitivity toggle', () => {
        test('Clicking toggle turns on case sensitivity', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/search')
            await driver.page.waitForSelector('.e2e-query-input', { visible: true })
            await driver.page.waitForSelector('.e2e-case-sensitivity-toggle')
            await driver.page.type('.e2e-query-input', 'test')
            await driver.page.click('.e2e-case-sensitivity-toggle')
            await driver.assertWindowLocation('/search?q=test&patternType=literal&case=yes')
        })

        test('Clicking toggle turns off case sensitivity and removes case= URL parameter', async () => {
            await driver.page.waitForSelector('.e2e-query-input', { visible: true })
            await driver.page.waitForSelector('.e2e-case-sensitivity-toggle')
            await driver.page.click('.e2e-case-sensitivity-toggle')
            await driver.assertWindowLocation('/search?q=test&patternType=literal')
        })
    })

    describe('Structural search toggle', () => {
        test('Clicking toggle turns on structural search', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/search')
            await driver.page.waitForSelector('.e2e-query-input', { visible: true })
            await driver.page.waitForSelector('.e2e-structural-search-toggle')
            await driver.page.type('.e2e-query-input', 'test')
            await driver.page.click('.e2e-structural-search-toggle')
            await driver.assertWindowLocation('/search?q=test&patternType=structural')
        })

        test('Clicking toggle turns on structural search and removes existing patternType parameter', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
            await driver.page.waitForSelector('.e2e-query-input', { visible: true })
            await driver.page.waitForSelector('.e2e-structural-search-toggle')
            await driver.page.click('.e2e-structural-search-toggle')
            await driver.assertWindowLocation('/search?q=test&patternType=structural')
        })

        test('Clicking toggle turns off structural saerch and reverts to default pattern type', async () => {
            await driver.page.waitForSelector('.e2e-query-input', { visible: true })
            await driver.page.waitForSelector('.e2e-structural-search-toggle')
            await driver.page.click('.e2e-structural-search-toggle')
            await driver.assertWindowLocation('/search?q=test&patternType=literal')
        })
    })
})
