import { action } from '@storybook/addon-actions'
import classNames from 'classnames'
import { flow, startCase } from 'lodash'
import React from 'react'
import 'storybook-addon-designs'

import styles from './ButtonVariants.module.scss'
import { preventDefault } from './utils'

interface ButtonVariantsProps {
    variantType?: 'btn' | 'btn-outline'
    variants: readonly string[]
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
        {variants.map(variant => (
            <React.Fragment key={variant}>
                <button
                    type="button"
                    key={variant}
                    className={classNames('btn', `${variantType}-${variant}`, small && 'btn-sm')}
                    onClick={flow(preventDefault, action('button clicked'))}
                >
                    {Icon && <Icon className="icon-inline mr-1" />}
                    {startCase(variant)}
                </button>
                <button
                    type="button"
                    key={`${variantType} - ${variant} - focus`}
                    className={classNames('btn', `${variantType}-${variant}`, small && 'btn-sm', 'focus')}
                >
                    {Icon && <Icon className="icon-inline mr-1" />}
                    Focus
                </button>
                <button
                    type="button"
                    key={`${variantType} - ${variant} - disabled`}
                    className={classNames('btn', `${variantType}-${variant}`, small && 'btn-sm')}
                    disabled={true}
                >
                    {Icon && <Icon className="icon-inline mr-1" />}
                    Disabled
                </button>
            </React.Fragment>
        ))}
    </div>
)
