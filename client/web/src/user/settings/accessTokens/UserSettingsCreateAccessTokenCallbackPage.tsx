import React, { useCallback, useEffect, useMemo, useState } from 'react'

import { useLocation, useNavigate } from 'react-router-dom'
import { NEVER, type Observable } from 'rxjs'
import { catchError, startWith, switchMap, tap } from 'rxjs/operators'

import { asError, isErrorLike, isMobile, pluralize } from '@sourcegraph/common'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { Button, Card, ErrorAlert, H1, H2, Link, Text, useEventObservable } from '@sourcegraph/wildcard'

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
    /** If true, the `requestFrom` param will be used to determine the localhost port to redirect to */
    hasRedirectPort?: boolean
    /** If set, the requester is only allowed on dotcom */
    onlyDotCom?: boolean
    /** If true, it will forward the `destination` param to the redirect URL if it starts with / */
    forwardDestination?: boolean
    /**
     * If true, will attempt to POST the newly created auth token to the URL supplied via the
     * tokenReceiverUrl param. This enables the clients to continue the authentication process
     * even if the redirect is not working as expected.
     */
    postTokenToReceiverUrl?: boolean
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
    CODY: {
        name: 'Cody - VS Code Extension',
        redirectURL: 'vscode://sourcegraph.cody-ai?code=$TOKEN',
        successMessage: 'Now opening VS Code...',
        infoMessage:
            'Please make sure you have VS Code running on your machine if you do not see an open dialog in your browser.',
        callbackType: 'new-tab',
        postTokenToReceiverUrl: true,
    },
    CODY_VSCODIUM: {
        name: 'Cody - VSCodium Extension',
        redirectURL: 'vscodium://sourcegraph.cody-ai?code=$TOKEN',
        successMessage: 'Now opening VS Code...',
        infoMessage:
            'Please make sure you have VS Code running on your machine if you do not see an open dialog in your browser.',
        callbackType: 'new-tab',
        postTokenToReceiverUrl: true,
    },
    CODY_CURSOR: {
        name: 'Cody - Cursor Extension',
        redirectURL: 'cursor://sourcegraph.cody-ai?code=$TOKEN',
        successMessage: 'Now opening Cursor...',
        infoMessage:
            'Please make sure you have Cursor running on your machine if you do not see an open dialog in your browser.',
        callbackType: 'new-tab',
    },
    CODY_INSIDERS: {
        name: 'Cody - VS Code Insiders Extension',
        redirectURL: 'vscode-insiders://sourcegraph.cody-ai?code=$TOKEN',
        successMessage: 'Now opening VS Code...',
        infoMessage:
            'Please make sure you have VS Code running on your machine if you do not see an open dialog in your browser.',
        callbackType: 'new-tab',
        postTokenToReceiverUrl: true,
    },
    JETBRAINS: {
        name: 'JetBrains IDE',
        redirectURL: 'http://localhost:$PORT/api/sourcegraph/token?token=$TOKEN',
        successMessage: 'Now opening your IDE...',
        infoMessage:
            'Please make sure you still have your IDE (IntelliJ, GoLand, PyCharm, etc.) running on your machine when clicking this link.',
        callbackType: 'open',
        hasRedirectPort: true,
    },
    ECLIPSE: {
        name: 'Eclipse',
        redirectURL: 'http://localhost:$PORT/api/sourcegraph/token?token=$TOKEN',
        successMessage: 'Now opening Eclipse...',
        infoMessage:
            'Please make sure you still have your Eclipse IDE running on your machine when clicking this link.',
        callbackType: 'open',
        hasRedirectPort: true,
    },
    VISUAL_STUDIO: {
        name: 'Visual Studio',
        redirectURL: 'http://localhost:$PORT/api/sourcegraph/token?token=$TOKEN',
        successMessage: 'Now opening Visual Studio...',
        infoMessage: 'Please make sure you still have Visual Studio running on your machine when clicking this link.',
        callbackType: 'open',
        hasRedirectPort: true,
    },
    CODY_CLI: {
        name: 'Cody CLI',
        redirectURL: 'http://localhost:$PORT/api/sourcegraph/token?token=$TOKEN',
        successMessage: 'Now opening your terminal...',
        infoMessage: 'Please make sure that you are running `cody auth login --web` in a background terminal session.',
        callbackType: 'open',
        hasRedirectPort: true,
    },
    NEOVIM: {
        name: 'Neovim',
        redirectURL: 'http://localhost:$PORT/api/sourcegraph/token?token=$TOKEN',
        successMessage: 'Restart Neovim and your credentials will be saved.',
        infoMessage: 'Please make sure you still have Neovim running on your machine when clicking this link.',
        callbackType: 'open',
        hasRedirectPort: true,
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
}) => {
    const isLightTheme = useIsLightTheme()
    const navigate = useNavigate()
    const location = useLocation()
    const defaultAccessTokenExpiryDays = window.context.accessTokensExpirationDaysDefault
    useEffect(() => {
        telemetryService.logPageView('NewAccessTokenCallback')
        telemetryRecorder.recordEvent('settings.tokens.newAccessToken', 'callback')
    }, [telemetryService, telemetryRecorder])
    /** Get the requester, port, and destination from the url parameters */
    const urlSearchParams = useMemo(() => new URLSearchParams(location.search), [location.search])
    let requestFrom = useMemo(() => urlSearchParams.get('requestFrom'), [urlSearchParams])
    const tokenReceiverUrl = useMemo(() => urlSearchParams.get('tokenReceiverUrl'), [urlSearchParams])
    let port = useMemo(() => urlSearchParams.get('port'), [urlSearchParams])

    // Allow a single query parameter `requestFrom=JETBRAIN-PORT_NUMBER`. The motivation for this parameter encoding is that
    // the separate `port=NUMBER` parameter is lost when we try to log in via GitHub directly from the JetBrains IDE. By encoding the
    // port number inside requestFrom, we have a single query parameter just like with VS Code.
    if (requestFrom?.includes('-')) {
        const [requestFrom1, port1, ...rest] = requestFrom?.split('-')
        if (
            requestFrom1 &&
            port1 &&
            rest.length === 0 &&
            port1.match(/^\d/) &&
            REQUESTERS[requestFrom1]?.hasRedirectPort
        ) {
            requestFrom = requestFrom1
            port = port1
        }
    }

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

        // SECURITY: If the request is coming from a redirectable source
        // (for example, JetBrains or CLI), verify if the port is valid.
        if (REQUESTERS[requestFrom].hasRedirectPort && (!port || !Number.isInteger(Number(port)))) {
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

        setRequester(nextRequester)
        setNote(REQUESTERS[requestFrom].name)
    }, [isSourcegraphDotCom, location.search, navigate, requestFrom, requester, port, destination])

    const isRequestFromMobileDevice = isMobile()

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
                        (requester
                            ? createAccessToken({
                                  user: user.id,
                                  scopes: [AccessTokenScopes.UserAll],
                                  note,
                                  durationSeconds: defaultAccessTokenExpiryDays * 86400, // days to seconds
                                  telemetryRecorder,
                              })
                            : NEVER
                        ).pipe(
                            tap(async result => {
                                // SECURITY: If the request was from a valid requester and from a non-mobile device,
                                // redirect to the allowlisted redirect URL. (https://github.com/sourcegraph/security-issues/issues/361)
                                // SECURITY: Local context ONLY
                                if (requester && !isRequestFromMobileDevice) {
                                    onDidCreateAccessToken(result)
                                    setNewToken(result.token)
                                    let uri = replacePlaceholder(requester?.redirectURL, 'TOKEN', result.token)
                                    if (requestFrom && REQUESTERS[requestFrom].hasRedirectPort && port) {
                                        uri = replacePlaceholder(uri, 'PORT', port)
                                    }

                                    if (requester.postTokenToReceiverUrl && tokenReceiverUrl) {
                                        await postTokenToReceiverUrl(tokenReceiverUrl, result.token)
                                    }

                                    switch (requester.callbackType) {
                                        case 'new-tab': {
                                            window.open(uri, '_blank')
                                        }

                                        // falls through
                                        default: {
                                            // open the redirect link in the same tab
                                            window.location.replace(uri)
                                        }
                                    }
                                }
                            }),
                            startWith('loading'),
                            catchError(error => [asError(error)])
                        )
                    )
                ),
            [
                requester,
                user.id,
                note,
                defaultAccessTokenExpiryDays,
                isRequestFromMobileDevice,
                onDidCreateAccessToken,
                requestFrom,
                port,
                tokenReceiverUrl,
                telemetryRecorder,
            ]
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
                            // we disable this if the request is made from a mobile device so the access token doesn't
                            // get created at all. This prevents redirecting to an external site from a mobile app.
                            disabled={isRequestFromMobileDevice}
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
                                    This is an access token to connect your account to {requester.name}. This token will
                                    expire in {defaultAccessTokenExpiryDays}{' '}
                                    {pluralize('day', defaultAccessTokenExpiryDays)}. You will not be able to see this
                                    token again once the window is closed.
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

async function postTokenToReceiverUrl(tokenReceiverUrl: string, accessToken: string): Promise<void> {
    const url = new URL(tokenReceiverUrl)

    // Do not post data on anything but the local computer
    if (url.hostname !== '127.0.0.1') {
        return
    }

    try {
        await fetch(tokenReceiverUrl, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ accessToken }),
        })
    } catch {
        // Ignore eventual errors
    }
}
