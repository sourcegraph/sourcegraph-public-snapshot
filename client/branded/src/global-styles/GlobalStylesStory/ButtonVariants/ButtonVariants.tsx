import { action } from '@storybook/addon-actions'
import classNames from 'classnames'
import { flow, startCase } from 'lodash'
import React from 'react'
import 'storybook-addon-designs'

import styles from './ButtonVariants.module.scss'
import { SEMANTIC_COLORS } from './constants'
import { preventDefault } from './utils'

interface ButtonVariantsProps {
    variantType?: 'btn' | 'btn-outline'
    variants: readonly typeof SEMANTIC_COLORS[number][]
    small?: boolean
    icon?: React.ComponentType<{ className?: string }>
}

export const ButtonVariants: React.FunctionComponent<ButtonVariantsProps> = ({
    variantType = 'btn',
    variants,
    small,
    icon: Icon,
}) => (
    <div className={styles.grid}>
        {variants.map(variant => {
            const className = classNames('btn', `${variantType}-${variant}`, small && 'btn-sm')
            return (
                <React.Fragment key={variant}>
                    <button
                        type="button"
                        key={variant}
                        className={className}
                        onClick={flow(preventDefault, action('button clicked'))}
                    >
                        {Icon && <Icon className="icon-inline mr-1" />}
                        {startCase(variant)}
                    </button>
                    <button
                        type="button"
                        key={`${variantType} - ${variant} - focus`}
                        className={classNames(className, 'focus')}
                    >
                        {Icon && <Icon className="icon-inline mr-1" />}
                        Focus
                    </button>
                    <button
                        type="button"
                        key={`${variantType} - ${variant} - disabled`}
                        className={className}
                        disabled={true}
                    >
                        {Icon && <Icon className="icon-inline mr-1" />}
                        Disabled
                    </button>
                </React.Fragment>
            )
        })}
    </div>
)
