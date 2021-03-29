/**
 * The browser revision used by Puppeteer, which is downloaded by BrowserFetcher.
 */
export const PUPPETEER_BROWSER_REVISION: { chrome: string; firefox: string } = {
    chrome: '818858',

    // TODO: When support is added for downloading Firefox revisions, pin this
    // to an exact revision.
    firefox: 'latest',
}
