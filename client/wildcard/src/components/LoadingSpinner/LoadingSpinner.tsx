import { type SVGProps, forwardRef } from 'react'

import classNames from 'classnames'

import type { ForwardReferenceComponent } from '../../types'
import { Icon } from '../Icon'

import styles from './LoadingSpinner.module.scss'

export interface LoadingSpinnerProps extends SVGProps<SVGSVGElement> {
    /**
     * Whether to show loading spinner with icon-inline
     *
     * @default true
     */
    inline?: boolean
}

export const LoadingSpinner = forwardRef(function LoadingSpinner(props, reference) {
    const { inline = true, className, ...attribute } = props

    return (
        <Icon
            as="div"
            inline={inline}
            aria-label="Loading"
            aria-live="polite"
            className={classNames(styles.loadingSpinner, className)}
            data-loading-spinner={true}
            ref={reference}
            {...attribute}
        />
    )
}) as ForwardReferenceComponent<'div', LoadingSpinnerProps>
