import { Polly, Request as PollyRequest } from '@pollyjs/core'
import Puppeteer from 'puppeteer'
import PollyAdapter from '@pollyjs/adapter'
import { Subscription, fromEvent } from 'rxjs'
import puppeteer from '../../../../shared/src/types/puppeteer-firefox'

interface Response {
    statusCode: number
    headers: Record<string, string>
    body: string
}

interface PollyRequestWithArguments {
    requestArguments: { request: Puppeteer.Request }
    response: Response
}

export class PuppeteerAdapter extends PollyAdapter {
    private subscriptions = new Subscription()
    private page: Puppeteer.Page
    private requestResourceTypes: Puppeteer.ResourceType[]
    private pendingRequests = new Map<
        puppeteer.Request,
        { responsePromise: Promise<Puppeteer.Response>; respond: (response: Puppeteer.Response) => void }
    >()
    private passThroughRequests = new Set<Puppeteer.Request>()
    public static get id(): string {
        return 'puppeteer'
    }
    // eslint-disable-next-line @typescript-eslint/ban-ts-ignore
    // @ts-ignore
    // public static name = 'puppeteer'
    constructor(polly: Polly) {
        // eslint-disable-next-line @typescript-eslint/ban-ts-ignore
        // @ts-ignore
        super(polly)
        this.page = this.options.page
        this.requestResourceTypes = this.options.requestResourceTypes
    }
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
                    // eslint-disable-next-line @typescript-eslint/ban-ts-ignore
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
            // eslint-disable-next-line @typescript-eslint/no-misused-promises
            fromEvent<Puppeteer.Response>(this.page, 'response').subscribe(response => {
                const request = response.request()
                const pendingRequest = this.pendingRequests.get(request)
                if (pendingRequest) {
                    pendingRequest.respond(response)
                }
            })
        )
    }
    public onDisconnect(): void {
        this.subscriptions.unsubscribe()
        this.subscriptions = new Subscription()
    }
    public async passthroughRequest(pollyRequest: PollyRequest): Promise<Response> {
        console.log('PASSTHROUGH', pollyRequest.absoluteUrl)
        const {
            requestArguments: { request },
        } = (pollyRequest as unknown) as PollyRequestWithArguments
        this.passThroughRequests.add(request)
        await request.continue()
        const pendingRequest = this.pendingRequests.get(request)
        if (!pendingRequest) {
            throw new Error('Pouet')
        }
        const response = await pendingRequest.responsePromise
        return {
            statusCode: response.status(),
            headers: response.headers(),
            body: await response.text().catch(() => ''),
        }
    }
    public async respondToRequest(
        {
            requestArguments: { request },
            response: { statusCode: status, headers, body },
        }: { requestArguments: { request: Puppeteer.Request }; response: Response },
        error: unknown
    ): Promise<void> {
        if (this.passThroughRequests.has(request)) {
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
    public onRequest({
        requestArguments: { request },
        promise,
    }: {
        requestArguments: { request: Puppeteer.Request }
        promise: {
            resolve: (response: Response) => void
        }
    }): void {
        let resolveResponsePromise!: (response: Puppeteer.Response) => void
        const respond = async (response: Puppeteer.Response): Promise<void> => {
            resolveResponsePromise(response)
            promise.resolve({
                statusCode: response.status(),
                headers: response.headers(),
                body: await response.text(),
            })
        }
        const responsePromise = new Promise<Puppeteer.Response>(resolve => (resolveResponsePromise = resolve))
        this.pendingRequests.set(request, {
            responsePromise,
            respond,
        })
    }
}
