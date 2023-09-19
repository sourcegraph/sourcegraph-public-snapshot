import { forwardRef, ReactNode } from 'react'

import classNames from 'classnames'

import { LoadingSpinner, Button, type ButtonProps } from '@sourcegraph/wildcard'

export interface LoaderButtonProps extends ButtonProps {
    loading?: boolean
    label?: ReactNode
    alwaysShowLabel?: boolean
    icon?: JSX.Element
}

export const LoaderButton = forwardRef<HTMLButtonElement, LoaderButtonProps>((props, ref) => {
    const { loading, label, alwaysShowLabel, icon, ...otherProps } = props

    return (
        <Button
            ref={ref}
            {...otherProps}
            className={classNames(props.className, 'd-flex justify-content-center align-items-center')}
        >
            {loading ? (
                <>
                    <LoadingSpinner />
                    {alwaysShowLabel && <span className="ml-1">{label}</span>}
                </>
            ) : icon ? (
                <>
                    {icon}
                    {label && <>&nbsp;{label}</>}
                </>
            ) : (
                label
            )}
        </Button>
    )
})
