import React from 'react'

import { Alert, Badge, BadgeVariantType, Tooltip } from '@sourcegraph/wildcard'

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

export const isUnknownTag = (tag: string): boolean =>
    !window.context.licenseInfo?.knownLicenseTags?.includes(tag) && !tag.startsWith('customer:')

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
