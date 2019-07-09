import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { HeroPage } from '../../../../components/HeroPage'
import { ChecksList } from '../../list/ChecksList'
import { useChecksForScope } from '../../util/useChecksForScope'
import { ChecksAreaContext } from '../ScopeChecksArea'

interface Props
    extends Pick<ChecksAreaContext, 'scope' | 'checksURL'>,
        ExtensionsControllerProps,
        PlatformContextProps {
    location: H.Location
    history: H.History
    authenticatedUser: GQL.IUser | null
    isLightTheme: boolean
}

const LOADING: 'loading' = 'loading'

/**
 * A page showing a list of checks for a particular scope.
 */
export const ChecksListPage: React.FunctionComponent<Props> = ({ scope, ...props }) => {
    const checksOrError = useChecksForScope(props.extensionsController, scope)
    if (checksOrError === LOADING) {
        return null // loading
    }
    if (isErrorLike(checksOrError)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={checksOrError.message} />
    }
    return <ChecksList {...props} checksInfoOrError={checksOrError} itemClassName="text-truncate" />
}
