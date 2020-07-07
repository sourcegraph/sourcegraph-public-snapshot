import { createDriverForTest } from '../../../shared/src/testing/driver'
import MockDate from 'mockdate'
import { getConfig } from '../../../shared/src/testing/config'
import assert from 'assert'
import expect from 'expect'
import { ExternalServiceKind } from '../../../shared/src/graphql/schema'
import { describeIntegration } from './helpers'

describeIntegration('Search', ({ initGeneration, describe }) => {
    initGeneration(async () => {
        // Reset date mocking
        MockDate.reset()
        const { gitHubToken, sourcegraphBaseUrl, headless, slowMo, testUserPassword } = getConfig(
            'gitHubToken',
            'sourcegraphBaseUrl',
            'headless',
            'slowMo',
            'testUserPassword'
        )

        // Start browser
        const driver = await createDriverForTest({
            sourcegraphBaseUrl,
            logBrowserConsole: true,
            headless,
            slowMo,
        })
        const repoSlugs = ['gorilla/mux', 'sourcegraph/jsonrpc2']
        await driver.ensureLoggedIn({ username: 'test', password: testUserPassword, email: 'test@test.com' })
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
        return { driver, sourcegraphBaseUrl }
    })

    describe('Interactive search mode', ({ test }) => {
        test('Search mode component appears', async ({ sourcegraphBaseUrl, driver }) => {
            await driver.page.goto(sourcegraphBaseUrl + '/search')
            await driver.page.waitForSelector('.e2e-search-mode-toggle')
            expect(
                await driver.page.evaluate(() => {
                    const toggles = document.querySelectorAll('.e2e-search-mode-toggle')
                    return toggles.length
                })
            ).toBe(1)
        })

        test('Filter buttons', async ({ sourcegraphBaseUrl, driver }) => {
            await driver.page.goto(sourcegraphBaseUrl + '/search')
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
            await driver.assertWindowLocation('/search?q=repo:%22gorilla%22&patternType=literal')

            // Edit the filter
            await driver.page.waitForSelector('.filter-input')
            await driver.page.click('.filter-input')
            await driver.page.waitForSelector('.filter-input__input-field')
            await driver.page.keyboard.type('/mux')
            // Press enter to lock in filter
            await driver.page.keyboard.press('Enter')
            // The main query input should be autofocused, so hit enter again to submit
            await driver.assertWindowLocation('/search?q=repo:%22gorilla/mux%22&patternType=literal')

            // Add a file filter from search results page
            await driver.page.waitForSelector('.e2e-add-filter-button-file', { visible: true })
            await driver.page.click('.e2e-add-filter-button-file')
            await driver.page.waitForSelector('.filter-input__input-field', { visible: true })
            await driver.page.keyboard.type('README')
            await driver.page.keyboard.press('Enter')
            await driver.page.keyboard.press('Enter')
            await driver.assertWindowLocation('/search?q=repo:%22gorilla/mux%22+file:%22README%22&patternType=literal')

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
            await driver.assertWindowLocation(
                '/search?q=repo:%22%5Egithub%5C%5C.com/gorilla/mux%24%22&patternType=literal'
            )

            // Test cancelling editing an input with escape key
            await driver.page.click('.filter-input__button-text')
            await driver.page.waitForSelector('.filter-input__input-field')
            await driver.page.keyboard.type('/mux')
            await driver.page.keyboard.press('Escape')
            await driver.page.click('.e2e-search-button')
            await driver.assertWindowLocation(
                '/search?q=repo:%22%5Egithub%5C%5C.com/gorilla/mux%24%22&patternType=literal'
            )

            // Test cancelling editing an input by clicking outside close button
            await driver.page.click('.filter-input__button-text')
            await driver.page.waitForSelector('.filter-input__input-field')
            await driver.page.keyboard.type('/mux')
            await driver.page.click('.e2e-search-button')
            await driver.assertWindowLocation(
                '/search?q=repo:%22%5Egithub%5C%5C.com/gorilla/mux%24%22&patternType=literal'
            )
        })

        test('Updates query when searching from directory page', async ({ sourcegraphBaseUrl, driver }) => {
            await driver.page.goto(sourcegraphBaseUrl + '/github.com/sourcegraph/jsonrpc2')
            await driver.page.waitForSelector('.filter-input')
            const filterInputValue = () =>
                driver.page.evaluate(() => {
                    const filterInput = document.querySelector<HTMLButtonElement>('.filter-input__button-text')
                    return filterInput ? filterInput.textContent : null
                })
            assert.strictEqual(await filterInputValue(), 'repo:^github\\.com/sourcegraph/jsonrpc2$')
        })

        test('Filter dropdown and finite-option filter inputs', async ({ sourcegraphBaseUrl, driver }) => {
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
            await driver.assertWindowLocation('/search?q=fork:%22no%22+test&patternType=literal')
            // Edit filter
            await driver.page.waitForSelector('.filter-input')
            await driver.page.waitForSelector('.e2e-filter-input__button-text-fork')
            await driver.page.click('.e2e-filter-input__button-text-fork')
            await driver.page.waitForSelector('.e2e-filter-input-radio-button-only')
            await driver.page.click('.e2e-filter-input-radio-button-only')
            await driver.page.click('.e2e-confirm-filter-button')
            await driver.assertWindowLocation('/search?q=fork:%22only%22+test&patternType=literal')
            // Edit filter by clicking dropdown menu
            await driver.page.waitForSelector('.e2e-filter-dropdown')
            await driver.page.click('.e2e-filter-dropdown')
            await driver.page.select('.e2e-filter-dropdown', 'fork')
            await driver.page.waitForSelector('.e2e-filter-input-finite-form')
            await driver.page.waitForSelector('.e2e-filter-input-radio-button-no')
            await driver.page.click('.e2e-filter-input-radio-button-no')
            await driver.page.click('.e2e-confirm-filter-button')
            await driver.assertWindowLocation('/search?q=fork:%22no%22+test&patternType=literal')
        })
    })

    describe('Case sensitivity toggle', ({ test }) => {
        test('Clicking toggle turns on case sensitivity', async ({ sourcegraphBaseUrl, driver }) => {
            await driver.page.goto(sourcegraphBaseUrl + '/search')
            await driver.page.waitForSelector('.e2e-query-input', { visible: true })
            await driver.page.waitForSelector('.e2e-case-sensitivity-toggle')
            await driver.page.type('.e2e-query-input', 'test')
            await driver.page.click('.e2e-case-sensitivity-toggle')
            await driver.assertWindowLocation('/search?q=test&patternType=literal&case=yes')
        })

        test('Clicking toggle turns off case sensitivity and removes case= URL parameter', async ({
            sourcegraphBaseUrl,
            driver,
        }) => {
            await driver.page.goto(sourcegraphBaseUrl + '/search?q=test&patternType=literal&case=yes')
            await driver.page.waitForSelector('.e2e-query-input', { visible: true })
            await driver.page.waitForSelector('.e2e-case-sensitivity-toggle')
            await driver.page.click('.e2e-case-sensitivity-toggle')
            await driver.assertWindowLocation('/search?q=test&patternType=literal')
        })
    })

    describe('Structural search toggle', ({ test }) => {
        test('Clicking toggle turns on structural search', async ({ sourcegraphBaseUrl, driver }) => {
            await driver.page.goto(sourcegraphBaseUrl + '/search')
            await driver.page.waitForSelector('.e2e-query-input', { visible: true })
            await driver.page.waitForSelector('.e2e-structural-search-toggle')
            await driver.page.type('.e2e-query-input', 'test')
            await driver.page.click('.e2e-structural-search-toggle')
            await driver.assertWindowLocation('/search?q=test&patternType=structural')
        })

        test('Clicking toggle turns on structural search and removes existing patternType parameter', async ({
            sourcegraphBaseUrl,
            driver,
        }) => {
            await driver.page.goto(sourcegraphBaseUrl + '/search?q=test&patternType=regexp')
            await driver.page.waitForSelector('.e2e-query-input', { visible: true })
            await driver.page.waitForSelector('.e2e-structural-search-toggle')
            await driver.page.click('.e2e-structural-search-toggle')
            await driver.assertWindowLocation('/search?q=test&patternType=structural')
        })

        test('Clicking toggle turns off structural saerch and reverts to default pattern type', async ({
            sourcegraphBaseUrl,
            driver,
        }) => {
            await driver.page.goto(sourcegraphBaseUrl + '/search?q=test&patternType=structural')
            await driver.page.waitForSelector('.e2e-query-input', { visible: true })
            await driver.page.waitForSelector('.e2e-structural-search-toggle')
            await driver.page.click('.e2e-structural-search-toggle')
            await driver.assertWindowLocation('/search?q=test&patternType=literal')
        })
    })
})
