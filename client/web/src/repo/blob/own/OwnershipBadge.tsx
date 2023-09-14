import React from 'react'

import { Badge, type BadgeProps, type BadgeVariantType, Link } from '@sourcegraph/wildcard'

import type {
    AssignedOwnerFields,
    CodeownersFileEntryFields,
    RecentContributorOwnershipSignalFields,
    RecentViewOwnershipSignalFields,
} from '../../../graphql-operations'

import styles from './FileOwnershipEntry.module.scss'

interface Props {
    reason:
        | CodeownersFileEntryFields
        | RecentContributorOwnershipSignalFields
        | RecentViewOwnershipSignalFields
        | AssignedOwnerFields
}

const BADGE_VARIANT_BY_REASON: Partial<Record<NonNullable<Props['reason']['__typename']>, BadgeVariantType>> = {
    // Orange badge for CODEOWNERS entry.
    CodeownersFileEntry: 'warning',
    // Purple badge for assigned owners and teams.
    AssignedOwner: 'merged',
}

// OwnershipBadge displays a badge like CODEOWNERS or RECENT CONTRIBUTOR
// next to the owner or inferred ownership signal.
//
// In case of the codeowners file, the badge links to the line that contains
// relevant rule in the CODEOWNERS file.
export const OwnershipBadge: React.FunctionComponent<Props> = ({ reason }) => {
    // For CODEOWNERS, the Badge links to the codeowners rule
    const linkProps: BadgeProps =
        reason.__typename === 'CodeownersFileEntry'
            ? ({ as: Link, to: `${reason.codeownersFile.url}?L${reason.ruleLineMatch}` } as BadgeProps)
            : {}
    const variant = reason.__typename !== undefined ? BADGE_VARIANT_BY_REASON[reason.__typename] : undefined
    return (
        <Badge
            key={reason.title}
            tooltip={reason.description}
            className={styles.badge}
            variant={variant}
            {...linkProps}
        >
            {reason.title}
        </Badge>
    )
}
