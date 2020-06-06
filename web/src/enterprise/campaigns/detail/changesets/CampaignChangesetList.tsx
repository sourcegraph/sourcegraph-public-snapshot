import H from 'history'
import React, { useCallback } from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { ErrorLike } from '../../../../../../shared/src/util/errors'
import { ConnectionListFilterContext } from '../../../../components/connectionList/ConnectionListFilterDropdownButton'
import { QueryParameterProps } from '../../../../util/useQueryParameter'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { ChangesetList } from './ChangesetList'
import { ChangesetListHeaderStates } from './ChangesetListHeaderStates'
import { Link } from 'react-router-dom'
import { ChangesetListHeaderCommonFilters } from './CampaignChangesetListCommonFilters'

interface Props extends QueryParameterProps, ExtensionsControllerProps, PlatformContextProps, ThemeProps {
    changesets: undefined | GQL.IChangesetConnection | ErrorLike
    onChangesetsUpdate: () => void
    campaign: Pick<GQL.ICampaign, 'id' | 'viewerCanAdminister'>
    action: React.ReactFragment

    className?: string
    location: H.Location
    history: H.History
}

export const CampaignChangesetList: React.FunctionComponent<Props> = ({
    changesets,
    onChangesetsUpdate,
    campaign,
    action,
    className = '',
    query,
    onQueryChange,
    locationWithQuery,
    extensionsController,
    ...props
}) => {
    const filterProps: ConnectionListFilterContext<GQL.IChangesetConnectionFilters> = {
        connection: changesets,
        query,
        onQueryChange,
        locationWithQuery,
    }

    return (
        <div className={`campaign-changeset-list ${className}`}>
            <ChangesetList
                {...props}
                changesets={changesets}
                campaign={campaign}
                query={query}
                onQueryChange={onQueryChange}
                locationWithQuery={locationWithQuery}
                headerItems={{
                    left: <ChangesetListHeaderStates {...props} changesets={changesets} query={query} />,
                    right: (
                        <>
                            <ChangesetListHeaderCommonFilters {...filterProps} />
                        </>
                    ),
                }}
                extensionsController={extensionsController}
            />
        </div>
    )
}
