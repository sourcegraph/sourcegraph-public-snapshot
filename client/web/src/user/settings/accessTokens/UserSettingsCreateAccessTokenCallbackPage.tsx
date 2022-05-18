import React, { useEffect, useMemo, useState } from 'react'

import { RouteComponentProps } from 'react-router'
import { NEVER } from 'rxjs'
import { catchError, startWith, tap } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { asError, isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Container,
    PageHeader,
    Button,
    useObservable,
    Link,
    Checkbox,
    LoadingSpinner,
    Alert,
    Typography,
} from '@sourcegraph/wildcard'

import { AccessTokenScopes } from '../../../auth/accessToken'
import { CopyableText } from '../../../components/CopyableText'
import { PageTitle } from '../../../components/PageTitle'
import { CreateAccessTokenResult } from '../../../graphql-operations'
import { UserSettingsAreaRouteContext } from '../UserSettingsArea'

import { createAccessToken } from './create'

interface Props
    extends Pick<UserSettingsAreaRouteContext, 'authenticatedUser' | 'user'>,
        Pick<RouteComponentProps<{}>, 'history' | 'match'>,
        TelemetryProps {
    /**
     * Called when a new access token is created and should be temporarily displayed to the user.
     */
    onDidCreateAccessToken: (value: CreateAccessTokenResult['createAccessToken']) => void
}

interface TokenRequester {
    name: string
    redirectURL: string
    description: string
    message?: string
}

// SECURITY: Only accept callback requests from requesters on this allowed list
const REQUESTERS: Record<string, TokenRequester> = {
    LOGINVSCE: {
        name: 'VS Code Extension',
        redirectURL: 'vscode://sourcegraph.sourcegraph?code=$TOKEN',
        description: 'VS Code Extension for Sourcegraph',
        message: 'Click popup to redirect back to VS Code',
    },
}

/**
 * This page acts as a callback URL after the authentication process has been completed by a user.
 * This can be shared among different SG integrations as long as the value that is being passed in
 * using the 'requestFrom' param (.../user/settings/tokens/new/callback?requestFrom=$SOURCE) is included in
 * the REQUESTERS allow list above.
 * Once the request has been validated, the user will then be redirected back to the source with the newly created token passing
 * in as a new URL param, using the redirect URL associated with the allowlisted requester The token should then be processed by the extension's
 * URL handler (For example, "vscode://sourcegraph/sourcegraph?code=$TOKEN" for the VS Code extension)
 */
export const UserSettingsCreateAccessTokenCallbackPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    telemetryService,
    onDidCreateAccessToken,
    user,
    history,
    match,
}) => {
    useMemo(() => {
        telemetryService.logPageView('NewAccessToken')
    }, [telemetryService])

    /** Get the token description from the url parameters if any */
    const requestFrom = new URLSearchParams(history.location.search).get('requestFrom')
    /** The validated requester where the callback request originally comes from. */
    const [requester, setRequester] = useState<TokenRequester | null | undefined>(undefined)
    /** The contents of the note input field. */
    const [note, setNote] = useState<string>('')
    /** The selected scopes checkboxes. */
    const [scopes, setScopes] = useState<string[]>([AccessTokenScopes.UserAll])
    /** The newly created token if any. */
    const [newToken, setNewToken] = useState('')

    // Check and Match URL Search Params
    useEffect((): void => {
        // SECURITY: If the URL contains ?requestFrom, verify it is an allowlisted source
        if (requestFrom && requester === undefined) {
            const uriPattern = REQUESTERS[requestFrom ?? ''].redirectURL
            setRequester(uriPattern ? REQUESTERS[requestFrom] : null)
            setNote(uriPattern ? REQUESTERS[requestFrom].name : '')
        }
        // Redirect users back to tokens page if none or invalid url params provided
        if (!requestFrom || (!requester && requester !== undefined)) {
            console.error('Error: Cannot process requests from unknown source.')
            history.push(`${match.url.replace(/\/new\/callback$/, '')}`)
        }
    }, [history, match.url, requestFrom, requester])

    /**
     * We use this to handle token creation request from redirections.
     * Don't create token if this page wasn't linked to from a valid
     * requester (e.g. VS Code extension).
     */
    const creationOrError = useObservable(
        useMemo(
            () =>
                (requester ? createAccessToken(user.id, scopes, note) : NEVER).pipe(
                    tap(result => {
                        // SECURITY: If the request was from a valid requestor, redirect to the allowlisted redirect URL.
                        if (requester) {
                            onDidCreateAccessToken(result)
                            setNewToken(result.token)
                            const uri = requester?.redirectURL.replace('$TOKEN', result.token)
                            window.location.replace(uri)
                        }
                    }),
                    startWith('loading'),
                    catchError(error => [asError(error)])
                ),
            [requester, user.id, scopes, note, onDidCreateAccessToken]
        )
    )
    /**
     * If there's a uriPattern but no result or error yet, we can assume that creation is
     * in progress and show a loading spinner + message.
     */
    if (creationOrError === 'loading') {
        return <LoadingSpinner />
    }
    return (
        <div className="user-settings-create-access-token-page">
            <PageTitle title="Create access token" />
            <PageHeader
                path={[{ text: `New access token ${requester ? 'for ' + requester.name : ''}` }]}
                headingElement="h2"
                className="mb-3"
            />
            {newToken && requester && (
                <Form>
                    <Container className="mb-3">
                        <div className="form-group">
                            <label htmlFor="user-settings-create-access-token-page__note">Token description</label>
                            <input
                                type="text"
                                className="form-control test-create-access-token-description"
                                id="user-settings-create-access-token-page__note"
                                placeholder={note}
                                disabled={true}
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
                                onChange={() => {
                                    setScopes([AccessTokenScopes.UserAll])
                                }}
                                disabled={true}
                            />
                        </div>
                        <Alert className="access-token-created-alert mt-3" variant="success">
                            <p>Copy the new access token now. You won't be able to see it again.</p>
                            <CopyableText className="test-access-token" text={newToken} size={48} />
                            <Typography.H5 className="mt-4 mb-2">
                                <strong>{requester?.message}</strong>
                            </Typography.H5>
                        </Alert>
                    </Container>
                    <div className="mb-3">
                        <Button
                            to={requester.redirectURL.replace('$TOKEN', newToken)}
                            disabled={creationOrError === 'loading'}
                            variant="primary"
                            as={Link}
                        >
                            Import token to {requester.name}
                        </Button>
                        <Button
                            className="ml-2"
                            to={match.url.replace(/\/new\/callback$/, '')}
                            disabled={creationOrError === 'loading'}
                            variant="secondary"
                            as={Link}
                        >
                            Back
                        </Button>
                    </div>
                </Form>
            )}
            {isErrorLike(creationOrError) && <ErrorAlert className="my-3" error={creationOrError} />}
        </div>
    )
}
