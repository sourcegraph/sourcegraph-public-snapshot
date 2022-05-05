import React, { useCallback, useMemo, useState } from 'react'

import AddIcon from 'mdi-react/AddIcon'
import { RouteComponentProps } from 'react-router'
import { concat, Observable, Subject } from 'rxjs'
import { catchError, concatMap, map, tap } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { asError, createAggregateError, isErrorLike } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Container,
    PageHeader,
    LoadingSpinner,
    Button,
    useObservable,
    Link,
    Icon,
    Checkbox,
} from '@sourcegraph/wildcard'

import { AccessTokenScopes } from '../../../auth/accessToken'
import { requestGraphQL } from '../../../backend/graphql'
import { PageTitle } from '../../../components/PageTitle'
import { CreateAccessTokenResult, CreateAccessTokenVariables, Scalars } from '../../../graphql-operations'
import { SiteAdminAlert } from '../../../site-admin/SiteAdminAlert'
import { eventLogger } from '../../../tracking/eventLogger'
import { UserSettingsAreaRouteContext } from '../UserSettingsArea'

function createAccessToken(
    user: Scalars['ID'],
    scopes: string[],
    note: string
): Observable<CreateAccessTokenResult['createAccessToken']> {
    return requestGraphQL<CreateAccessTokenResult, CreateAccessTokenVariables>(
        gql`
            mutation CreateAccessToken($user: ID!, $scopes: [String!]!, $note: String!) {
                createAccessToken(user: $user, scopes: $scopes, note: $note) {
                    id
                    token
                }
            }
        `,
        { user, scopes, note }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.createAccessToken || (errors && errors.length > 0)) {
                eventLogger.log('CreateAccessTokenFailed')
                throw createAggregateError(errors)
            }
            eventLogger.log('AccessTokenCreated')
            return data.createAccessToken
        })
    )
}

interface Props
    extends Pick<UserSettingsAreaRouteContext, 'authenticatedUser' | 'user'>,
        Pick<RouteComponentProps<{}>, 'history' | 'match'>,
        TelemetryProps {
    /**
     * Called when a new access token is created and should be temporarily displayed to the user.
     */
    onDidCreateAccessToken: (value: CreateAccessTokenResult['createAccessToken']) => void
}

/**
 * A page with a form to create an access token for a user.
 */
export const UserSettingsCreateAccessTokenPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    telemetryService,
    onDidCreateAccessToken,
    authenticatedUser,
    user,
    history,
    match,
}) => {
    useMemo(() => {
        telemetryService.logViewEvent('NewAccessToken')
    }, [telemetryService])

    /** The contents of the note input field. */
    const [note, setNote] = useState<string>('')
    /** The selected scopes checkboxes. */
    const [scopes, setScopes] = useState<string[]>([AccessTokenScopes.UserAll])

    const onNoteChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setNote(event.currentTarget.value)
    }, [])

    const onScopesChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        const checked = event.currentTarget.checked
        const value = event.currentTarget.value
        setScopes(previous => (checked ? [...previous, value] : previous.filter(scope => scope !== value)))
    }, [])

    const submits = useMemo(() => new Subject<React.FormEvent<HTMLFormElement>>(), [])
    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(event => submits.next(event), [submits])

    const creationOrError = useObservable(
        useMemo(
            () =>
                submits.pipe(
                    tap(event => event.preventDefault()),
                    concatMap(() =>
                        concat(
                            ['loading'],
                            createAccessToken(user.id, scopes, note).pipe(
                                tap(result => {
                                    // Go back to access tokens list page and display the token secret value.
                                    history.push(`${match.url.replace(/\/new$/, '')}`)
                                    onDidCreateAccessToken(result)
                                }),
                                catchError(error => [asError(error)])
                            )
                        )
                    )
                ),
            [history, match.url, note, onDidCreateAccessToken, scopes, submits, user.id]
        )
    )

    const siteAdminViewingOtherUser = authenticatedUser && authenticatedUser.id !== user.id

    return (
        <div className="user-settings-create-access-token-page">
            <PageTitle title="Create access token" />
            <PageHeader path={[{ text: 'New access token' }]} headingElement="h2" className="mb-3" />

            {siteAdminViewingOtherUser && (
                <SiteAdminAlert className="sidebar__alert">
                    Creating access token for other user <strong>{user.username}</strong>
                </SiteAdminAlert>
            )}

            <Form onSubmit={onSubmit}>
                <Container className="mb-3">
                    <div className="form-group">
                        <label htmlFor="user-settings-create-access-token-page__note">Token description</label>
                        <input
                            type="text"
                            className="form-control test-create-access-token-description"
                            id="user-settings-create-access-token-page__note"
                            onChange={onNoteChange}
                            required={true}
                            autoFocus={true}
                            placeholder="What's this token for?"
                        />
                    </div>
                    <div className="form-group mb-0">
                        <label htmlFor="user-settings-create-access-token-page__scope-user:all" className="mb-0">
                            Token scope
                        </label>
                        <p>
                            <small className="form-help text-muted">
                                Tokens with limited user scopes are not yet supported.
                            </small>
                        </p>

                        <Checkbox
                            id="user-settings-create-access-token-page__scope-user:all"
                            checked={true}
                            label={
                                <>
                                    <strong>{AccessTokenScopes.UserAll}</strong> — Full control of all resources
                                    accessible to the user account
                                </>
                            }
                            value={AccessTokenScopes.UserAll}
                            onChange={onScopesChange}
                            disabled={true}
                        />
                        {user.siteAdmin && !window.context.sourcegraphDotComMode && (
                            <Checkbox
                                wrapperClassName="mt-2"
                                id="user-settings-create-access-token-page__scope-site-admin:sudo"
                                checked={scopes.includes(AccessTokenScopes.SiteAdminSudo)}
                                value={AccessTokenScopes.SiteAdminSudo}
                                onChange={onScopesChange}
                                label={
                                    <>
                                        <strong>{AccessTokenScopes.SiteAdminSudo}</strong> — Ability to perform any
                                        action as any other user
                                    </>
                                }
                            />
                        )}
                    </div>
                </Container>
                <div className="mb-3">
                    <Button
                        type="submit"
                        disabled={creationOrError === 'loading'}
                        className="test-create-access-token-submit"
                        variant="primary"
                    >
                        {creationOrError === 'loading' ? <LoadingSpinner /> : <Icon as={AddIcon} />} Generate token
                    </Button>
                    <Button
                        className="ml-2 test-create-access-token-cancel"
                        to={match.url.replace(/\/new$/, '')}
                        variant="secondary"
                        as={Link}
                    >
                        Cancel
                    </Button>
                </div>
            </Form>

            {isErrorLike(creationOrError) && <ErrorAlert className="my-3" error={creationOrError} />}
        </div>
    )
}
