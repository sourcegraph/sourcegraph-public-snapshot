import type { BillingInterval } from './teamSubscriptions'

export interface CreatePaymentSessionRequest {
    interval: BillingInterval
    seats: number

    customerEmail?: string
    showPromoCodeField: boolean

    returnUrl?: string
}

export interface CreatePaymentSessionResponse {
    clientSecret: string
}

export interface GetPaymentSessionResponse {
    // The only valid state is "complete". Any other string implies that the
    // checkout was not successful, and no new Cody Pro team was registered.
    status: string

    // The only valid state is "paid" (IFF status is also "complete").
    // Anything else means the team/subscription was not registered.
    paymentStatus: string
}
