import { isBackground, isInPage } from '../../shared/context'
import { BackgroundMessageHandlers } from './types'

const messageSender = <T extends keyof BackgroundMessageHandlers>(type: T): BackgroundMessageHandlers[T] => (
    payload?: any
) => {
    if (isBackground) {
        throw new Error('Tried to call background page function from background page itself')
    }
    if (isInPage) {
        throw new Error('Tried to call background page function from in-page integration')
    }
    return browser.runtime.sendMessage({ type, payload })
}

/**
 * Functions that can be invoked from content scripts that will be executed in the background page.
 */
export const background: BackgroundMessageHandlers = {
    createBlobURL: messageSender('createBlobURL'),
    openOptionsPage: messageSender('openOptionsPage'),
    requestGraphQL: messageSender('requestGraphQL'),
}
