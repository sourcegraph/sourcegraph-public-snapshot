import React from 'react'

import { Alert, Badge, type BadgeVariantType, Tooltip } from '@sourcegraph/wildcard'

import {
    ALL_PLANS,
    TAG_AIR_GAPPED,
    TAG_BATCH_CHANGES,
    TAG_CODE_INSIGHTS,
    TAG_TRIAL,
    TAG_TRUEUP,
} from '../site-admin/dotcom/productSubscriptions/plandata'

const getBadgeVariant = (tag: string): BadgeVariantType => {
    if (isUnknownTag(tag)) {
        return 'danger'
    }
    if (tag.startsWith('plan:')) {
        return 'primary'
    }
    if (tag.startsWith('customer:')) {
        return 'success'
    }
    return 'secondary'
}

const getTagDescription = (tag: string): string => {
    if (isUnknownTag(tag)) {
        return 'Uknown tag. Please check if it is correct.'
    }
    if (tag.startsWith('plan:')) {
        return 'Subscription plan'
    }
    if (tag.startsWith('customer:')) {
        return 'Customer name'
    }
    return 'Plan features'
}

const knownLicenseTags = [
    ...ALL_PLANS.map(plan => `plan:${plan.label}`),
    TAG_AIR_GAPPED.tagValue,
    TAG_BATCH_CHANGES.tagValue,
    TAG_CODE_INSIGHTS.tagValue,
    TAG_TRIAL.tagValue,
    TAG_TRUEUP.tagValue,
]

// TODO: Maybe a field for a custom comment on what instance the key is for?
// In accordance with:
// Add trial and instance:test or instance:whatever_name_is_appropriate tags, so that we can identify which license keys are test and which are not
// from https://handbook.sourcegraph.com/departments/technical-success/ce/process/license_keys.

// TODO: Add MAU switch.

// TODO: Should team-0 still exist? If so, we should simplify
// Teams Licenses: Only applicable for team license renewals. Add plan:team-0,acls,monitoring, plus the customer name, to all Teams licenses.
// from https://handbook.sourcegraph.com/departments/technical-success/ce/process/license_keys
// so that it's no longer required to add the labels manually.

// TODO:
// If no plan:* tag is supplied, the license will be treated as legacy enterprise tier which has unlimited access to all features.
// When can we remove this behavior? Let's check if there are active licenses still.

export const isUnknownTag = (tag: string): boolean => !knownLicenseTags.includes(tag) && !tag.startsWith('customer:')

export const hasUnknownTags = (tags: string[]): boolean => tags.some(isUnknownTag)

export const UnknownTagWarning: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <Alert className={className} variant="danger">
        License tags contain unknown values (marked red), please check if the tags are correct.
    </Alert>
)

export const ProductLicenseTags: React.FunctionComponent<
    React.PropsWithChildren<{
        tags: string[]
        small?: boolean
    }>
> = ({ tags, small }) => (
    <>
        {tags.map(tag => (
            <Tooltip key={tag} content={getTagDescription(tag)}>
                <Badge variant={getBadgeVariant(tag)} className="mr-1" small={small}>
                    {tag}
                </Badge>
            </Tooltip>
        ))}
    </>
)
