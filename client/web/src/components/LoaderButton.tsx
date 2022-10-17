import React from 'react'

import classNames from 'classnames'

import { LoadingSpinner, Button, ButtonProps } from '@sourcegraph/wildcard'

interface Props extends ButtonProps {
    loading: boolean
    label: string
    alwaysShowLabel: boolean
}

export const LoaderButton: React.FunctionComponent<React.PropsWithChildren<Partial<Props>>> = ({
    loading,
    label,
    alwaysShowLabel,
    ...props
}) => (
    <Button {...props} className={classNames(props.className, 'd-flex justify-content-center align-items-center')}>
        {loading ? (
            <>
                <LoadingSpinner />
                {alwaysShowLabel && <span className="ml-1">{label}</span>}
            </>
        ) : (
            label
        )}
    </Button>
)
