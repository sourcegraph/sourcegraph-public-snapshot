import classnames from 'classnames'
import React from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { Button, ButtonProps } from '../Button'

interface LoadingButtonProps extends ButtonProps {
    loading: boolean
    alwaysShowChildren: boolean
    spinnerClassName?: string
}

export const LoadingButton: React.FunctionComponent<Partial<LoadingButtonProps>> = ({
    loading,
    children,
    alwaysShowChildren,
    spinnerClassName,
    ...props
}) => (
    <Button {...props}>
        {loading ? (
            <>
                <LoadingSpinner className={classnames(spinnerClassName, 'icon-inline')} />{' '}
                {alwaysShowChildren && children}
            </>
        ) : (
            children
        )}
    </Button>
)
