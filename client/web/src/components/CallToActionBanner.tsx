import type { FunctionComponent, ReactNode } from 'react'

import { mdiArrowRight } from '@mdi/js'
import classNames from 'classnames'

import { Icon, Text } from '@sourcegraph/wildcard'

import styles from './CallToActionBanner.module.scss'

export interface CloudCtaBanner {
    variant?: 'filled' | 'outlined' | 'underlined' | string | undefined
    small?: boolean
    className?: string
    children: ReactNode
}

export const CallToActionBanner: FunctionComponent<CloudCtaBanner> = ({ variant, small, className, children }) => (
    <section
        className={classNames(className, 'd-flex justify-content-center', {
            [styles.filled]: variant === 'filled',
            [styles.outlined]: variant === 'outlined',
            [styles.underlined]: variant === 'underlined',
        })}
    >
        <Icon className="mr-2 text-merged" size={small ? 'sm' : 'md'} aria-hidden={true} svgPath={mdiArrowRight} />

        <Text size={small ? 'small' : 'base'} className="my-auto">
            {children}
        </Text>
    </section>
)
