import { Polly, Request as PollyRequest } from '@pollyjs/core'
import type * as Puppeteer from 'puppeteer'
import PollyAdapter from '@pollyjs/adapter'
import { Subscription, fromEvent } from 'rxjs'

interface Response {
    statusCode: number
    headers: Record<string, string>
    body: string
}

interface PollyRequestArguments {
    requestArguments: { request: Puppeteer.Request }
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
    private subscriptions = new Subscription()

    /**
     * The puppeteer Page this adapter is attached to, obtained from
     * options passed to the Polly constructor.
     */
    private page: Puppeteer.Page

    /**
     * The request resource types tis adapter should intercept.
     */
    private requestResourceTypes: Puppeteer.ResourceType[]

    /**
     * A map of all intercepted requests to their respond function, which will be called by the
     * 'response' event listener, causing Polly to record the response content.
     */
    private pendingRequests = new Map<Puppeteer.Request, { respond: (response: Puppeteer.Response) => void }>()

    /**
     * Maps passthrough requests to an object containing:
     * - The response promise, which will be awaited in this.onPassthrough
     * - A respond function, called by 'response' event listener, which resolves the response promise.
     */
    private passThroughRequests = new Map<
        Puppeteer.Request,
        { responsePromise: Promise<Puppeteer.Response>; respond: (response: Puppeteer.Response) => void }
    >()

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
        this.subscriptions.add(
            fromEvent<Puppeteer.Request>(this.page, 'request').subscribe(request => {
                const url = request.url()
                const method = request.method()
                const headers = request.headers()
                const isPreflight =
                    method === 'OPTIONS' && !!headers.origin && !!headers['access-control-request-method']
                if (isPreflight || !this.requestResourceTypes.includes(request.resourceType())) {
                    // eslint-disable-next-line @typescript-eslint/no-floating-promises
                    request.continue()
                } else {
                    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
                    // @ts-ignore
                    this.handleRequest({
                        headers,
                        url,
                        method,
                        body: request.postData() ?? '',
                        requestArguments: {
                            request,
                        },
                    })
                }
            })
        )
        this.subscriptions.add(
            fromEvent<Puppeteer.Response>(this.page, 'response').subscribe(response => {
                const request = response.request()
                if (this.pendingRequests.has(request)) {
                    // eslint-disable-next-line no-unused-expressions
                    this.pendingRequests.get(request)?.respond(response)
                    this.pendingRequests.delete(request)
                }
                if (this.passThroughRequests.has(request)) {
                    // eslint-disable-next-line no-unused-expressions
                    this.passThroughRequests.get(request)?.respond(response)
                }
            })
        )
    }

    /**
     * Called when disconnecting from a Puppeteer.page.
     */
    public onDisconnect(): void {
        this.subscriptions.unsubscribe()
    }

    /**
     * Given a request that should be allowed to pass through (not be intercepted),
     * return a Promise of the Response for that request, which will be passed to
     * request.respond().
     */
    public async passthroughRequest(pollyRequest: PollyRequest): Promise<Response> {
        const {
            requestArguments: { request },
        } = (pollyRequest as unknown) as PollyRequestArguments
        let respond: (response: Puppeteer.Response) => void
        const responsePromise = new Promise<Puppeteer.Response>(resolve => (respond = resolve))
        this.passThroughRequests.set(request, { respond: response => respond(response), responsePromise })
        await request.continue()
        const response = await responsePromise
        return {
            statusCode: response.status(),
            headers: response.headers(),
            body: await response.text().catch(() => ''),
        }
    }

    /**
     * Responds to an intercepted request with the given response.
     *
     * If an error happened when retreiving the response, abort the request.
     */
    public async respondToRequest(
        {
            requestArguments: { request },
            response: { statusCode: status, headers, body },
        }: { requestArguments: { request: Puppeteer.Request }; response: Response },
        error?: unknown
    ): Promise<void> {
        // Do nothing for passthrough requests: Polly calls request.respond() internally.
        if (this.passThroughRequests.has(request)) {
            this.passThroughRequests.delete(request)
            return
        }
        if (error) {
            await request.abort()
        } else {
            await request.respond({
                status,
                headers,
                body,
            })
        }
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
        requestArguments: { request: Puppeteer.Request }
        promise: {
            resolve: (response: Response) => void
        }
    }): void {
        if (this.passThroughRequests.has(request)) {
            return
        }
        const respond = async (response: Puppeteer.Response): Promise<void> => {
            promise.resolve({
                statusCode: response.status(),
                headers: response.headers(),
                body: await response.text(),
            })
        }
        this.pendingRequests.set(request, {
            respond,
        })
    }
}
