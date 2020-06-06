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
    patchsetID: GQL.ID | null
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
 * A page for updating a campaign.
 *
 * TODO(sqs): warn before navigating away if that would result in data loss
 */
export const UpdateCampaignPage: React.FunctionComponent<Props> = ({
    campaign,
    patchsetID,
    history,
    location,
    isLightTheme,
    fetchPatchSetById,
    queryPatchFileDiffs,
    queryPatchesFromCampaign,
    queryPatchesFromPatchSet,
    queryChangesets,
    _updateCampaign = updateCampaign,
}) => {
    const [description, setDescription] = useState<string>('')

    const [branch, setBranch] = useState<string>('')
    const [branchModified, setBranchModified] = useState<boolean>(false)
    const onBranchChange = useCallback((newValue: string): void => {
        setBranch(newValue)
        setBranchModified(true)
    }, [])

    // The branch is only editable when no changesets have been published or are being published.
    const specifyingBranchAllowed =
        campaign.changesets.totalCount === 0 && campaign.status.state !== GQL.BackgroundProcessState.PROCESSING

    const [name, setName] = useState<string>('')
    const onNameChange = useCallback(
        (newName: string): void => {
            if (!branchModified) {
                setBranch(slugify(newName, { remove: /[!"'()*+.:@\\^~]/g, lower: true }))
            }
            setName(newName)
        },
        [branchModified]
    )

    const patchset = useObservable(
        useMemo(() => (patchsetID ? fetchPatchSetById(patchsetID) : NEVER), [patchsetID, fetchPatchSetById])
    )

    const [onSubmit, submitOrError] = useEventObservable(
        useCallback(
            (submits: Observable<React.FormEvent>) =>
                submits.pipe(
                    tap(event => event.preventDefault()),
                    switchMap(() =>
                        from(
                            _updateCampaign({
                                id: campaign.id,
                                name,
                                description,
                                branch,
                            })
                        ).pipe(
                            catchError(err => of(asError(err))),
                            startWith(LOADING)
                        )
                    )
                ),
            [_updateCampaign, branch, campaign.id, description, name]
        )
    )
    const isLoading = submitOrError === LOADING

    // TODO(sqs): show update preview
    // TODO(sqs): only allow branch update if no changesets published
    return (
        <>
            <PageTitle title={`Update campaign - ${campaign.name}`} />
            <h1>Update campaign</h1>
            {submitOrError !== undefined && submitOrError !== LOADING ? (
                isErrorLike(submitOrError) ? (
                    <ErrorAlert error={submitOrError} history={history} />
                ) : (
                    <Redirect to={submitOrError.url} />
                )
            ) : null}
            <Form onSubmit={onSubmit} className="e2e-update-campaign-form">
                <CampaignTitleField value={name} onChange={onNameChange} disabled={isLoading} />
                <CampaignDescriptionField value={description} onChange={setDescription} disabled={isLoading} />
                <CampaignBranchField
                    value={branch}
                    onChange={onBranchChange}
                    disabled={isLoading || !specifyingBranchAllowed}
                />
                <div className="form-group d-flex align-items-center">
                    <button type="submit" className="btn btn-primary mr-2 e2e-campaign-update-btn" disabled={isLoading}>
                        Update campaign
                    </button>
                </div>
            </Form>
            <hr />
            <h2>Preview</h2>
            {patchsetID &&
                (patchset === undefined ? (
                    'Loading patchset TODO(sqs)'
                ) : patchset === null ? (
                    'patchset not found TODO(sqs)'
                ) : (
                    <CampaignUpdateDiff
                        campaign={campaign}
                        patchSet={patchset}
                        queryChangesets={queryChangesets}
                        queryPatchesFromCampaign={queryPatchesFromCampaign}
                        queryPatchesFromPatchSet={queryPatchesFromPatchSet}
                        queryPatchFileDiffs={queryPatchFileDiffs}
                        history={history}
                        location={location}
                        isLightTheme={isLightTheme}
                        className="mt-4"
                    />
                ))}
        </>
    )
}
