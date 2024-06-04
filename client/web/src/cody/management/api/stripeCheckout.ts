import type { BillingInterval } from './teamSubscriptions'

export interface CreateCheckoutSessionRequest {
    interval: BillingInterval
    seats: number
    canChangeSeatCount?: boolean

    customerEmail?: string
    showPromoCodeField: boolean

    returnUrl?: string

    stripeUiMode?: 'embedded' | 'custom'
}

export interface CreateCheckoutSessionResponse {
    clientSecret: string
}

export interface GetCheckoutSessionResponse {
    // The only valid state is "complete". Any other string implies that the
    // checkout was not successful, and no new Cody Pro team was registered.
    status: string

    // The only valid state is "paid" (IFF status is also "complete").
    // Anything else means the team/subscription was not registered.
    paymentStatus: string
}
