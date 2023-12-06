import React, { useCallback, useMemo, useState } from 'react'

import { mdiPlus } from '@mdi/js'
import { useNavigate } from 'react-router-dom'
import { concat, Subject } from 'rxjs'
import { catchError, concatMap, tap } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Container,
    PageHeader,
    LoadingSpinner,
    Button,
    useObservable,
    Link,
    Icon,
    Checkbox,
    Input,
    Text,
    Label,
    ErrorAlert,
    Form,
} from '@sourcegraph/wildcard'

import { AccessTokenScopes } from '../../../auth/accessToken'
import { PageTitle } from '../../../components/PageTitle'
import type { CreateAccessTokenResult } from '../../../graphql-operations'
import type { UserSettingsAreaRouteContext } from '../UserSettingsArea'

import { createAccessToken } from './create'

interface Props extends Pick<UserSettingsAreaRouteContext, 'user'>, TelemetryProps, TelemetryV2Props {
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
    telemetryRecorder,
    onDidCreateAccessToken,
    user,
}) => {
    const navigate = useNavigate()

    useMemo(() => {
        telemetryService.logViewEvent('NewAccessToken')
        telemetryRecorder.recordEvent('NewAccessToken', 'created')
    }, [telemetryService, telemetryRecorder])

    /** The contents of the note input field. */
    const defaultNoteValue = new URLSearchParams(location.search).get('description') || undefined
    const [note, setNote] = useState<string>(defaultNoteValue ?? '')
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
                                    navigate('..', { relative: 'path' })
                                    onDidCreateAccessToken(result)
                                }),
                                catchError(error => [asError(error)])
                            )
                        )
                    )
                ),
            [navigate, note, onDidCreateAccessToken, scopes, submits, user.id]
        )
    )

    return (
        <div className="user-settings-create-access-token-page">
            <PageTitle title="Create access token" />
            <PageHeader path={[{ text: 'New access token' }]} headingElement="h2" className="mb-3" />

            <Form onSubmit={onSubmit}>
                <Container className="mb-3">
                    <Input
                        data-testid="test-create-access-token-description"
                        id="user-settings-create-access-token-page__note"
                        onChange={onNoteChange}
                        required={true}
                        autoFocus={true}
                        placeholder="What's this token for?"
                        defaultValue={defaultNoteValue}
                        className="form-group"
                        label="Token description"
                    />

                    <div className="form-group mb-0">
                        <Label htmlFor="user-settings-create-access-token-page__scope-user:all" className="mb-0">
                            Token scope
                        </Label>
                        <Text>
                            <small className="form-help text-muted">
                                Tokens with limited user scopes are not yet supported.
                            </small>
                        </Text>

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
                        {creationOrError === 'loading' ? (
                            <LoadingSpinner />
                        ) : (
                            <Icon aria-hidden={true} svgPath={mdiPlus} />
                        )}{' '}
                        Generate token
                    </Button>
                    <Button className="ml-2 test-create-access-token-cancel" to=".." variant="secondary" as={Link}>
                        Cancel
                    </Button>
                </div>
            </Form>

            {isErrorLike(creationOrError) && <ErrorAlert className="my-3" error={creationOrError} />}
        </div>
    )
}
