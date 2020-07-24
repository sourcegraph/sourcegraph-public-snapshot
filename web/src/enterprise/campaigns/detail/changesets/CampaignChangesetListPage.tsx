import React, { useMemo } from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { MinimalCampaign, MinimalChangeset } from '../CampaignArea'
import { CampaignChangesetsEditButton } from './CampaignChangesetsEditButton'
import { CampaignChangesets } from './CampaignChangesets'
import { queryChangesets } from '../backend'
import { CampaignDiffStat } from '../CampaignDiffStat'
import H from 'history'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../../../shared/src/telemetry/telemetryService'
import { useObservable } from '../../../../../../shared/src/util/useObservable'
import { useQueryParameter } from '../../../../util/useQueryParameter'
import { CampaignChangesetList } from './CampaignChangesetList'
import { startWith } from 'rxjs/operators'
import { CampaignChangesetListCommonQueriesButtonDropdown } from './list/filters/CampaignChangesetListCommonQueriesButtonDropdown'
import { ConnectionListFilterQueryInput } from '../../../../components/connectionList/ConnectionListFilterQueryInput'

interface Props extends ThemeProps, ExtensionsControllerProps, PlatformContextProps, TelemetryProps {
    campaign: MinimalCampaign
    history: H.History
    location: H.Location

    queryChangesets: typeof queryChangesets
}

export const CampaignChangesetListPage: React.FunctionComponent<Props> = ({
    campaign,
    history,
    location,
    queryChangesets,
    ...props
}) => {
    const [query, onQueryChange, locationWithQuery] = useQueryParameter({ history, location })

    const queryArguments: GQL.IChangesetsOnCampaignArguments = useMemo(() => {
        const parameters = new URLSearchParams(query)
        return {
            state: parameters.get('state') as GQL.ChangesetState | null,
            checkState: parameters.get('checkState') as GQL.ChangesetCheckState | null,
            reviewState: parameters.get('reviewState') as GQL.ChangesetReviewState | null,
            // TODO(sqs): `first` param?
        }
    }, [query])
    const changesets = useObservable(
        useMemo(() => queryChangesets(campaign.id, queryArguments).pipe(startWith(undefined)), [
            campaign.id,
            queryArguments,
            queryChangesets,
        ])
    )

    return (
        <>
            <a id="changesets" />
            <header className="d-flex align-items-center mb-2">
                <ConnectionListFilterQueryInput
                    query={query}
                    onQueryChange={onQueryChange}
                    locationWithQuery={locationWithQuery}
                    beforeInputFragment={
                        <div className="input-group-prepend">
                            <CampaignChangesetListCommonQueriesButtonDropdown locationWithQuery={locationWithQuery} />
                        </div>
                    }
                    className="flex-1"
                />
            </header>
            <CampaignChangesetList
                campaign={campaign}
                changesets={changesets}
                history={history}
                location={location}
                query={query}
                onQueryChange={onQueryChange}
                locationWithQuery={locationWithQuery}
                {...props}
            />
        </>
    )
}
