import React, { useState, useCallback } from 'react'
import H from 'history'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PageTitle } from '../../../components/PageTitle'
import { CampaignTitleField } from '../detail/form/CampaignTitleField'
import { CampaignDescriptionField } from '../detail/form/CampaignDescriptionField'
import { CampaignBranchField } from '../detail/form/CampaignBranchField'
import { isErrorLike, asError } from '../../../../../shared/src/util/errors'
import { ErrorAlert } from '../../../components/alerts'
import { useEventObservable } from '../../../../../shared/src/util/useObservable'
import { startWith, switchMap, catchError, tap } from 'rxjs/operators'
import { from, Observable, of } from 'rxjs'
import { Redirect } from 'react-router'
import slugify from 'slugify'
import { createCampaign } from '../detail/backend'
import { Form } from '../../../components/Form'

interface Props {
    authenticatedUser: Pick<GQL.IUser, 'id' | 'username' | 'avatarURL'>
    history: H.History

    _createCampaign?: typeof createCampaign
}

const LOADING = 'loading' as const

/**
 * A page for creating a new campaign.
 *
 * TODO(sqs): warn before navigating away if that would result in data loss
 */
export const NewCampaignPage: React.FunctionComponent<Props> = ({
    authenticatedUser,
    history,
    _createCampaign = createCampaign,
}) => {
    const [description, setDescription] = useState<string>('')

    const [branch, setBranch] = useState<string>('')
    const [branchModified, setBranchModified] = useState<boolean>(false)
    const onBranchChange = useCallback((newValue: string): void => {
        setBranch(newValue)
        setBranchModified(true)
    }, [])

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

    const [onSubmit, submitOrError] = useEventObservable(
        useCallback(
            (submits: Observable<React.FormEvent>) =>
                submits.pipe(
                    tap(event => event.preventDefault()),
                    switchMap(() =>
                        from(
                            _createCampaign({
                                name,
                                description,
                                namespace: authenticatedUser.id,
                                branch,
                            })
                        ).pipe(
                            catchError(err => of(asError(err))),
                            startWith(LOADING)
                        )
                    )
                ),
            [_createCampaign, authenticatedUser.id, branch, description, name]
        )
    )
    const isLoading = submitOrError === LOADING

    return (
        <>
            <PageTitle title="New campaign" />
            <h1>New campaign</h1>
            {submitOrError !== undefined && submitOrError !== LOADING ? (
                isErrorLike(submitOrError) ? (
                    <ErrorAlert error={submitOrError} history={history} />
                ) : (
                    <Redirect to={submitOrError.url} />
                )
            ) : null}
            <Form onSubmit={onSubmit} className="e2e-new-campaign-form">
                <CampaignTitleField value={name} onChange={onNameChange} disabled={isLoading} />
                <CampaignDescriptionField value={description} onChange={setDescription} disabled={isLoading} />
                <CampaignBranchField value={branch} onChange={onBranchChange} disabled={isLoading} />
                <div className="form-group d-flex align-items-center">
                    <button type="submit" className="btn btn-primary mr-2 e2e-campaign-create-btn" disabled={isLoading}>
                        Create campaign
                    </button>
                    <p className="small text-muted mb-0">Next, you can upload patches and add changesets.</p>
                </div>
            </Form>
        </>
    )
}
