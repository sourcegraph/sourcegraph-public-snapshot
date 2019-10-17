import { EndpointPair } from '../../shared/src/platform/context'

export interface PuppeteerComlinkMessage {
    recipient: keyof EndpointPair
    data: any
}

declare global {
    /* eslint-disable no-var */

    /** Provided by the puppeteer, called by us */
    var postToPuppeteer: (data: PuppeteerComlinkMessage) => void

    /** Provided by us, called by Puppeteer */
    var onPuppeteerMessage: (data: PuppeteerComlinkMessage) => void

    /* eslint-enable no-var */
}
