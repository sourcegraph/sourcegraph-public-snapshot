package ssc

type BillingInterval string

const (
	BillingIntervalDaily   BillingInterval = "daily"
	BillingIntervalMonthly BillingInterval = "monthly"
	BillingIntervalYearly  BillingInterval = "yearly"
)

type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusPastDue  SubscriptionStatus = "past_due"
	SubscriptionStatusUnpaid   SubscriptionStatus = "unpaid"
	SubscriptionStatusCanceled SubscriptionStatus = "canceled"
	SubscriptionStatusTrialing SubscriptionStatus = "trialing"
	SubscriptionStatusOther    SubscriptionStatus = "other"
	// NOTE: The "pending" status is only temporary, and will be removed after Feb 15 2024.
	// This is to support the pre-release state where a user has opted for a free Pro trial
	// but has not put in their cc in SSC yet.
	SubscriptionStatusPending SubscriptionStatus = "pending"
)

type Subscription struct {
	// Status is the current status of the subscription, e.g. "active" or "canceled".
	Status          SubscriptionStatus `json:"status"`
	BillingInterval BillingInterval    `json:"billingInterval"`

	// CancelAtPeriodEnd flags whether or not a subscription will automatically cancel at the end
	// of the current billing cycle, or if it will renew.
	CancelAtPeriodEnd bool `json:"cancelAtPeriodEnd"`

	// Billing cycle anchors are times represented as an ISO-8601 string.
	// e.g. "2024-01-17T13:18:05−07:00"
	CurrentPeriodStart string `json:"currentPeriodStart"`
	CurrentPeriodEnd   string `json:"currentPeriodEnd"`
}
