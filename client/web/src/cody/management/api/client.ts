import type * as types from './types'

// Client provides the metadata for the methods exposed from the Cody Pro API client.
//
// This doesn't _do_ anything, it just returns the metadata for what needs to be done.
// It is used in conjunction with a Caller implementation for actually fetching data.
// eslint-disable-next-line @typescript-eslint/prefer-namespace-keyword, @typescript-eslint/no-namespace
export module Client {
    // Subscriptions

    export function getCurrentSubscription(): Call<types.Subscription> {
        return { method: 'GET', urlSuffix: '/team/current/subscription' }
    }

    export function getCurrentSubscriptionSummary(): Call<types.SubscriptionSummary> {
        return { method: 'GET', urlSuffix: '/team/current/subscription/summary' }
    }

    export function updateCurrentSubscription(requestBody: types.UpdateSubscriptionRequest): Call<types.Subscription> {
        return { method: 'PATCH', urlSuffix: '/team/current/subscription', requestBody }
    }

    export function previewUpdateCurrentSubscription(
        requestBody: types.PreviewUpdateSubscriptionRequest
    ): Call<types.PreviewResult> {
        return { method: 'PATCH', urlSuffix: '/team/current/subscription/preview', requestBody }
    }

    export function getCurrentSubscriptionInvoices(): Call<types.GetSubscriptionInvoicesResponse> {
        return { method: 'GET', urlSuffix: '/team/current/subscription/invoices' }
    }

    export function reactivateCurrentSubscription(
        requestBody: types.ReactivateSubscriptionRequest
    ): Call<types.GetSubscriptionInvoicesResponse> {
        return { method: 'POST', urlSuffix: '/team/current/subscription/reactivate', requestBody }
    }

    // Teams

    export function createTeam(requestBody: types.CreateTeamRequest): Call<string> {
        return { method: 'POST', urlSuffix: '/team', requestBody }
    }

    export function previewCreateTeam(requestBody: types.PreviewCreateTeamRequest): Call<types.PreviewResult> {
        return { method: 'POST', urlSuffix: '/team/preview', requestBody }
    }

    // Team members

    export function getCurrentTeamMembers(): Call<types.ListTeamMembersResponse> {
        return { method: 'GET', urlSuffix: '/team/current/members' }
    }

    export function updateTeamMember(requestBody: types.UpdateTeamMembersRequest): Call<unknown> {
        return { method: 'PATCH', urlSuffix: '/team/current/members', requestBody }
    }

    // Invites

    export function getInvite(teamId: string, inviteId: string): Call<types.TeamInvite> {
        return { method: 'GET', urlSuffix: `/team/${teamId}/invites/${inviteId}` }
    }

    export function getTeamInvites(): Call<types.ListTeamInvitesResponse> {
        return { method: 'GET', urlSuffix: '/team/current/invites' }
    }

    export function sendInvite(requestBody: types.CreateTeamInviteRequest): Call<types.ListTeamInvitesResponse> {
        return { method: 'POST', urlSuffix: '/team/current/invites', requestBody }
    }

    export function resendInvite(inviteId: string): Call<unknown> {
        return { method: 'POST', urlSuffix: `/team/current/invites/${inviteId}/resend` }
    }

    export function acceptInvite(teamId: string, inviteId: string): Call<unknown> {
        return { method: 'POST', urlSuffix: `/team/${teamId}/invites/${inviteId}/accept` }
    }

    export function cancelInvite(teamId: string, inviteId: string): Call<unknown> {
        return { method: 'POST', urlSuffix: `/team/${teamId}/invites/${inviteId}/cancel` }
    }

    // Stripe Checkout

    export function createStripeCheckoutSession(
        requestBody: types.CreateCheckoutSessionRequest
    ): Call<types.CreateCheckoutSessionResponse> {
        return { method: 'POST', urlSuffix: '/checkout/session', requestBody }
    }
}

// Call is the bundle of data necessary for making an API request.
// This is a sort of "meta request" in the same vein as the `gql`
// template tag, see: https://github.com/apollographql/graphql-tag
export interface Call<Resp> {
    method: 'GET' | 'POST' | 'PATCH' | 'DELETE'
    urlSuffix: string
    requestBody?: unknown

    // Unused. This will never be set, it is only to
    // pass along the expected response type.
    responseBody?: Resp
}

// Caller is a wrapper around an HTTP client. An implementation of this interface
// will be responsible for making API calls to the backend.
export interface Caller {
    // call performs the described HTTP request, returning the response body deserialized from
    // JSON as `data`, and the full HTTP response object as `response`.
    call<Data>(call: Call<Data>): Promise<{ data?: Data; response: Response }>
}

// CodyProApiCaller is an implementation of the Caller interface which issues API calls to
// the current Sourcegraph instance's SSC proxy API endpoint.
export class CodyProApiCaller implements Caller {
    // e.g. "https://sourcegraph.com"
    private readonly origin: string

    constructor() {
        this.origin = window.location.origin
    }

    public async call<Data>(call: Call<Data>): Promise<{ data?: Data; response: Response }> {
        let bodyJson: string | undefined
        if (call.requestBody) {
            bodyJson = JSON.stringify(call.requestBody)
        }

        const fetchResponse = await fetch(`${this.origin}/.api/ssc/proxy${call.urlSuffix}`, {
            // Pass along the "sgs" session cookie to identify the caller.
            credentials: 'same-origin',
            method: call.method,
            body: bodyJson,
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
            },
        })

        if (fetchResponse.status >= 200 && fetchResponse.status <= 299) {
            const rawBody = await fetchResponse.text()
            const typedResp = JSON.parse(rawBody) as Data
            return {
                data: typedResp,
                response: fetchResponse,
            }
        }

        // Otherwise just return the raw response. We rely on the caller
        // to confirm that the Response object indicates success, and to
        // handle any 4xx or 5xx status codes.
        return {
            data: undefined,
            response: fetchResponse,
        }
    }
}
