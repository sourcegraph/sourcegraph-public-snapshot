interface Message {
    type: string
    payload?: any
}

type ResponseHandler = (res?: any) => void
type MessageHandler = (
    message: Message,
    sender: chrome.runtime.MessageSender,
    sendResponse?: (response?: any) => void
) => void

// Doesn't need to be the most robust solution and we also don't want a heavy hashing function
// https://gist.github.com/gordonbrander/2230317#gistcomment-1713405
const generateUniqueID = () => {
    const id =
        Date.now().toString(36) +
        Math.random()
            .toString(36)
            .substr(2, 5)

    return id.toUpperCase()
}

const safari = window.safari

/**
 * SafariMessager acts as the message passing layer for Safari so that we can use
 * runtime.onMessage and runtime.sendMessage to talk between the content script and background
 * as easily and as intuitive as in chrome and firefox.
 *
 * SafariMessager gets built and behaves differently for each of the environments
 * it will be run in:
 *
 * Content Script: Dispatches event to background and waits for a response and calls the callback with that response
 * if there is one
 *
 * Options Script: Acts as a proxy to the background since it has access to the background
 *
 * Background: Listens for messages and passes the message to each handler declared with runtime.onMessage
 * and then dispatches an event with the response.
 *
 * Non-safari: noops
 */
export class SafariMessager {
    private contentTarget: SafariContentWebPage | null = null
    private optionsTarget: SafariExtensionPopover | null = null
    private backgroundTarget: SafariExtensionGlobalPage | null = null
    private responseHandlers = new Map<string, ResponseHandler>()
    private messageHandlers: MessageHandler[] = []

    constructor() {
        if (safari && !safari.application) {
            this.contentTarget = safari.self as SafariContentWebPage
            this.contentTarget.addEventListener('message', this.handle, false)
        } else if (safari && safari.application) {
            if ((safari.self as any).identifier === 'com.sourcegraph.options') {
                this.optionsTarget = safari.self as SafariExtensionPopover
            } else {
                this.backgroundTarget = safari.self as SafariExtensionGlobalPage

                safari.application.addEventListener('message', this.handleBackGroundMessages, false)
            }
        }
    }

    private handle = (event: SafariExtensionMessageEvent): void => {
        if (this.contentTarget) {
            const handler = this.responseHandlers.get(event.name)

            if (handler) {
                handler(event.message)
            }
        }
    }

    private handleBackGroundMessages = (event: SafariExtensionMessageEvent): void => {
        const callback = (res: any) => {
            const target = event.target as SafariBrowserTab

            target.page.dispatchMessage(event.message._id, res)
        }

        this.executeAll(event.message as Message, callback)
    }

    private executeAll(message: Message, cb?: ResponseHandler): void {
        for (const handler of this.messageHandlers) {
            handler(message, {}, cb)
        }
    }

    public send(message: Message, cb?: ResponseHandler): void {
        if (this.contentTarget) {
            const id = generateUniqueID()

            this.contentTarget.tab.dispatchMessage(message.type, {
                ...message,
                _id: id,
            })

            if (cb) {
                this.responseHandlers.set(id, cb)
            }
        } else if (this.optionsTarget) {
            // For options, we just proxy to the background script because most things
            // here need to be ran in the bg.
            //
            // TODO: fix safari typings to support this
            const ext = safari.extension as any
            ext.globalPage.contentWindow.safariMessager.send(message, cb)
        } else if (this.backgroundTarget) {
            this.executeAll(message, cb)
        }
    }

    public onMessage(handler: MessageHandler): void {
        this.messageHandlers.push(handler)
    }
}

const safariMessager = new SafariMessager()

// The Safari global page(the background script) needs to be exposed via the global property
// so that we can access it from the options context. This is required to simulate message passing.
if (safari && safari.application && (safari.self as any).identifier !== 'com.sourcegraph.options') {
    window.safariMessager = safariMessager
}

export default safariMessager
