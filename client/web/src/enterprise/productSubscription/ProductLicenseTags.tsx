import React from 'react'

import { Alert, Badge, BadgeVariantType, Tooltip } from '@sourcegraph/wildcard'

// TODO: instead of hardcoding, generate this list based on /enterprise/internal/licensing/data.go
const PLAN_TAGS = new Set<string>([
    'starter',
    'plan:old-starter-0',
    'plan:old-enterprise-0',
    'plan:team-0',
    'plan:enterprise-0',
    'plan:business-0',
    'plan:enterprise-1',
    'plan:enterprise-extension',
    'plan:free-0',
])

const FEATURE_TAGS = new Set<string>([
    'sso',
    'acls',
    'explicit-permissions-api',
    'private-extension-registry',
    'remote-extensions-allow-disallow',
    'branding',
    'campaigns',
    'monitoring',
    'backup-and-restore',
    'code-insights',
    'batch-changes',
])

const getBadgeVariant = (tag: string): BadgeVariantType => {
    if (PLAN_TAGS.has(tag)) {
        return 'primary'
    }
    if (FEATURE_TAGS.has(tag)) {
        return 'secondary'
    }
    if (tag.startsWith('customer')) {
        return 'success'
    }

    return 'danger'
}

const getTagDescription = (tag: string): string => {
    if (PLAN_TAGS.has(tag)) {
        return 'Subscription plan'
    }
    if (FEATURE_TAGS.has(tag)) {
        return 'Plan features'
    }
    if (tag.startsWith('customer')) {
        return 'Customer tag'
    }

    return 'Uknown tag. Please check if it is correct.'
}

export const isUnknownTag = (tag: string): boolean =>
    !PLAN_TAGS.has(tag) && !FEATURE_TAGS.has(tag) && !tag.startsWith('customer:')

export const hasUnknownTags = (tags: string[]): boolean => tags.some(isUnknownTag)

export const UnknownTagWarning: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Alert className="mt-3 mb-3" variant="warning">
        License tags contain unknown values (marked red), please check if the tags are correct.
    </Alert>
)

export const ProductLicenseTags: React.FunctionComponent<
    React.PropsWithChildren<{
        tags: string[]
    }>
> = ({ tags }) => (
    <>
        {tags.map(tag => (
            <Tooltip key={tag} content={getTagDescription(tag)}>
                <Badge variant={getBadgeVariant(tag)} className="mr-1" as="div">
                    {tag}
                </Badge>
            </Tooltip>
        ))}
    </>
)
