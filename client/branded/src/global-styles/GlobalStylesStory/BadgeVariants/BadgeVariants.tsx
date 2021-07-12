import classNames from 'classnames'
import { startCase } from 'lodash'
import React from 'react'
import 'storybook-addon-designs'

import { SEMANTIC_COLORS } from '../constants'

import styles from './BadgeVariants.module.scss'

interface BadgeProps {
    variant?: string
    small?: boolean
}

const Badge: React.FunctionComponent<BadgeProps> = ({ variant, small }) => {
    const className = classNames('badge', small && 'badge-sm', variant && `badge-${variant}`)
    return (
        <>
            <span className={className}>{startCase(variant || 'Default')}</span>
            <span className={classNames(className, 'text-uppercase')}>Uppercase</span>
            <a href="/" className={className}>
                Link
            </a>
        </>
    )
}

interface BadgeVariantProps {
    variants?: readonly (typeof SEMANTIC_COLORS[number] | 'outline-secondary')[]
    small?: boolean
}

export const BadgeVariants: React.FunctionComponent<BadgeVariantProps> = ({ variants, small }) => (
    <div className={styles.grid}>
        <Badge small={small} />
        {variants?.map(variant => (
            <Badge key={variant} small={small} variant={variant} />
        ))}
    </div>
)
