import React, { useState, useCallback, useMemo } from 'react'
import H from 'history'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PageTitle } from '../../../../components/PageTitle'
import { CampaignTitleField } from '../form/CampaignTitleField'
import { CampaignDescriptionField } from '../form/CampaignDescriptionField'
import { CampaignBranchField } from '../form/CampaignBranchField'
import { isErrorLike, asError } from '../../../../../../shared/src/util/errors'
import { ErrorAlert } from '../../../../components/alerts'
import { useEventObservable, useObservable } from '../../../../../../shared/src/util/useObservable'
import { startWith, switchMap, catchError, tap } from 'rxjs/operators'
import { from, Observable, of, NEVER } from 'rxjs'
import { Redirect } from 'react-router'
import slugify from 'slugify'
import {
    updateCampaign,
    fetchPatchSetById,
    queryPatchFileDiffs,
    queryChangesets,
    queryPatchesFromPatchSet,
    queryPatchesFromCampaign,
} from '../backend'
import { Form } from '../../../../components/Form'
import { MinimalPatchSet, MinimalCampaign } from '../CampaignArea'
import { CampaignUpdateDiff } from '../CampaignUpdateDiff'
import { ThemeProps } from '../../../../../../shared/src/theme'

interface Props extends ThemeProps {
    campaign: Pick<
        MinimalCampaign,
        'id' | 'name' | 'description' | 'branch' | 'viewerCanAdminister' | 'changesets' | 'patches' | 'status'
    >
    location: H.Location
    history: H.History

    fetchPatchSetById: typeof fetchPatchSetById | ((patchSet: GQL.ID) => Observable<MinimalPatchSet | null>)
    queryPatchFileDiffs: typeof queryPatchFileDiffs
    queryPatchesFromCampaign: typeof queryPatchesFromCampaign
    queryPatchesFromPatchSet: typeof queryPatchesFromPatchSet
    queryChangesets: typeof queryChangesets

    /** For testing only. */
    _updateCampaign?: typeof updateCampaign
}

const LOADING = 'loading' as const

/**
 * A page for closing (or deleting) a campaign.
 *
 * TODO(sqs): add a way to add a comment to all PRs when closing
 */
export const CloseCampaignPage: React.FunctionComponent<Props> = ({
    campaign,
    history,
    location,
    isLightTheme,
    queryPatchFileDiffs,
    queryPatchesFromCampaign,
    queryPatchesFromPatchSet,
    queryChangesets,
    _updateCampaign = updateCampaign,
}) => {
    const [onSubmit, submitOrError] = useEventObservable(
        useCallback(
            (submits: Observable<React.FormEvent>) =>
                submits.pipe(
                    tap(event => event.preventDefault()),
                    switchMap(() =>
                        from(
                            _updateCampaign({
                                id: campaign.id,
                                state: GQL.CampaignState.CLOSED,
                            })
                        ).pipe(
                            catchError(err => of(asError(err))),
                            startWith(LOADING)
                        )
                    )
                ),
            [_updateCampaign, campaign.id]
        )
    )
    const isLoading = submitOrError === LOADING

    const [closeChangesets, setCloseChangesets] = useState(false)
    const onCloseChangesetsChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        event => setCloseChangesets(event.currentTarget.checked),
        []
    )

    // TODO(sqs): show update preview
    // TODO(sqs): allow permanently deleting *after* closed, but not before, to reduce confusion
    return (
        <>
            <PageTitle title={`Close campaign - ${campaign.name}`} />
            <h1>Close campaign</h1>
            {submitOrError !== undefined && submitOrError !== LOADING ? (
                isErrorLike(submitOrError) ? (
                    <ErrorAlert error={submitOrError} history={history} />
                ) : (
                    <Redirect to={submitOrError.url} />
                )
            ) : null}
            <Form onSubmit={onSubmit} className="e2e-close-campaign-form">
                <div className="form-check">
                    <input
                        className="form-check-input"
                        type="checkbox"
                        checked={closeChangesets}
                        onChange={onCloseChangesetsChange}
                        disabled={isLoading}
                        id="campaignCloseChangesets"
                    />
                </div>
                <div className="form-group d-flex align-items-center">
                    <button type="submit" className="btn btn-primary mr-2 e2e-campaign-close-btn" disabled={isLoading}>
                        Close campaign
                    </button>
                </div>
            </Form>
            <hr />
            <h2>Preview</h2>
            {/* TODO(sqs): factor out preview */}
            <CampaignUpdateDiff
                campaign={campaign}
                patchSet={{} as any /* TODO(sqs) */}
                queryChangesets={queryChangesets}
                queryPatchesFromCampaign={queryPatchesFromCampaign}
                queryPatchesFromPatchSet={queryPatchesFromPatchSet}
                queryPatchFileDiffs={queryPatchFileDiffs}
                history={history}
                location={location}
                isLightTheme={isLightTheme}
                className="mt-4"
            />
        </>
    )
}
