import type { TeamRole } from './teamMembers'

// BillingInterval is the subscription's billing cycle. 'daily' is only
// available in the dev environment.
export type BillingInterval = 'daily' | 'monthly' | 'yearly'

// UsdCents is used to wrap any situation involving money, which
// should always be an integer referring to USD cents. e.g.
// 105 corresponds to $1.05.
export type UsdCents = number

export type InvoiceStatus = 'draft' | 'open' | 'paid' | 'other'

export type SubscriptionStatus = 'active' | 'past_due' | 'unpaid' | 'canceled' | 'trailing' | 'other'

export interface Address {
    line1: string
    line2: string
    city: string
    state: string
    postalCode: string
    country: string
}

export interface PaymentMethod {
    expMonth: number
    expYear: number
    last4: string
}

export interface PreviewResult {
    dueNow: UsdCents
    newPrice: UsdCents
    dueDate: string
}

export interface DiscountInfo {
    description: string
    expiresAt?: string
}

export interface Invoice {
    date: string

    amountDue: UsdCents
    amountPaid: UsdCents
    status: InvoiceStatus

    periodStart: string
    periodEnd: string

    hostedInvoiceUrl?: string
    pdfUrl?: string
}

export interface SubscriptionSummary {
    teamId: string

    userRole: TeamRole
    teamCurrentMembers: number
    teamMaxMembers: number

    subscriptionStatus: SubscriptionStatus
    cancelAtPeriodEnd: boolean
}

export interface Subscription {
    createdAt: string
    endedAt?: string

    primaryEmail: string
    name: string
    address: Address

    subscriptionStatus: SubscriptionStatus
    cancelAtPeriodEnd: boolean
    billingInterval: BillingInterval

    discountInfo?: DiscountInfo

    currentPeriodStart: string
    currentPeriodEnd: string

    paymentMethod?: PaymentMethod
    nextInvoice?: PreviewResult

    maxSeats: number
}

export interface CustomerUpdateOptions {
    newName?: string
    newEmail?: string
    newAddress?: Address
    newCreditCardToken?: string
}

// Is a discriminated union. Exactly one field should be set at a time.
export interface SubscriptionUpdateOptions {
    newSeatCount?: number
    newBillingInterval?: BillingInterval
    newCancelAtPeriodEnd?: boolean
}

export interface ReactivateSubscriptionRequest {
    seatLimit: number
    billingInterval: BillingInterval
    creditCardToken?: string
}

export interface UpdateSubscriptionRequest {
    customerUpdate?: CustomerUpdateOptions
    subscriptionUpdate?: SubscriptionUpdateOptions
}

export interface PreviewUpdateSubscriptionRequest {
    newSeatCount?: number
    newBillingInterval?: BillingInterval
    newCancelAtPeriodEnd?: boolean
}

export interface GetSubscriptionInvoicesResponse {
    invoices: Invoice[]
    continuationToken?: string
}

export interface CreateTeamRequest {
    name: string
    slug: string
    seats: number
    address: Address
    billingInterval: BillingInterval
    couponCode?: string
    creditCardToken: string
}

export interface PreviewCreateTeamRequest {
    seats: number
    billingInterval: BillingInterval
    couponCode?: string
}
