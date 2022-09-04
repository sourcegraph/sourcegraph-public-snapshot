import React, { useEffect } from 'react'

import { GOOGLE_OAUTH_CLIENT_ID } from './authProvider'
import { useScript } from './useScript'

/** @see https://developers.google.com/identity/gsi/web/reference/js-reference#GsiButtonConfiguration */
export interface GsiButtonConfiguration {
    type?: 'standard' | 'icon'
    theme?: 'outline' | 'filled_blue' | 'filled_black'
    size?: 'large' | 'medium' | 'small'
    text?: 'signin_with' | 'signup_with' | 'continue_with' | 'signup_with'
    shape?: 'rectangular' | 'pill' | 'circle' | 'square'
    logo_alignment?: 'left' | 'center'
    locale?: string
    width?: number
}

/** @see https://developers.google.com/identity/gsi/web/reference/js-reference#IdConfiguration */
interface IdConfiguration {
    client_id?: string | undefined
    auto_select?: boolean | undefined
    callback?: ((credentialResponse: CredentialResponse) => void) | undefined
    login_uri?: string
    native_callback?: (() => void) | undefined
    cancel_on_tap_outside?: boolean | undefined
    prompt_parent_id?: string | undefined
    nonce?: string | undefined
    context?: string | undefined
    state_cookie_domain?: string | undefined
    ux_mode?: 'popup' | 'redirect'
    allowed_parent_origin?: string | string[] | undefined
    intermediate_iframe_close_callback?: (() => void) | undefined
}

/** @see https://developers.google.com/identity/gsi/web/reference/js-reference#CredentialResponse */
interface CredentialResponse {
    credential: string
    select_by: string
}

/** @see https://developers.google.com/identity/gsi/web/reference/js-reference#RevocationResponse */
interface RevocationResponse {
    successful: boolean
    error?: string
}

declare global {
    interface Window {
        google?: {
            accounts: {
                id: {
                    /** @see https://developers.google.com/identity/gsi/web/reference/js-reference#google.accounts.id.initialize */
                    initialize(idConfiguration: IdConfiguration): void

                    /** @see https://developers.google.com/identity/gsi/web/reference/js-reference#google.accounts.id.cancel */
                    cancel(): void

                    /** @see https://developers.google.com/identity/gsi/web/reference/js-reference#google.accounts.id.revoke */
                    revoke(hint: string, callback?: RevocationResponse): void
                }
            }
        }
    }
}

export const SignInWithGoogle: React.FunctionComponent<{
    onComplete: (jwt: string) => void
    buttonConfig?: GsiButtonConfiguration
}> = ({ onComplete, buttonConfig }) => {
    useScript(
        'https://accounts.google.com/gsi/client',
        () => {
            if (window.google) {
                window.google.accounts.id.initialize({
                    client_id: GOOGLE_OAUTH_CLIENT_ID,
                    callback: ({ credential: jwt }: CredentialResponse) => {
                        onComplete(jwt)

                        // HACK TODO(sqs): get scopes
                        const client = window.google?.accounts.oauth2.initTokenClient({
                            client_id: GOOGLE_OAUTH_CLIENT_ID,
                            scope: 'https://www.googleapis.com/auth/tasks.readonly',
                            callback: tokenResponse => {
                                // TODO(sqs): blocked by popup blockers
                            },
                        })
                        client.requestAccessToken()
                    },
                })
            }
        },
        error => {
            console.error('Error loading Google Identity API:', error)
        }
    )

    useEffect(
        () => () => {
            if (window.google) {
                window.google?.accounts.id.cancel()
            }
        },
        []
    )

    return (
        <button
            className="g_id_signin"
            data-type={buttonConfig?.type}
            data-size={buttonConfig?.size}
            data-theme={buttonConfig?.theme}
            data-text={buttonConfig?.text}
            data-shape={buttonConfig?.shape}
            data-logo_alignment={buttonConfig?.logo_alignment}
            data-locale={buttonConfig?.locale}
            data-width={buttonConfig?.width}
            style={{
                border: 0,
                outline: 0,
                padding: 0,
                margin: 0,
                backgroundColor: 'transparent',
                colorScheme: 'normal',
                height: '40px',
            }}
        />
    )
}
