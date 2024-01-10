import React from 'react'

import { startCase } from 'lodash'

import '@storybook/addon-designs'

import { logger } from '@sourcegraph/common'

import { Icon } from '../../Icon'
import { Button, type ButtonProps } from '../Button'
import type { BUTTON_VARIANTS } from '../constants'

import styles from './ButtonVariants.module.scss'

interface ButtonVariantsProps extends Pick<ButtonProps, 'size' | 'outline'> {
    variants: readonly typeof BUTTON_VARIANTS[number][]
    icon?: React.ComponentType<{ className?: string }>
}

export const ButtonVariants: React.FunctionComponent<React.PropsWithChildren<ButtonVariantsProps>> = ({
    variants,
    size,
    outline,
    icon: ButtonIcon,
}) => (
    <div className={styles.grid}>
        {variants.map(variant => (
            <React.Fragment key={variant}>
                <Button variant={variant} size={size} outline={outline} onClick={logger.log}>
                    {ButtonIcon && <Icon aria-hidden={true} as={ButtonIcon} className="mr-1" />}
                    {startCase(variant)}
                </Button>
                <Button variant={variant} size={size} outline={outline} onClick={logger.log} className="focus">
                    {ButtonIcon && <Icon aria-hidden={true} as={ButtonIcon} className="mr-1" />}
                    Focus
                </Button>
                <Button variant={variant} size={size} outline={outline} onClick={logger.log} disabled={true}>
                    {ButtonIcon && <Icon aria-hidden={true} as={ButtonIcon} className="mr-1" />}
                    Disabled
                </Button>
            </React.Fragment>
        ))}
    </div>
)
