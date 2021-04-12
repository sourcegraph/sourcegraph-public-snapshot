import puppeteer from 'puppeteer'
import signale from 'signale'

import { PUPPETEER_BROWSER_REVISION } from '../src/testing/puppeteer-browser-revision'

async function main(): Promise<void> {
    const browserName = process.env.BROWSER || 'chrome'
    if (browserName !== 'chrome' && browserName !== 'firefox') {
        signale.error(`Puppeteer browser must be "chrome" or "firefox", but got: "${browserName}"`)
        process.exit(1)
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

main().catch(error => {
    console.error(error)
    process.exit(1)
})
