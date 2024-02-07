import React from 'react'

import classNames from 'classnames'

import { Alert, Badge, type BadgeVariantType, Tooltip } from '@sourcegraph/wildcard'

import {
    ALL_PLANS,
    DEPRECATED_TAGS,
    TAG_AIR_GAPPED,
    TAG_BATCH_CHANGES,
    TAG_CODE_INSIGHTS,
    TAG_DEV,
    TAG_DISABLE_TELEMETRY_EXPORT,
    TAG_INTERNAL,
    TAG_TRIAL,
    TAG_TRUEUP,
} from '../site-admin/dotcom/productSubscriptions/plandata'

import styles from './ProductLicenseTags.module.scss'

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
    if (DEPRECATED_TAGS.some(deprecatedTag => deprecatedTag.tagValue === tag)) {
        return 'Deprecated plan feature, this tag has no effect anymore'
    }

    return 'Plan features'
}

const knownLicenseTags = new Set([
    ...ALL_PLANS.map(plan => `plan:${plan.label}`),
    TAG_AIR_GAPPED.tagValue,
    TAG_BATCH_CHANGES.tagValue,
    TAG_CODE_INSIGHTS.tagValue,
    TAG_TRIAL.tagValue,
    TAG_TRUEUP.tagValue,
    TAG_DISABLE_TELEMETRY_EXPORT.tagValue,
    TAG_DEV.tagValue,
    TAG_INTERNAL.tagValue,
    ...DEPRECATED_TAGS.map(tag => tag.tagValue),
])

export const isUnknownTag = (tag: string): boolean => !knownLicenseTags.has(tag) && !tag.startsWith('customer:')

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
    <div className={styles.tagsWrapper}>
        {tags.map(tag => (
            <Tooltip key={tag} content={getTagDescription(tag)}>
                <Badge
                    variant={getBadgeVariant(tag)}
                    small={small}
                    className={classNames(
                        DEPRECATED_TAGS.some(deprecatedTag => deprecatedTag.tagValue === tag) && styles.deprecatedTag
                    )}
                >
                    {tag}
                </Badge>
            </Tooltip>
        ))}
    </div>
)
