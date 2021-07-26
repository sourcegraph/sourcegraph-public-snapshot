import classnames from 'classnames'
import React from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

import { Button, ButtonProps } from '../Button'

interface LoadingButtonProps extends ButtonProps {
    /**
     * Loading state. If `true` a loading spinner will be rendered.
     */
    loading?: boolean
    /**
     * Always render `children` even if `loading` is `true`.
     */
    alwaysShowChildren?: boolean
    spinnerClassName?: string
}

/**
 * Loading button.
 *
 * A wrapper around the `<Button />` component that allows consumers to render a loading spinner consistently.
 *
 * This component should typically be used for asynchronous actions (e.g. fetching data from a remote server).
 *
 * Be mindful of how the output of the action is signalled for the user.
 * For example, it may be appropriate to use the 'success' variant once the action has completed successfully.
 * Likewise, you may want to use the 'danger' variant if the action failed.
 */
export const LoadingButton: React.FunctionComponent<LoadingButtonProps> = ({
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
