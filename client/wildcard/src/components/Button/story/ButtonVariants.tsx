import { startCase } from 'lodash'
import React from 'react'
import 'storybook-addon-designs'

import { Button, ButtonProps } from '../Button'
import { BUTTON_VARIANTS } from '../constants'

import styles from './ButtonVariants.module.scss'

interface ButtonVariantsProps extends Pick<ButtonProps, 'size' | 'outline' | 'as'> {
    variants: readonly typeof BUTTON_VARIANTS[number][]
    icon?: React.ComponentType<{ className?: string }>
}

export const ButtonVariants: React.FunctionComponent<ButtonVariantsProps> = ({
    variants,
    size,
    outline,
    icon: Icon,
}) => (
    <div className={styles.grid}>
        {variants.map(variant => (
            <React.Fragment key={variant}>
                <Button variant={variant} size={size} outline={outline} onClick={console.log}>
                    {Icon && <Icon className="icon-inline mr-1" />}
                    {startCase(variant)}
                </Button>
                <Button variant={variant} size={size} outline={outline} onClick={console.log} className="focus">
                    {Icon && <Icon className="icon-inline mr-1" />}
                    Focus
                </Button>
                <Button variant={variant} size={size} outline={outline} onClick={console.log} disabled={true}>
                    {Icon && <Icon className="icon-inline mr-1" />}
                    Disabled
                </Button>
            </React.Fragment>
        ))}
    </div>
)
