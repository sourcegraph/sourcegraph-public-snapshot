import classNames from 'classnames'
import { startCase } from 'lodash'
import React from 'react'
import 'storybook-addon-designs'

import { SEMANTIC_COLORS } from '../constants'

import styles from './BadgeVariants.module.scss'

interface BadgeProps {
    variant?: string
}

const Badge: React.FunctionComponent<BadgeProps> = ({ variant }) => {
    const className = classNames('badge', variant && `badge-${variant}`)
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
    variants: readonly typeof SEMANTIC_COLORS[number][]
}

export const BadgeVariants: React.FunctionComponent<BadgeVariantProps> = ({ variants }) => (
    <div className={styles.grid}>
        <Badge />
        {variants.map(variant => (
            <Badge key={variant} variant={variant} />
        ))}
    </div>
)
