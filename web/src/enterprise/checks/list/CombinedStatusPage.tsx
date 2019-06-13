import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React from 'react'
import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { HeroPage } from '../../../components/HeroPage'
import { useChecksForScope } from '../util/useChecksForScope'
import { CombinedStatus } from './CombinedStatus'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    /** The check scope. */
    scope: sourcegraph.CheckScope | sourcegraph.WorkspaceRoot

    checksURL: string
    location: H.Location
    history: H.History
    authenticatedUser: GQL.IUser | null
    isLightTheme: boolean
}

const LOADING: 'loading' = 'loading'

/**
 * A page showing a combined status for a particular scope.
 */
export const CombinedStatusPage: React.FunctionComponent<Props> = ({ scope, ...props }) => {
    const statusesOrError = useChecksForScope(props.extensionsController, scope)
    if (statusesOrError === LOADING) {
        return null // loading
    }
    if (isErrorLike(statusesOrError)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={statusesOrError.message} />
    }
    return <CombinedStatus {...props} statuses={statusesOrError} itemClassName="text-truncate" />
}
