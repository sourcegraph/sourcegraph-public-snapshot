import React from 'react'

import { Alert, Badge, type BadgeVariantType, Tooltip } from '@sourcegraph/wildcard'

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
