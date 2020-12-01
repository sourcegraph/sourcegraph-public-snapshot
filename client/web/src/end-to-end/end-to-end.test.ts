import { describe, test, before, beforeEach, after } from 'mocha'
import { afterEachSaveScreenshotIfFailed } from '../../../shared/src/testing/screenshotReporter'
import { afterEachRecordCoverage } from '../../../shared/src/testing/coverage'
import { createDriverForTest, Driver } from '../../../shared/src/testing/driver'
import MockDate from 'mockdate'
import { ExternalServiceKind } from '../../../shared/src/graphql/schema'
import { getConfig } from '../../../shared/src/testing/config'
import expect from 'expect'

const { gitHubToken, sourcegraphBaseUrl } = getConfig('gitHubToken', 'sourcegraphBaseUrl')

describe('e2e test suite', () => {
    let driver: Driver

    before(async function () {
        // Cloning the repositories takes ~1 minute, so give initialization 2
        // minutes instead of 1 (which would be inherited from
        // `jest.setTimeout(1 * 60 * 1000)` above).
        this.timeout(5 * 60 * 1000)

        // Reset date mocking
        MockDate.reset()

        const config = getConfig('headless', 'slowMo', 'testUserPassword')

        // Start browser
        driver = await createDriverForTest({
            sourcegraphBaseUrl,
            logBrowserConsole: true,
            ...config,
        })
        const clonedRepoSlugs = [
            'sourcegraph/java-langserver',
            'gorilla/mux',
            'gorilla/securecookie',
            'sourcegraph/jsonrpc2',
            'sourcegraph/go-diff',
            'sourcegraph/appdash',
            'sourcegraph/sourcegraph-typescript',
            'sourcegraph-testing/automation-e2e-test',
            'sourcegraph/e2e-test-private-repository',
        ]
        const alwaysCloningRepoSlugs = ['sourcegraphtest/AlwaysCloningTest']
        await driver.ensureLoggedIn({ username: 'test', password: config.testUserPassword, email: 'test@test.com' })
        await driver.resetUserSettings()
        await driver.ensureHasExternalService({
            kind: ExternalServiceKind.GITHUB,
            displayName: 'test-test-github',
            config: JSON.stringify({
                url: 'https://github.com',
                token: gitHubToken,
                repos: clonedRepoSlugs.concat(alwaysCloningRepoSlugs),
            }),
            ensureRepos: clonedRepoSlugs.map(slug => `github.com/${slug}`),
            alwaysCloning: alwaysCloningRepoSlugs.map(slug => `github.com/${slug}`),
        })
    })

    after('Close browser', () => driver?.close())

    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEachRecordCoverage(() => driver)

    beforeEach(async () => {
        if (driver) {
            // Clear local storage to reset sidebar selection (files or tabs) for each test
            await driver.page.evaluate(() => {
                localStorage.setItem('repo-revision-sidebar-last-tab', 'files')
            })

            await driver.resetUserSettings()
        }
    })

    describe('Core functionality', () => {
        test('Check settings are saved and applied', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/users/test/settings')
            await driver.page.waitForSelector('.test-settings-file .monaco-editor')

            const message = 'A wild notice appears!'
            await driver.replaceText({
                selector: '.test-settings-file .monaco-editor',
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
            await driver.page.click('.test-settings-file .test-save-toolbar-save')
            await driver.page.waitForSelector('.test-global-alert .notices .global-alerts__alert', { visible: true })
            await driver.page.evaluate((message: string) => {
                const element = document.querySelector<HTMLElement>('.test-global-alert .notices .global-alerts__alert')
                if (!element) {
                    throw new Error('No .test-global-alert .notices .global-alerts__alert element found')
                }
                if (!element.textContent?.includes(message)) {
                    throw new Error(`Expected "${message}" message, but didn't find it`)
                }
            }, message)
        })

        test('Check allowed usernames', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/users/test/settings/profile')
            await driver.page.waitForSelector('.test-UserProfileFormFields-username')

            const name = 'alice.bob-chris-'

            await driver.replaceText({
                selector: '.test-UserProfileFormFields-username',
                newText: name,
                selectMethod: 'selectall',
            })

            await driver.page.click('#test-EditUserProfileForm__save')
            await driver.page.waitForSelector('.test-EditUserProfileForm__success', { visible: true })

            await driver.page.goto(sourcegraphBaseUrl + `/users/${name}/settings/profile`)
            await driver.replaceText({
                selector: '.test-UserProfileFormFields-username',
                newText: 'test',
                selectMethod: 'selectall',
            })

            await driver.page.click('#test-EditUserProfileForm__save')
            await driver.page.waitForSelector('.test-EditUserProfileForm__success', { visible: true })
        })
    })

    describe('External services', () => {
        test('External service add, edit, delete', async () => {
            const displayName = 'test-github-test-2'
            await driver.ensureHasExternalService({
                kind: ExternalServiceKind.GITHUB,
                displayName,
                config:
                    '{"url": "https://github.myenterprise.com", "token": "initial-token", "repositoryQuery": ["none"]}',
            })
            await driver.page.goto(sourcegraphBaseUrl + '/site-admin/external-services')
            await (
                await driver.page.waitForSelector(
                    `[data-test-external-service-name="${displayName}"] .test-edit-external-service-button`
                )
            ).click()

            // Type in a new external service configuration.
            await driver.replaceText({
                selector: '.test-external-service-editor .monaco-editor',
                newText:
                    '{"url": "https://github.myenterprise.com", "token": "second-token", "repositoryQuery": ["none"]}',
                selectMethod: 'selectall',
                enterTextMethod: 'paste',
            })
            await driver.page.click('.test-update-external-service-button')
            // Must wait for the operation to complete, or else a "Discard changes?" dialog will pop up
            await driver.page.waitForSelector('.test-update-external-service-button:not([disabled])', { visible: true })

            await (
                await driver.page.waitForSelector('.list-group-item[href="/site-admin/external-services"]', {
                    visible: true,
                })
            ).click()

            await Promise.all([
                driver.acceptNextDialog(),
                (
                    await driver.page.waitForSelector(
                        '[data-test-external-service-name="test-github-test-2"] .test-delete-external-service-button',
                        { visible: true }
                    )
                ).click(),
            ])

            await driver.page.waitFor(
                () => !document.querySelector('[data-test-external-service-name="test-github-test-2"]')
            )
        })

        test('External service repositoryPathPattern', async () => {
            const repo = 'sourcegraph/go-blame' // Tiny repo, fast to clone
            const repositoryPathPattern = 'foobar/{host}/{nameWithOwner}'
            const slug = `github.com/${repo}`
            const pathPatternSlug = `foobar/github.com/${repo}`

            const config = {
                kind: ExternalServiceKind.GITHUB,
                displayName: 'test-test-github-repoPathPattern',
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
                displayName: 'test-aws-code-commit',
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
            const blob = (await (
                await driver.page.waitFor(() => document.querySelector<HTMLElement>('.test-repo-blob')?.textContent)
            ).jsonValue()) as string | null

            expect(blob).toBe('README\n\nchange')
        })

        const bbsURL = process.env.BITBUCKET_SERVER_URL
        const bbsToken = process.env.BITBUCKET_SERVER_TOKEN
        const bbsUsername = process.env.BITBUCKET_SERVER_USERNAME

        const testIfBBSCredentialsSet = bbsURL && bbsToken && bbsUsername ? test : test.skip.bind(test)

        testIfBBSCredentialsSet('Bitbucket Server', async () => {
            await driver.ensureHasExternalService({
                kind: ExternalServiceKind.BITBUCKETSERVER,
                displayName: 'test-bitbucket-server',
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
            const blob = (await (
                await driver.page.waitFor(() => document.querySelector<HTMLElement>('.test-repo-blob')?.textContent)
            ).jsonValue()) as string | null

            expect(blob).toBe('language: go\ngo: \n - 1.x\n\nscript:\n - go test -race -v ./...')
        })
    })

    describe('Saved searches', () => {
        test('Save search from search results page', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/search?q=test')
            await driver.page.waitForSelector('.test-save-search-link', { visible: true })
            await driver.page.click('.test-save-search-link')
            await driver.page.waitForSelector('.test-saved-search-modal')
            await driver.page.waitForSelector('.test-saved-search-modal-save-button')
            await driver.page.click('.test-saved-search-modal-save-button')
            await driver.assertWindowLocation('/users/test/searches/add?query=test&patternType=regexp')

            await driver.page.waitForSelector('.test-saved-search-form-input-description', { visible: true })
            await driver.page.click('.test-saved-search-form-input-description')
            await driver.page.keyboard.type('test query')
            await driver.page.waitForSelector('.test-saved-search-form-submit-button', { visible: true })
            await driver.page.click('.test-saved-search-form-submit-button')
            await driver.assertWindowLocation('/users/test/searches')

            const nodes = await driver.page.evaluate(
                () => document.querySelectorAll('.test-saved-search-list-page-row').length
            )
            expect(nodes).toEqual(1)

            expect(
                await driver.page.evaluate(
                    () => document.querySelector('.test-saved-search-list-page-row-title')!.textContent
                )
            ).toEqual('test query')
        })
        test('Delete saved search', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/users/test/searches')
            await driver.page.waitForSelector('.test-delete-saved-search-button', { visible: true })
            driver.page.on('dialog', async dialog => {
                await dialog.accept()
            })
            await driver.page.click('.test-delete-saved-search-button')
            await driver.page.waitFor(() => !document.querySelector('.test-saved-search-list-page-row'))
            const nodes = await driver.page.evaluate(
                () => document.querySelectorAll('.test-saved-search-list-page-row').length
            )
            expect(nodes).toEqual(0)
        })
        test('Save search from saved searches page', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/users/test/searches')
            await driver.page.waitForSelector('.test-add-saved-search-button', { visible: true })
            await driver.page.click('.test-add-saved-search-button')
            await driver.assertWindowLocation('/users/test/searches/add')

            await driver.page.waitForSelector('.test-saved-search-form-input-description', { visible: true })
            await driver.page.click('.test-saved-search-form-input-description')
            await driver.page.keyboard.type('test query 2')

            await driver.page.waitForSelector('.test-saved-search-form-input-query', { visible: true })
            await driver.page.click('.test-saved-search-form-input-query')
            await driver.page.keyboard.type('test patternType:literal')

            await driver.page.waitForSelector('.test-saved-search-form-submit-button', { visible: true })
            await driver.page.click('.test-saved-search-form-submit-button')
            await driver.assertWindowLocation('/users/test/searches')

            const nodes = await driver.page.evaluate(
                () => document.querySelectorAll('.test-saved-search-list-page-row').length
            )
            expect(nodes).toEqual(1)

            expect(
                await driver.page.evaluate(
                    () => document.querySelector('.test-saved-search-list-page-row-title')!.textContent
                )
            ).toEqual('test query 2')
        })
        test('Edit saved search', async () => {
            await driver.page.goto(sourcegraphBaseUrl + '/users/test/searches')
            await driver.page.waitForSelector('.test-edit-saved-search-button', { visible: true })
            await driver.page.click('.test-edit-saved-search-button')

            await driver.page.waitForSelector('.test-saved-search-form-input-description', { visible: true })
            await driver.page.click('.test-saved-search-form-input-description')
            await driver.page.keyboard.type(' edited')

            await driver.page.waitForSelector('.test-saved-search-form-submit-button', { visible: true })
            await driver.page.click('.test-saved-search-form-submit-button')
            await driver.page.goto(sourcegraphBaseUrl + '/users/test/searches')
            await driver.page.waitForSelector('.test-saved-search-list-page-row-title')

            expect(
                await driver.page.evaluate(
                    () => document.querySelector('.test-saved-search-list-page-row-title')!.textContent
                )
            ).toEqual('test query 2 edited')
        })
    })
})
