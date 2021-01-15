import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import classNames from 'classnames'

interface Props extends React.DetailedHTMLProps<React.ButtonHTMLAttributes<HTMLButtonElement>, HTMLButtonElement> {
    loading: boolean
    label: string
    spinnerClassName: string
    alwaysShowLabel: boolean
}

export const LoaderButton: React.FunctionComponent<Partial<Props>> = ({
    loading,
    label,
    spinnerClassName,
    alwaysShowLabel,
    ...props
}) => (
    // eslint-disable-next-line react/button-has-type
    <button {...props} type={props.type ?? 'button'}>
        {loading ? (
            <>
                <LoadingSpinner className={classNames(spinnerClassName, 'icon-inline')} /> {alwaysShowLabel && label}
            </>
        ) : (
            label
        )}
    </button>
)
