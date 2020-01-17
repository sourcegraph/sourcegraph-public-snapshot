import expect from 'expect'
import { Driver } from '../../../../shared/src/e2e/driver'
import { Config } from '../../../../shared/src/e2e/config'
import { ElementHandle } from 'puppeteer'

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
                const tokenEl = await findTokenElement(driver, line, token)

                // Check hover
                await tokenEl.hover()
                await waitForHover(driver, expectedHoverContains)

                // Check click
                await clickOnEmptyPartOfCodeView(driver)
                await tokenEl.click()
                await waitForHover(driver, expectedHoverContains)

                // Find-references
                if (expectedReferences) {
                    await clickOnEmptyPartOfCodeView(driver)
                    await tokenEl.hover()
                    await waitForHover(driver, expectedHoverContains)
                    await (await driver.findElementWithText('Find references')).click()

                    await driver.page.waitForSelector('.e2e-search-result')
                    const refLinks = await collectLinks(driver)
                    for (const expectedReference of expectedReferences) {
                        expect(refLinks).toContain(expectedReference)
                    }
                    await clickOnEmptyPartOfCodeView(driver)
                }

                // Go-to-definition

                await clickOnEmptyPartOfCodeView(driver)
                await tokenEl.hover()
                await waitForHover(driver, expectedHoverContains)
                await (await driver.findElementWithText('Go to definition')).click()

                if (Array.isArray(expectedDefinition)) {
                    await driver.page.waitForSelector('.e2e-search-result')
                    const defLinks = await collectLinks(driver)
                    for (const definition of expectedDefinition) {
                        expect(defLinks).toContain(definition)
                    }
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

async function collectLinks(driver: Driver): Promise<Set<string>> {
    const panelTabTitles = await getPanelTabTitles(driver)
    if (panelTabTitles.length === 0) {
        return new Set(await collectVisibleLinks(driver))
    }

    const links = new Set<string>()
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

async function getPanelTabTitles(driver: Driver): Promise<string[]> {
    return (
        await Promise.all(
            (await driver.page.$$('.hierarchical-locations-view > div:nth-child(1) span[title]')).map(e =>
                e.evaluate(e => e.getAttribute('title') || '')
            )
        )
    ).map(normalizeWhitespace)
}

function collectVisibleLinks(driver: Driver): Promise<string[]> {
    return driver.page.evaluate(() =>
        Array.from(document.querySelectorAll<HTMLElement>('.e2e-search-result')).flatMap(e =>
            Array.from(e.querySelectorAll<HTMLElement>('a[href]')).map(a => a.getAttribute('href') || '')
        )
    )
}

async function clickOnEmptyPartOfCodeView(driver: Driver): Promise<void> {
    await driver.page.click('.e2e-blob tr:nth-child(1) .line')
    await driver.page.waitForFunction(() => document.querySelectorAll('.e2e-tooltip-go-to-definition').length === 0)
}

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

async function waitForHover(driver: Driver, expectedHoverContains: string): Promise<void> {
    await driver.page.waitForSelector('.e2e-tooltip-go-to-definition')
    await driver.page.waitForSelector('.e2e-tooltip-content')
    expect(normalizeWhitespace(await getTooltip(driver))).toContain(normalizeWhitespace(expectedHoverContains))
}

function normalizeWhitespace(s: string): string {
    return s.replace(/\s+/g, ' ')
}
