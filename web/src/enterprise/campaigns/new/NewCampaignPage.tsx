import React, { useState, useCallback } from 'react'
import H from 'history'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PageTitle } from '../../../components/PageTitle'
import { isErrorLike, asError } from '../../../../../shared/src/util/errors'
import { ErrorAlert } from '../../../components/alerts'
import { useEventObservable } from '../../../../../shared/src/util/useObservable'
import { startWith, switchMap, catchError, tap } from 'rxjs/operators'
import { from, Observable, of } from 'rxjs'
import { Redirect } from 'react-router'
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
    const [name, setName] = useState<string>('')

    const [onSubmit, submitOrError] = useEventObservable(
        useCallback(
            (submits: Observable<React.FormEvent>) =>
                submits.pipe(
                    tap(event => event.preventDefault()),
                    switchMap(() =>
                        from(
                            _createCampaign({
                                name,
                                namespace: authenticatedUser.id,
                            })
                        ).pipe(
                            catchError(error => of(asError(error))),
                            startWith(LOADING)
                        )
                    )
                ),
            [_createCampaign, authenticatedUser.id, name]
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
                TODO(sqs): form for namespace and name
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
