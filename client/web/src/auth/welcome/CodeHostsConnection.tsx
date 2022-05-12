import React, { useEffect } from 'react'

import { ErrorLike } from '@sourcegraph/common'
import { LoadingSpinner, Typography } from '@sourcegraph/wildcard'

import { ListExternalServiceFields } from '../../graphql-operations'
import { UserCodeHosts } from '../../user/settings/codeHosts/UserCodeHosts'
import { useSteps } from '../Steps'

interface CodeHostsConnection extends Omit<UserCodeHosts, 'onDidRemove' | 'onDidError' | 'externalServices'> {
    refetch: UserCodeHosts['onDidRemove']
    loading: boolean
    onError: (error: ErrorLike) => void
    onNavigation?: (called: boolean) => void
    externalServices: ListExternalServiceFields[] | undefined
}

export const CodeHostsConnection: React.FunctionComponent<React.PropsWithChildren<CodeHostsConnection>> = ({
    user,
    context,
    refetch,
    externalServices,
    loading,
    onNavigation,
    onError,
}) => {
    const { setComplete, currentIndex, resetToTheRight } = useSteps()

    useEffect(() => {
        if (Array.isArray(externalServices) && externalServices.length > 0) {
            setComplete(currentIndex, true)
        } else {
            setComplete(currentIndex, false)
            resetToTheRight(currentIndex)
        }
    }, [currentIndex, externalServices, resetToTheRight, setComplete])

    if (loading || !externalServices) {
        return (
            <div className="d-flex justify-content-center">
                <LoadingSpinner />
            </div>
        )
    }

    return (
        <>
            <div className="mb-4 mt-3">
                <Typography.H3>Connect with code hosts</Typography.H3>
                <p className="text-muted">
                    Connect with providers where your source code is hosted. Then, choose the repositories you’d like to
                    search with Sourcegraph.
                </p>
            </div>
            <UserCodeHosts
                user={user}
                externalServices={externalServices}
                context={context}
                onNavigation={onNavigation}
                onDidError={onError}
                onDidRemove={refetch}
            />
        </>
    )
}
