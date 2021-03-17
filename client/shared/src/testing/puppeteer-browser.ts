import puppeteer, { RevisionInfo } from 'puppeteer'

/**
 * Promise that resolves to the downloaded browser revision, resolving once the download is finished.
 */
let cachedPuppeteerBrowserRevisionInfoPromise: Promise<RevisionInfo> | undefined

/**
 * Get the revision info for the given browser revision (including the
 * executable path), downloading a new copy of the browser if needed. If called
 * multiple times in succession, it will start only one download and will return
 * a promise that resolves when the download is complete.
 */
export function getPuppeteerBrowser(browserName: string, revision: string): Promise<RevisionInfo> {
    if (cachedPuppeteerBrowserRevisionInfoPromise) {
        console.log('(debug) Using existing cached puppeteer browser promise')
        return cachedPuppeteerBrowserRevisionInfoPromise
    }
    const browserFetcher = puppeteer.createBrowserFetcher({ product: browserName })
    const revisionInfo = browserFetcher.revisionInfo(revision)
    if (revisionInfo.local) {
        console.log(`Found existing browser for Puppeteer: ${browserName} revision ${revision}`)
        cachedPuppeteerBrowserRevisionInfoPromise = Promise.resolve(revisionInfo)
        return cachedPuppeteerBrowserRevisionInfoPromise
    }

    console.log(`Downloading browser for Puppeteer: ${browserName} revision ${revision}`)
    cachedPuppeteerBrowserRevisionInfoPromise = browserFetcher.download(revision)
    return cachedPuppeteerBrowserRevisionInfoPromise
}
