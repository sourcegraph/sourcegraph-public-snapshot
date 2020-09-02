import { Polly, Request as PollyRequest } from '@pollyjs/core'
import type * as Puppeteer from 'puppeteer'
import { intercept, patterns, Interceptor } from 'puppeteer-interceptor'
import Protocol from 'devtools-protocol'
import PollyAdapter from '@pollyjs/adapter'

interface PollyResponse {
    statusCode: number
    headers: Record<string, string>
    body: string
}

interface PollyRequestArguments {
    requestArguments: { request: Protocol.Network.Request }
}

const puppeteerToCDPPatterns: Record<Puppeteer.ResourceType, keyof typeof patterns> = {
    document: 'Document',
    eventsource: 'EventSource',
    fetch: 'Fetch',
    font: 'Font',
    image: 'Image',
    manifest: 'Manifest',
    media: 'Media',
    other: 'Other',
    script: 'Script',
    stylesheet: 'Stylesheet',
    texttrack: 'TextTrack',
    websocket: 'WebSocket',
    xhr: 'XHR',
}

/**
 * A Puppeteer adapter for Polly that supports all request resource types.
 *
 * TODO: upstream this?
 *
 * Polly's own Puppeteer adapter hangs when attempting to capture the page's initial document
 * (requestResourceType==='document'). See https://github.com/Netflix/pollyjs/issues/121
 *
 * Its very complex internal flow makes it hard to modify/fix. The internal flow of this adapter is much simpler,
 * and handles all request resource types.
 *
 */
export class PuppeteerAdapter extends PollyAdapter {
    /**
     * The puppeteer Page this adapter is attached to, obtained from
     * options passed to the Polly constructor.
     */
    private page: Puppeteer.Page

    /**
     * The request resource types this adapter should intercept.
     */
    private requestResourceTypes: Puppeteer.ResourceType[]

    /**
     * A map of all intercepted requests to their respond function, which will be called by the
     * 'response' event listener, causing Polly to record the response content.
     */
    private pendingRequests = new Map<Protocol.Network.Request, { respond: (response: PollyResponse) => void }>()

    /**
     * A map of all intercepted requests to their control callbacks, which will be called in the respondToRequest() method.
     */
    private controlCallbacks = new Map<Protocol.Network.Request, Interceptor.ControlCallbacks>()

    /**
     * A map of all intercepted requests to their passthrough callbacks function, which will be called by the
     * onResponseReceived event listener, causing Polly to record the response content.
     */
    private passthroughCallbacks = new Map<Protocol.Network.Request, (response: PollyResponse) => void>()

    private controlPromises = new Map<Protocol.Network.Request, () => void>()

    /**
     * The adapter's ID, used to reference it in the Polly constructor.
     */
    public static get id(): string {
        return 'puppeteer'
    }

    constructor(polly: Polly) {
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        super(polly)
        this.page = this.options.page
        this.requestResourceTypes = this.options.requestResourceTypes
    }

    /**
     * Called when connecting to a Puppeteer page. Sets up request and response interceptors.
     */
    public onConnect(): void {
        console.log('onConnect')
        // Fulfill requests without sending them to the server
        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        intercept(
            this.page as any,
            this.requestResourceTypes.map(type => patterns[puppeteerToCDPPatterns[type]]('*')).flat(),
            {
                onInterception: ({ request }, controls) => {
                    console.log('onInterception', request.url)
                    this.controlCallbacks.set(request, controls)
                    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
                    // @ts-ignore
                    this.handleRequest(request)
                    const controlPromise = new Promise<void>(resolve => {
                        this.controlPromises.set(request, resolve)
                    })
                    // Resolves when respondToRequest is called
                    // or when passthroughRequest is called
                    return controlPromise
                },
                onResponseReceived: ({ request, response }) => {
                    console.log('onResponseReceived', request.url)
                    const pollyResponse = {
                        statusCode: response.statusCode,
                        headers: Object.fromEntries(
                            (response.headers || []).map(({ name, value }) => [name, value] as const)
                        ),
                        body: response.body,
                    }
                    this.pendingRequests.get(request)?.respond(pollyResponse)
                    this.passthroughCallbacks.get(request)?.(pollyResponse)
                },
            }
        )
    }

    /**
     * Called when disconnecting from a Puppeteer.page.
     */
    public onDisconnect(): void {
        // noop
    }

    /**
     * Given a request that should be allowed to pass through (not be intercepted),
     * return a Promise of the Response for that request, which will be passed to
     * request.respond().
     */
    public async passthroughRequest(pollyRequest: PollyRequest): Promise<PollyResponse> {
        console.log('passthrough request', pollyRequest.url)
        const {
            requestArguments: { request },
        } = (pollyRequest as unknown) as PollyRequestArguments
        this.controlPromises.get(request)?.()
        return new Promise<PollyResponse>(resolve => {
            this.passthroughCallbacks.set(request, resolve)
        })
    }

    /**
     * Responds to an intercepted request with the given response.
     *
     * If an error happened when retreiving the response, abort the request.
     */
    public respondToRequest(
        {
            requestArguments: { request },
            response: { statusCode: status, headers, body },
        }: { requestArguments: { request: Protocol.Network.Request }; response: PollyResponse },
        error?: unknown
    ): void {
        console.log('passthrough request', pollyRequest.url)
        if (error) {
            // TODO figure out if we can pass a more precise reason
            this.controlCallbacks.get(request)?.abort('Failed')
        } else {
            this.controlCallbacks.get(request)?.fulfill(status, {
                responseHeaders: Object.entries(headers).map(([name, value]) => ({ name, value })),
                body,
            })
        }
        this.controlPromises.get(request)?.()
    }

    /**
     * Called when a request is intercepted, for all requests (passthrough or stubbed).
     *
     * Adds an entry to pendingRequests, that will call the provided promise.resolve function
     * when a response for this request is received.
     */
    public onRequest({
        requestArguments: { request },
        promise,
    }: {
        requestArguments: { request: Protocol.Network.Request }
        promise: {
            resolve: (response: PollyResponse) => void
        }
    }): void {
        const respond = (response: PollyResponse): void => {
            promise.resolve(response)
        }
        this.pendingRequests.set(request, {
            respond,
        })
    }
}
