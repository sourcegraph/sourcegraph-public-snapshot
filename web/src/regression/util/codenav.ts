import { Driver } from '../../../../shared/src/e2e/driver'
import { Config } from '../../../../shared/src/e2e/config'
import { ElementHandle } from 'puppeteer'

async function setGlobalLSIFSetting(
    driver: Driver,
    config: Pick<Config, 'sourcegraphBaseUrl'>,
    enabled: boolean
): Promise<void> {
    await driver.page.goto(`${config.sourcegraphBaseUrl}/site-admin/global-settings`)
    const globalSettings = `{"codeIntel.lsif": ${enabled}}`
    await driver.replaceText({
        selector: '.monaco-editor',
        newText: globalSettings,
        selectMethod: 'keyboard',
        enterTextMethod: 'type',
    })
    await (
        await driver.findElementWithText('Save changes', {
            selector: 'button',
            wait: { timeout: 500 },
        })
    ).click()
}

export const enableLSIF = (driver: Driver, config: Pick<Config, 'sourcegraphBaseUrl'>): Promise<void> =>
    setGlobalLSIFSetting(driver, config, true)

export const disableLSIF = (driver: Driver, config: Pick<Config, 'sourcegraphBaseUrl'>): Promise<void> =>
    setGlobalLSIFSetting(driver, config, false)

export interface TestCase {
    repoRev: string
    files: {
        path: string
        locations: {
            line: number
            token: string
            expectedHoverContains: string
            expectedDefinition: string | string[]
            expectedReferences?: string[]
        }[]
    }[]
}

export async function testCodeNavigation(
    driver: Driver,
    config: Pick<Config, 'sourcegraphBaseUrl'>,
    testCases: TestCase[]
): Promise<void> {
    for (const { repoRev, files } of testCases) {
        for (const { path, locations } of files) {
            await driver.page.goto(config.sourcegraphBaseUrl + `/${repoRev}/-/blob${path}`)
            await driver.page.waitForSelector('.e2e-blob')
            for (const { line, token, expectedHoverContains, expectedDefinition, expectedReferences } of locations) {
                const { tokenEl, xpathQuery } = await findTokenElement(driver, line, token)
                if (tokenEl.length === 0) {
                    throw new Error(
                        `did not find token ${JSON.stringify(token)} on page. XPath query was: ${xpathQuery}`
                    )
                }

                // Check hover and click
                await tokenEl[0].hover()
                await waitForHover(driver, expectedHoverContains)
                const { tokenEl: emptyTokenEl } = await findTokenElement(driver, line, '')
                await emptyTokenEl[0].hover()
                await driver.page.waitForFunction(
                    () => document.querySelectorAll('.e2e-tooltip-go-to-definition').length === 0
                )
                await tokenEl[0].click()
                await waitForHover(driver, expectedHoverContains)

                // Find-references
                if (expectedReferences) {
                    await (await driver.findElementWithText('Find references')).click()
                    await driver.page.waitForSelector('.e2e-search-result')
                    const refLinks = await collectLinks(driver, '.e2e-search-result')
                    for (const expectedReference of expectedReferences) {
                        expect(refLinks.includes(expectedReference)).toBeTruthy()
                    }
                    await clickOnEmptyPartOfCodeView(driver)
                }

                // Go-to-definition
                await (await driver.findElementWithText('Go to definition')).click()
                if (Array.isArray(expectedDefinition)) {
                    await driver.page.waitForSelector('.e2e-search-result')
                    const defLinks = await collectLinks(driver, '.e2e-search-result')
                    expect(expectedDefinition.every(l => defLinks.includes(l))).toBeTruthy()
                } else {
                    await driver.page.waitForFunction(
                        defURL => document.location.href.endsWith(defURL),
                        { timeout: 2000 },
                        expectedDefinition
                    )
                    await driver.page.goBack()
                }

                await driver.page.keyboard.press('Escape')
            }
        }
    }
}

async function getTooltip(driver: Driver): Promise<string> {
    return driver.page.evaluate(() => (document.querySelector('.e2e-tooltip-content') as HTMLElement).innerText)
}

function collectLinks(driver: Driver, selector: string): Promise<string[]> {
    return driver.page.evaluate(selector => {
        const links: string[] = []
        document.querySelectorAll<HTMLElement>(selector).forEach(e => {
            e.querySelectorAll<HTMLElement>('a[href]').forEach(a => {
                const href = a.getAttribute('href')
                if (href) {
                    links.push(href)
                }
            })
        })
        return links
    }, selector)
}

async function clickOnEmptyPartOfCodeView(driver: Driver): Promise<ElementHandle<Element>[]> {
    return driver.page.$x('//*[contains(@class, "e2e-blob")]//tr[1]//*[text() = ""]')
}

async function findTokenElement(
    driver: Driver,
    line: number,
    token: string
): Promise<{ tokenEl: ElementHandle<Element>[]; xpathQuery: string }> {
    const lineQuery = `//*[contains(@class, "e2e-blob")]//tr[${line}]`
    const xpathQuery = `${lineQuery}//*[normalize-space(text()) = ${JSON.stringify(token)}]`

    const lineEl = await driver.page.$x(lineQuery)
    if (lineEl.length === 0) {
        throw new Error(`line ${line} does not exist`)
    }

    // Force tokenization if the line requires it. If we don't do this then some tokens
    // will not be found as they have additional punctuation next to it (eg. `Type{`).
    await lineEl[0].hover()

    // If there's an open toast, close it. If the toast remains open and our target
    // identifier happens to be hidden by it, we won't be able to select the correct
    // token. This condition was reproducible in the codenav test for `StdioLogger`.
    const closeToast = await driver.page.$('.e2e-close-toast')
    if (closeToast) {
        await closeToast.click()
    }

    return {
        tokenEl: await driver.page.$x(xpathQuery),
        xpathQuery,
    }
}

function normalizeWhitespace(s: string): string {
    return s.replace(/\s+/g, ' ')
}

async function waitForHover(driver: Driver, expectedHoverContains: string): Promise<void> {
    await driver.page.waitForSelector('.e2e-tooltip-go-to-definition')
    await driver.page.waitForSelector('.e2e-tooltip-content')
    expect(normalizeWhitespace(await getTooltip(driver))).toContain(normalizeWhitespace(expectedHoverContains))
}
