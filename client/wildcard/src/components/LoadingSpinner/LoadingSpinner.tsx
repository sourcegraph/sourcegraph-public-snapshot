import { FC, SVGProps } from 'react'

import classNames from 'classnames'

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

export const LoadingSpinner: FC<LoadingSpinnerProps> = props => {
    const { inline = true, className, ...attribute } = props

    return (
        <Icon
            as="div"
            inline={inline}
            aria-label="Loading"
            aria-live="polite"
            className={classNames(styles.loadingSpinner, className)}
            data-loading-spinner={true}
            {...attribute}
        />
    )
}
