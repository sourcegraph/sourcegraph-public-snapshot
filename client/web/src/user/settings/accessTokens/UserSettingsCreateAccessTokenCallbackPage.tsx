import React, { useCallback, useEffect, useMemo, useState } from 'react'

import { useLocation, useNavigate } from 'react-router-dom'
import { NEVER, type Observable } from 'rxjs'
import { catchError, startWith, switchMap, tap } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { Button, Link, Text, ErrorAlert, Card, H1, H2, useEventObservable } from '@sourcegraph/wildcard'

import { tauriShellOpen } from '../../../app/tauriIcpUtils'
import { AccessTokenScopes } from '../../../auth/accessToken'
import { BrandLogo } from '../../../components/branding/BrandLogo'
import { CopyableText } from '../../../components/CopyableText'
import { LoaderButton } from '../../../components/LoaderButton'
import type { CreateAccessTokenResult } from '../../../graphql-operations'
import type { UserSettingsAreaRouteContext } from '../UserSettingsArea'

import { createAccessToken } from './create'

import styles from './UserSettingsCreateAccessTokenCallbackPage.module.scss'

interface Props
    extends Pick<UserSettingsAreaRouteContext, 'authenticatedUser' | 'user'>,
        TelemetryProps,
        TelemetryV2Props {
    /**
     * Called when a new access token is created and should be temporarily displayed to the user.
     */
    onDidCreateAccessToken: (value: CreateAccessTokenResult['createAccessToken']) => void
    isSourcegraphDotCom: boolean
    isCodyApp: boolean
}
interface TokenRequester {
    /** The name of the source */
    name: string
    /** The URL where the token should be added to */
    /** SECURITY: Local context only! Do not send token to non-local servers*/
    redirectURL: string
    /** The message to show users when the token has been created successfully */
    successMessage?: string
    /** The message to show users in case the token cannot be imported automatically */
    infoMessage?: string
    /** How the redirect URL should be open: open in same tab vs open in a new-tab */
    /** Default: Open link in same tab */
    callbackType?: 'open' | 'new-tab'
    /** If set, the requester is only allowed on dotcom */
    onlyDotCom?: boolean
    /** If true, it will forward the `destination` param to the redirect URL if it starts with / */
    forwardDestination?: boolean
}
// SECURITY: Only accept callback requests from requesters on this allowed list
const REQUESTERS: Record<string, TokenRequester> = {
    VSCEAUTH: {
        name: 'VS Code Extension',
        redirectURL: 'vscode://sourcegraph.sourcegraph?code=$TOKEN',
        successMessage: 'Now opening VS Code...',
        infoMessage:
            'Please make sure you have VS Code running on your machine if you do not see an open dialog in your browser.',
        callbackType: 'new-tab',
    },
    APP: {
        name: 'Cody App',
        redirectURL: 'sourcegraph://app/auth/callback?code=$TOKEN',
        successMessage: 'Now opening the Cody App...',
        infoMessage: 'You will be redirected to Cody App.',
        callbackType: 'open',
        onlyDotCom: true,
        forwardDestination: true,
    },
    CODY: {
        name: 'Cody AI by Sourcegraph - VS Code Extension',
        redirectURL: 'vscode://sourcegraph.cody-ai?code=$TOKEN',
        successMessage: 'Now opening VS Code...',
        infoMessage:
            'Please make sure you have VS Code running on your machine if you do not see an open dialog in your browser.',
        callbackType: 'new-tab',
    },
    CODY_INSIDERS: {
        name: 'Cody AI by Sourcegraph - VS Code Insiders Extension',
        redirectURL: 'vscode-insiders://sourcegraph.cody-ai?code=$TOKEN',
        successMessage: 'Now opening VS Code...',
        infoMessage:
            'Please make sure you have VS Code running on your machine if you do not see an open dialog in your browser.',
        callbackType: 'new-tab',
    },
    JETBRAINS: {
        name: 'JetBrains IDE',
        redirectURL: 'http://localhost:$PORT/api/sourcegraph/token?token=$TOKEN',
        successMessage: 'Now opening your IDE...',
        infoMessage:
            'Please make sure you still have your IDE (IntelliJ, GoLand, PyCharm, etc.) running on your machine when clicking this link.',
        callbackType: 'open',
    },
}

export function isAccessTokenCallbackPage(): boolean {
    return location.pathname.endsWith('/settings/tokens/new/callback')
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
export const UserSettingsCreateAccessTokenCallbackPage: React.FC<Props> = ({
    telemetryService,
    telemetryRecorder,
    onDidCreateAccessToken,
    user,
    isSourcegraphDotCom,
    isCodyApp,
}) => {
    const isLightTheme = useIsLightTheme()
    const navigate = useNavigate()
    const location = useLocation()
    useEffect(() => {
        telemetryService.logPageView('NewAccessTokenCallback')
        telemetryRecorder.recordEvent('NewAccessTokenCallback', 'completed')
    }, [telemetryService, telemetryRecorder])

    /** Get the requester, port, and destination from the url parameters */
    const urlSearchParams = useMemo(() => new URLSearchParams(location.search), [location.search])
    const requestFrom = useMemo(() => urlSearchParams.get('requestFrom'), [urlSearchParams])
    const port = useMemo(() => urlSearchParams.get('port'), [urlSearchParams])
    const destination = useMemo(() => urlSearchParams.get('destination'), [urlSearchParams])

    /** The validated requester where the callback request originally comes from. */
    const [requester, setRequester] = useState<TokenRequester | null | undefined>(undefined)
    /** The contents of the note input field. */
    const [note, setNote] = useState<string>('')
    /** The newly created token if any. */
    const [newToken, setNewToken] = useState('')

    // Check and Match URL Search Prams
    useEffect((): void => {
        // If a requester is already set, we don't need to run this effect
        if (requester) {
            return
        }

        // SECURITY: Verify if the request is coming from an allowlisted source
        const isRequestValid = requestFrom && requestFrom in REQUESTERS
        if (!isRequestValid || !requestFrom || requester !== undefined) {
            navigate('../..', { relative: 'path' })
            return
        }

        if (REQUESTERS[requestFrom].onlyDotCom && !isSourcegraphDotCom) {
            navigate('../..', { relative: 'path' })
            return
        }

        // SECURITY: If the request is coming from JetBrains, verify if the port is valid
        if (requestFrom === 'JETBRAINS' && (!port || !Number.isInteger(Number(port)))) {
            navigate('../..', { relative: 'path' })
            return
        }

        const nextRequester = { ...REQUESTERS[requestFrom] }

        if (nextRequester.forwardDestination) {
            // SECURITY: only destinations starting with a "/" are allowed to prevent an open redirect vulnerability.
            if (destination?.startsWith('/')) {
                const redirectURL = new URL(nextRequester.redirectURL)
                redirectURL.searchParams.set('destination', destination)
                nextRequester.redirectURL = redirectURL.toString()
            }
        }

        if (isCodyApp) {
            // Append type=app to the url to indicate to the requester that the callback is fulfilled by App
            const redirectURL = new URL(nextRequester.redirectURL)
            redirectURL.searchParams.set('type', 'app')
            nextRequester.redirectURL = redirectURL.toString()
        }

        setRequester(nextRequester)
        setNote(REQUESTERS[requestFrom].name)
    }, [isSourcegraphDotCom, isCodyApp, location.search, navigate, requestFrom, requester, port, destination])

    /**
     * We use this to handle token creation request from redirections.
     * Don't create token if this page wasn't linked to from a valid
     * requester (e.g. VS Code extension).
     */
    const [onAuthorize, creationOrError] = useEventObservable(
        useCallback(
            (click: Observable<React.MouseEvent>) =>
                click.pipe(
                    switchMap(() =>
                        (requester ? createAccessToken(user.id, [AccessTokenScopes.UserAll], note) : NEVER).pipe(
                            tap(result => {
                                // SECURITY: If the request was from a valid requester, redirect to the allowlisted redirect URL.
                                // SECURITY: Local context ONLY
                                if (requester) {
                                    onDidCreateAccessToken(result)
                                    setNewToken(result.token)
                                    let uri = replacePlaceholder(requester?.redirectURL, 'TOKEN', result.token)
                                    if (requestFrom === 'JETBRAINS' && port) {
                                        uri = replacePlaceholder(uri, 'PORT', port)
                                    }

                                    // If we're in App, override the callbackType
                                    // because we need to use tauriShellOpen to open the
                                    // callback in a browser.
                                    // Then navigate back to the home page since App doesn't
                                    // have a back button or tab that can be closed.
                                    if (isCodyApp) {
                                        tauriShellOpen(uri)
                                        navigate('/')
                                        return
                                    }

                                    switch (requester.callbackType) {
                                        case 'new-tab':
                                            window.open(uri, '_blank')

                                        // falls through
                                        default:
                                            // open the redirect link in the same tab
                                            window.location.replace(uri)
                                    }
                                }
                            }),
                            startWith('loading'),
                            catchError(error => [asError(error)])
                        )
                    )
                ),
            [requester, user.id, note, onDidCreateAccessToken, requestFrom, port, isCodyApp, navigate]
        )
    )

    if (!requester) {
        return null
    }

    return (
        <div className={styles.wrapper}>
            <BrandLogo className={styles.logo} isLightTheme={isLightTheme} variant="logo" />

            <Card className={styles.card}>
                <H2 as={H1} className={styles.heading}>
                    Authorize {requester.name}?
                </H2>

                <Text weight="bold">This grants access to:</Text>
                <ul>
                    <li>Your Sourcegraph.com account</li>
                    <li>Perform actions on your behalf</li>
                </ul>
                <Text>{requester.infoMessage}</Text>
                <Text>If you are not trying to connect {requester.name}, click cancel.</Text>

                {!newToken && (
                    <div className={styles.buttonRow}>
                        <LoaderButton
                            className="flex-1"
                            variant="primary"
                            label="Authorize"
                            loading={creationOrError === 'loading'}
                            onClick={onAuthorize}
                        />
                        <Button
                            className="flex-1"
                            variant="secondary"
                            to={location.pathname.replace(/\/new\/callback$/, '')}
                            disabled={creationOrError === 'loading'}
                            as={Link}
                        >
                            Cancel
                        </Button>
                    </div>
                )}

                {newToken && (
                    <>
                        <Text weight="bold">{requester.successMessage}</Text>
                        <details>
                            <summary>Authorization details</summary>
                            <div className="mt-2">
                                <Text>{requester.name} access token successfully generated.</Text>
                                <CopyableText className="test-access-token" text={newToken} />
                                <Text className="form-help text-muted" size="small">
                                    This is a one-time access token to connect your account to {requester.name}. You
                                    will not be able to see this token again once the window is closed.
                                </Text>
                            </div>
                        </details>
                    </>
                )}

                {isErrorLike(creationOrError) && <ErrorAlert className="my-3 " error={creationOrError} />}
            </Card>
        </div>
    )
}

function replacePlaceholder(subject: string, search: string, replace: string): string {
    // %24 is the URL encoded version of $
    return subject.replace('$' + search, replace).replace('%24' + search, replace)
}
