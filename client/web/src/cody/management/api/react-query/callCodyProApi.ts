import type { Call } from '../client'

export class CodyProApiError extends Error {
    constructor(message: string, public status: number) {
        super(message)
    }
}

/**
 * Builds the RequestInit object for the fetch API with the necessary headers and options
 * to authenticate the request with the Sourcegraph backend.
 */
const buildRequestInit = ({ headers = {}, ...init }: RequestInit): RequestInit => ({
    // Pass along the "sgs" session cookie to identify the caller.
    credentials: 'same-origin',
    headers: {
        // In order for the Sourcegraph backend to authenticate the request, we need to
        // ensure we don't run afoul of our CSRF protections (see csrf_security_model.md).
        //
        // Setting the `x-requested-with` header, along with other the current CORS config
        // is sufficient for backend request to be authenticated. (See `CookieMiddlewareWithCSRFSafety()`.)
        //
        // On a related note, the `fetch` API does NOT include the "origin" header for GET
        // or HEAD requests by spec. (See https://fetch.spec.whatwg.org/#origin-header.)
        'x-requested-with': 'Sourcegraph/CodyProApiClient',
        ...headers,
    },
    ...init,
})

const signOutAndRedirectToSignIn = async (): Promise<void> => {
    const response = await fetch('/-/sign-out', buildRequestInit({ method: 'GET' }))
    if (response.ok) {
        window.location.href = `/sign-in?returnTo=${
            window.location.pathname + encodeURIComponent(window.location.search)
        }`
    }
}

// Important: This function has the side effect of logging the user out and redirecting them
// to the sign-in page with the current page as the return URL if they are not authenticated.
export const callCodyProApi = async (call: Call<unknown>): Promise<Response> => {
    const response = await fetch(
        `/.api/ssc/proxy${call.urlSuffix}`,
        buildRequestInit({
            method: call.method,
            body: call.requestBody ? JSON.stringify(call.requestBody) : undefined,
        })
    )

    if (!response.ok) {
        if (response.status === 401) {
            await signOutAndRedirectToSignIn()
            // User is redirected to another page, so no need to throw an error.
            return response
        }

        // Throw errors for unsuccessful HTTP calls so that `callCodyProApi` callers don't need to check whether the response is OK.
        // Motivation taken from here: https://tanstack.com/query/latest/docs/framework/react/guides/query-functions#usage-with-fetch-and-other-clients-that-do-not-throw-by-default
        throw new CodyProApiError(await response.text(), response.status)
    }

    return response
}
