import React, { useMemo } from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { MinimalCampaign, MinimalPatchSet } from '../CampaignArea'
import { Link } from '../../../../../../shared/src/components/Link'
import { CampaignChangesetsAddExistingButton } from './CampaignChangesetsAddExistingButton'
import { CampaignChangesetsEditButton } from './CampaignChangesetsEditButton'
import { CampaignChangesets } from './CampaignChangesets'
import {
    queryChangesets,
    fetchPatchSetById,
    queryPatchesFromCampaign,
    queryPatchesFromPatchSet,
    queryPatchFileDiffs,
} from '../backend'
import { CampaignDiffStat } from '../CampaignDiffStat'
import H from 'history'
import { Observable, Subject, NEVER } from 'rxjs'
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

    fetchPatchSetById: typeof fetchPatchSetById | ((patchSet: GQL.ID) => Observable<MinimalPatchSet | null>)
    queryPatchesFromCampaign: typeof queryPatchesFromCampaign
    queryPatchesFromPatchSet: typeof queryPatchesFromPatchSet
    queryPatchFileDiffs: typeof queryPatchFileDiffs
    queryChangesets: typeof queryChangesets
}

export const CampaignChangesetListPage: React.FunctionComponent<Props> = ({
    campaign,
    history,
    location,
    queryChangesets,
    ...props
}) => {
    const patchSet = useObservable(
        useMemo(() => (campaign.patchSet?.id ? fetchPatchSetById(campaign.patchSet.id) : NEVER), [
            campaign.patchSet,
            fetchPatchSetById,
        ])
    )

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
                    className="flex-1 mr-5"
                />
                <CampaignChangesetsAddExistingButton
                    campaign={campaign}
                    buttonClassName="btn btn-secondary mr-2 pr-1"
                    history={history}
                />
                <CampaignChangesetsEditButton campaign={campaign} buttonClassName="btn btn-secondary pr-1" />
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
                action={
                    campaign.viewerCanAdminister && (
                        <div className="d-flex align-items-center">
                            {patchSet && (
                                <CampaignDiffStat campaign={campaign} patchSet={patchSet} className="ml-2 mr-2 mb-0" />
                            )}
                            <button type="button" className="btn btn-secondary">
                                Publish all
                            </button>
                        </div>
                    )
                }
            />
        </>
    )
}
