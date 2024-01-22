package ssc

import (
	"time"
)

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
	SubscriptionStatusPending  SubscriptionStatus = "pending"
	SubscriptionStatusOther    SubscriptionStatus = "other"
)

type SSCSubscription struct {
	SAMSAccountID string
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

type UserCodySubscription struct {
	Status               SubscriptionStatus
	IsPro                bool
	ApplyProRateLimits   bool
	CurrentPeriodStartAt time.Time
	CurrentPeriodEndAt   time.Time
}
