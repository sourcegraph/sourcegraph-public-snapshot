import puppeteer from 'puppeteer'
import signale from 'signale'
import { PUPPETEER_BROWSER_REVISION } from '../src/testing/driver'

async function downloadPuppeteerBrowser(): Promise<void> {
    const browserName = process.env.BROWSER || 'chrome'
    if (browserName !== 'chrome' && browserName !== 'firefox') {
        throw new Error(`Puppeteer browser must be "chrome" or "firefox", but got: "${browserName}"`)
    }
    const browserFetcher = puppeteer.createBrowserFetcher({ product: browserName })
    const revision = PUPPETEER_BROWSER_REVISION[browserName]
    const revisionInfo = browserFetcher.revisionInfo(revision)
    if (!revisionInfo.local) {
        signale.await(`Puppeteer browser: downloading ${browserName} revision ${revision}.`)
        const revisionInfo = await browserFetcher.download(revision)
        signale.success(`Done downloading browser to: ${revisionInfo.executablePath}`)
    } else {
        signale.success(`Puppeteer browser: found existing ${browserName} revision ${revision}, skipping download.`)
    }
}

downloadPuppeteerBrowser().catch(error => {
    console.error(error)
    process.exit(1)
})
