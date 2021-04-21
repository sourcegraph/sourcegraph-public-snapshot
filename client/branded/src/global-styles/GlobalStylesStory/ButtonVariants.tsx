import { action } from '@storybook/addon-actions'
import classNames from 'classnames'
import { flow, startCase } from 'lodash'
import SearchIcon from 'mdi-react/SearchIcon'
import React from 'react'
import 'storybook-addon-designs'

import styles from './ButtonVariants.module.scss'
import { SEMANTIC_COLORS } from './constants'
import { preventDefault } from './utils'

type VariantType = 'btn' | 'btn-outline'

const variants: Record<VariantType, readonly string[]> = {
    btn: SEMANTIC_COLORS,
    'btn-outline': ['primary', 'secondary', 'danger'],
}

interface ButtonVariantsProps {
    variantType?: VariantType
}

export const ButtonVariants: React.FunctionComponent<ButtonVariantsProps> = ({ variantType = 'btn' }) => (
    <div className={styles.grid}>
        {variants[variantType].map(variant => (
            <React.Fragment key={variant}>
                <button
                    type="button"
                    key={variant}
                    className={classNames('btn', `${variantType}-${variant}`)}
                    onClick={flow(preventDefault, action('button clicked'))}
                >
                    {startCase(variant)}
                </button>
                <button
                    type="button"
                    key={`${variantType} - ${variant} - focus`}
                    className={classNames('btn', `${variantType}-${variant}`, 'focus')}
                >
                    Focus
                </button>
                <button
                    type="button"
                    key={`${variantType} - ${variant} - disabled`}
                    className={classNames('btn', `${variantType}-${variant}`)}
                    disabled={true}
                >
                    Disabled
                </button>
                <button
                    type="button"
                    key={`${variantType} - ${variant} - icon`}
                    className={classNames('btn', `${variantType}-${variant}`)}
                >
                    <SearchIcon className="icon-inline mr-1" />
                    With icon
                </button>
            </React.Fragment>
        ))}
    </div>
)
