import React from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

interface Props extends React.DetailedHTMLProps<React.ButtonHTMLAttributes<HTMLButtonElement>, HTMLButtonElement> {
    loading: boolean
    label: string
    alwaysShowLabel: boolean
}

export const LoaderButton: React.FunctionComponent<Partial<Props>> = ({
    loading,
    label,
    alwaysShowLabel,
    ...props
}) => (
    // eslint-disable-next-line react/button-has-type
    <button {...props} type={props.type ?? 'button'}>
        {loading ? (
            <>
                <LoadingSpinner className="icon-inline" /> {alwaysShowLabel && label}
            </>
        ) : (
            label
        )}
    </button>
)
