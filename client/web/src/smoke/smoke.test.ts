import { createDriverForTest, Driver } from '@sourcegraph/shared/src/testing/driver'

describe('Search test', () => {
    let driver: Driver
    before(async () => {
        driver = await createDriverForTest()
    })
    after(() => driver?.close())

    it('Clicking search button executes search', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
        await driver.page.waitForSelector('.test-search-button', { visible: true })
        // Note: Delay added because this test has been intermittently failing without it. Monaco search bar may drop events if it gets too many too fast.
        await driver.page.keyboard.type('test', { delay: 50 })
        await driver.page.click('.test-search-button')
        await driver.assertWindowLocation('/search?q=context:global+test&patternType=literal')
        await driver.page.waitForSelector('.test-search-result', { visible: true, timeout: 5000 })
    })
})
