import React, { useEffect, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'
import { getViewsForContainer } from '../../../../shared/src/api/client/services/viewService'
import { ContributableViewContainer } from '../../../../shared/src/api/protocol'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { SearchPatternType } from '../../../../shared/src/graphql/schema'
import { VersionContextProps } from '../../../../shared/src/search/util'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { PageTitle } from '../../components/PageTitle'
import { GraphSelectionProps } from '../../enterprise/graphs/selector/graphSelectionProps'
import { ViewGrid } from '../../repo/tree/ViewGrid'
import { CaseSensitivityProps, CopyQueryButtonProps, PatternTypeProps } from '../../search'
import { eventLogger } from '../../tracking/eventLogger'
import { UserAreaRouteContext } from './UserArea'
import H from 'history'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'

interface Props
    extends Pick<UserAreaRouteContext, 'user'>,
        ExtensionsControllerProps<'services'>,
        SettingsCascadeProps,
        Pick<PatternTypeProps, 'patternType'>,
        Pick<CaseSensitivityProps, 'caseSensitive'>,
        CopyQueryButtonProps,
        VersionContextProps,
        Pick<GraphSelectionProps, 'selectedGraph'> {
    globbing: boolean // TODO(sqs): refactor this
    location: H.Location
    history: H.History
}

/**
 * The user overview page.
 */
export const UserOverviewPage: React.FunctionComponent<Props> = ({
    user,
    extensionsController: { services },
    ...props
}) => {
    useEffect(() => eventLogger.logViewEvent('UserOverview'), [])

    const views = useObservable(
        useMemo(
            () =>
                getViewsForContainer(
                    ContributableViewContainer.Profile,
                    {
                        id: user.id,
                        type: user.__typename,
                    },
                    services.view
                ),
            [services.view, user.__typename, user.id]
        )
    )

    return (
        <div className="user-page user-overview-page">
            <PageTitle title={user.username} />
            {views && <ViewGrid {...props} className="mb-5" viewGridStorageKey="user-profile-page" views={views} />}
        </div>
    )
}
