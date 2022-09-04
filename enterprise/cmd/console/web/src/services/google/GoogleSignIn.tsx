import { decodeJwt, JWTPayload } from 'jose'
import React, { useCallback } from 'react'

import { SettingsProps } from '../../app/useSettings'
import { GsiButtonConfiguration, SignInWithGoogle } from './SignInWithGoogle'

export interface GoogleAuthSettings {
    jwt: string
    jwtPayload: GoogleIdentityJWTPayload
}

/** @see https://developers.google.com/identity/gsi/web/reference/js-reference#credential */
interface GoogleIdentityJWTPayload extends Required<Pick<JWTPayload, 'iss' | 'sub' | 'nbf' | 'exp' | 'iat' | 'jti'>> {
    aud: string
    hd?: string
    email: string
    email_verified: boolean
    name: string
    picture?: string
    given_name: string
    family_name: string
}

const BUTTON_CONFIG: GsiButtonConfiguration = {
    theme: 'filled_blue',
    type: 'standard',
    size: 'large',
}

export const GoogleSignIn: React.FunctionComponent<SettingsProps> = ({ settings: { googleAuth }, setSettings }) => {
    const onSignIn = useCallback(
        (jwt: string) => {
            setSettings(settings => ({
                ...settings,
                googleAuth: {
                    jwt,
                    jwtPayload: decodeJwt(jwt) as unknown as GoogleIdentityJWTPayload,
                },
            }))
        },
        [setSettings]
    )

    const onSignOut = useCallback(() => {
        setSettings(settings => ({ ...settings, googleAuth: undefined }))
        if (window.google && googleAuth) {
            window.google.accounts.id.revoke(googleAuth.jwtPayload.sub)
        }
    }, [googleAuth, setSettings])

    return googleAuth ? (
        <>
            Google ({googleAuth.jwtPayload.email}){' '}
            <button type="button" onClick={onSignOut}>
                Sign out
            </button>
        </>
    ) : (
        <SignInWithGoogle onComplete={onSignIn} buttonConfig={BUTTON_CONFIG} />
    )
}
