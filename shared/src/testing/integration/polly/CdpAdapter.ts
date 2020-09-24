import { Polly, Request as PollyRequest } from '@pollyjs/core'
import Puppeteer from 'puppeteer'
import { patterns } from 'puppeteer-interceptor'
import Protocol from 'devtools-protocol'
import PollyAdapter from '@pollyjs/adapter'

function toBase64(input: string): string {
    return Buffer.from(input).toString('base64')
}

function fromBase64(input: string): string {
    return Buffer.from(input, 'base64').toString()
}

interface PollyResponse {
    statusCode: number
    headers: Record<string, string>
    body: string
}

interface PollyRequestArguments {
    requestArguments: { requestId: string }
}

interface PollyPromise extends Promise<PollyResponse> {
    resolve(response: PollyResponse): void
    reject(error: any): void
}

// @ts-ignore
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
 * A Puppeteer adapter for Polly that uses "Fetch domain" of the Chrome Devtools
 * protocol to intercept and fulfill requests.
 *
 */
export class CdpAdapter extends PollyAdapter {
    /**
     * The puppeteer Page this adapter is attached to, obtained from
     * options passed to the Polly constructor.
     */
    private page: Puppeteer.Page

    /**
     * The request resource types this adapter should intercept.
     */
    // @ts-ignore
    private requestResourceTypes: Puppeteer.ResourceType[]

    /**
     * A map of all intercepted requests to their respond function, which will be called by the
     * 'response' event listener, causing Polly to record the response content.
     */
    private pendingRequests = new Map<string, PollyPromise>()

    /**
     * TODO: write doc comment
     */
    private cdpSession?: Puppeteer.CDPSession

    /**
     * A map of all intercepted requests to their passthrough callbacks function, which will be called by the
     * onResponseReceived event listener, causing Polly to record the response content.
     */
    private passthroughCallbacks = new Map<string, (response: PollyResponse) => void>()

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
     * Called when connecting to a Puppeteer page. Sets up CDP request
     * interception using the CDP "Fetch domain".
     */
    public async onConnect(): Promise<void> {
        console.log('CDP adapter connecting')
        this.cdpSession = await this.page.target().createCDPSession()

        // TODO: This is where we narrow down the interception with patterns.
        // Request and respond stages are independant, so we can set a different
        // set of patterns for each.
        const fetchEnableRequest: Protocol.Fetch.EnableRequest = {
            patterns: [{ requestStage: 'Request' }, { requestStage: 'Response' }],
        }
        await this.cdpSession.send('Fetch.enable', fetchEnableRequest)

        this.cdpSession.on(
            'Fetch.requestPaused',
            async (event: Protocol.Fetch.RequestPausedEvent): Promise<void> => {
                const isInResponseStage = eventIsInResponseStage(event)
                if (isInResponseStage) {
                    await this.handlePausedRequestInResponseStage(event)
                } else {
                    await this.handlePausedRequestInRequestStage(event)
                }
            }
        )
    }

    /**
     * Called when disconnecting from a Puppeteer.page.
     */
    public async onDisconnect(): Promise<void> {
        console.log('CDP adapter disconnecting')
        await this.trySendCdpRequest('Fetch.disable')
    }

    /**
     * Given a request that should be allowed to pass through (not be intercepted),
     * return a Promise of the Response for that request, which will be passed to
     * request.respond().
     */
    public passthroughRequest(pollyRequest: PollyRequest & PollyRequestArguments): Promise<PollyResponse> {
        const {
            requestArguments: { requestId },
        } = pollyRequest

        return new Promise<PollyResponse>((resolve, reject) => {
            this.passthroughCallbacks.set(requestId, resolve)
            this.continuePausedRequest({ requestId })
        })
    }

    /**
     * Responds to an intercepted request with the given response.
     *
     * If an error happened when retrieving the response, abort the request.
     */
    public async respondToRequest(
        pollyRequest: PollyRequest & { response: PollyResponse } & PollyRequestArguments,
        error?: unknown
    ): Promise<void> {
        const { response: pollyResponse, requestArguments } = pollyRequest
        const { headers, body = '' } = pollyResponse
        const { requestId } = requestArguments
        if (error) {
            // This function receives a value in the `error` argument if we're
            // intercepting a request with the Polly server route which throws
            // an error.
            console.warn('Aborting request:', error)
            await this.abortPausedRequest({ requestId, errorReason: 'Failed' })
        } else {
            // Fulfill by converting the Polly response to a CDP response
            const cdpRequestToFulfill = {
                requestId,
                responseCode: 200, // statusCode,
                responseHeaders: headerObjectToHeaderEntries(headers),
                body: toBase64(body),
            }
            console.log(this.pendingRequests)
            console.log('Fulfilling', cdpRequestToFulfill)
            await this.fulfillPausedRequest(cdpRequestToFulfill)
        }
    }

    /**
     * Called when a request is intercepted, for all requests (passthrough or stubbed).
     *
     * Adds an entry to pendingRequests, that will call the provided promise.resolve function
     * when a response for this request is received.
     */
    public onRequest(pollyRequest: PollyRequest & PollyRequestArguments & { promise: PollyPromise }): void {
        console.log('onRequest', pollyRequest.requestArguments.requestId)
        const { requestArguments, promise } = pollyRequest

        const { requestId } = requestArguments
        this.pendingRequests.set(requestId, promise)
    }

    private async fulfillPausedRequest(request: Protocol.Fetch.FulfillRequestRequest): Promise<void> {
        await this.trySendCdpRequest('Fetch.fulfillRequest', request)
    }

    private async continuePausedRequest(request: Protocol.Fetch.ContinueRequestRequest): Promise<void> {
        await this.trySendCdpRequest('Fetch.continueRequesst', request)
    }

    /**
     * Perform a CDP request call that doesn't return a result, while ignoring
     * errors due to the page being closed already.
     */
    private async trySendCdpRequest(cdpRequestName: string, request?: object) {
        try {
            await this.cdpSession?.send(cdpRequestName, request)
        } catch (error) {
            // TODO: also ignore "target closed" error
            if (error.message.endsWith('Session closed. Most likely the page has been closed.')) {
                return
            }
            throw error
        }
    }

    private async abortPausedRequest(request: Protocol.Fetch.FailRequestRequest): Promise<void> {
        await this.cdpSession?.send('Fetch.failRequest', request)
    }

    private async getResponseBody(event: Protocol.Fetch.RequestPausedEvent): Promise<string> {
        if (getLocationHeader(event)) {
            return '' // CDP doesn't let us obtain the body of redirect requests, so we don't attempt it.
        }
        if (!this.cdpSession) {
            throw new Error('Fetch.getResponseBody called before CDP session created')
        }
        const body = (await this.cdpSession.send('Fetch.getResponseBody', {
            requestId: event.requestId,
        })) as Protocol.Fetch.GetResponseBodyResponse

        return getBodyStringFromCdpBody(body)
    }

    private async handlePausedRequestInRequestStage(event: Protocol.Fetch.RequestPausedEvent): Promise<void> {
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        this.handleRequest({
            ...event.request,
            url: event.request.url,
            method: event.request.method,
            headers: event.request.headers,
            body: event.request.postData ?? '',
            requestArguments: { requestId: event.requestId },
        })
    }

    private async handlePausedRequestInResponseStage(event: Protocol.Fetch.RequestPausedEvent): Promise<void> {
        const { requestId } = event

        if (event.responseErrorReason) {
            // Response is an error
            this.pendingRequests.get(requestId)?.reject(new Error(event.responseErrorReason))
            this.pendingRequests.delete(requestId)
        }

        const body = await this.getResponseBody(event)

        // Convert the CDP response into a Polly response

        const pollyResponse: PollyResponse = {
            statusCode: event.responseStatusCode ?? 0, // TODO: what if the response is a failure
            headers: headerEntriesToHeaderObject(event.responseHeaders),
            body,
        }

        // The requestId may or may not have an associated passthrough callback.
        // If it does, call it and delete it. TODO: verify if we need the
        // ability to reject the passthrough promise as well, in which case we
        // need both resolve and reject callbacks to be available here.
        this.passthroughCallbacks.get(requestId)?.(pollyResponse)
        this.passthroughCallbacks.delete(requestId)

        // Each pending request has an associated promise. Because at this point
        // the request is done (given that a response has been received), we can
        // resolve the pending request promise.
        this.pendingRequests.get(requestId)?.resolve(pollyResponse)
        this.pendingRequests.delete(requestId)
    }
}

/**
 * Determine if the request is paused in the response stage. If false, then the
 * request is paused in the request stage.
 */
function eventIsInResponseStage(event: Protocol.Fetch.RequestPausedEvent): boolean {
    return event.responseStatusCode !== undefined || event.responseErrorReason !== undefined
}

/**
 * Get the value of the "Location" response header.
 **/
function getLocationHeader(event: Protocol.Fetch.RequestPausedEvent): string | undefined {
    const { responseHeaders = [] } = event
    const foundLocationHeader = responseHeaders.find(header => header.name === 'location')
    if (foundLocationHeader) {
        return foundLocationHeader.value
    }
    return undefined
}

/**
 * Transform a header represented by an array of
 * {@link Protocol.Fetch.HeaderEntry} (the format used by CDP) into a headers
 * object (the format used by Polly)
 */
function headerEntriesToHeaderObject(responseHeaders: Protocol.Fetch.HeaderEntry[] = []): Record<string, string> {
    return Object.fromEntries(responseHeaders.map(({ name, value }) => [name, value]))
}

/** Transform a header object (the format used by Polly) into an array of header
 * entries (the format used by CDP) */
function headerObjectToHeaderEntries(headers: Record<string, string>): Protocol.Fetch.HeaderEntry[] {
    return Object.entries(headers).map(([name, value]) => ({ name, value }))
}

/**
 * Get the body data as a string, from the response of a `Fetch.getResponseBody` call.
 */
function getBodyStringFromCdpBody(body: Protocol.Fetch.GetResponseBodyResponse): string {
    if (!body) {
        return ''
    }
    if (body.base64Encoded) {
        return fromBase64(body.body)
    }
    return body.body
}
