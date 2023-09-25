import React from 'react'

import { mdiGithub } from '@mdi/js'
import classNames from 'classnames'

import { Link, Icon } from '@sourcegraph/wildcard'

import type { AuthProvider, SourcegraphContext } from '../../jscontext'

import styles from './ExternalsAuth.module.scss'

interface ExternalsAuthProps {
    context: Pick<SourcegraphContext, 'authProviders'>
    githubLabel: string
    gitlabLabel: string
    googleLabel: string
    onClick: (type: AuthProvider['serviceType']) => void
    withCenteredText?: boolean
    ctaClassName?: string
    iconClassName?: string
    redirect?: string
}

const GitlabColorIcon: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({ className }) => (
    <svg
        className={className}
        width="24"
        height="24"
        viewBox="-2 -2 26 26"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
    >
        <path d="M9.99944 19.2025L13.684 7.86902H6.32031L9.99944 19.2025Z" fill="#E24329" />
        <path
            d="M1.1594 7.8689L0.037381 11.3121C-0.0641521 11.6248 0.0454967 11.9699 0.313487 12.1648L9.99935 19.2023L1.1594 7.8689Z"
            fill="#FCA326"
        />
        <path
            d="M1.15918 7.86873H6.31995L4.0989 1.04315C3.98522 0.693949 3.48982 0.693949 3.37206 1.04315L1.15918 7.86873Z"
            fill="#E24329"
        />
        <path
            d="M18.8444 7.8689L19.9624 11.3121C20.0639 11.6248 19.9542 11.9699 19.6862 12.1648L9.99902 19.2023L18.8444 7.8689Z"
            fill="#FCA326"
        />
        <path
            d="M18.8449 7.86873H13.6841L15.901 1.04315C16.0147 0.693949 16.5101 0.693949 16.6279 1.04315L18.8449 7.86873Z"
            fill="#E24329"
        />
        <path d="M9.99902 19.2023L13.6835 7.8689H18.8444L9.99902 19.2023Z" fill="#FC6D26" />
        <path d="M9.99907 19.2023L1.15918 7.8689H6.31995L9.99907 19.2023Z" fill="#FC6D26" />
    </svg>
)

const GoogleIcon: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({ className }) => (
    <svg className={className} xmlns="http://www.w3.org/2000/svg" viewBox="0 0 48 48" width="24px" height="24px">
        <path
            fill="#FFC107"
            d="M43.611,20.083H42V20H24v8h11.303c-1.649,4.657-6.08,8-11.303,8c-6.627,0-12-5.373-12-12c0-6.627,5.373-12,12-12c3.059,0,5.842,1.154,7.961,3.039l5.657-5.657C34.046,6.053,29.268,4,24,4C12.955,4,4,12.955,4,24c0,11.045,8.955,20,20,20c11.045,0,20-8.955,20-20C44,22.659,43.862,21.35,43.611,20.083z"
        />
        <path
            fill="#FF3D00"
            d="M6.306,14.691l6.571,4.819C14.655,15.108,18.961,12,24,12c3.059,0,5.842,1.154,7.961,3.039l5.657-5.657C34.046,6.053,29.268,4,24,4C16.318,4,9.656,8.337,6.306,14.691z"
        />
        <path
            fill="#4CAF50"
            d="M24,44c5.166,0,9.86-1.977,13.409-5.192l-6.19-5.238C29.211,35.091,26.715,36,24,36c-5.202,0-9.619-3.317-11.283-7.946l-6.522,5.025C9.505,39.556,16.227,44,24,44z"
        />
        <path
            fill="#1976D2"
            d="M43.611,20.083H42V20H24v8h11.303c-0.792,2.237-2.231,4.166-4.087,5.571c0.001-0.001,0.002-0.001,0.003-0.002l6.19,5.238C36.971,39.205,44,34,44,24C44,22.659,43.862,21.35,43.611,20.083z"
        />
    </svg>
)

export const ExternalsAuth: React.FunctionComponent<React.PropsWithChildren<ExternalsAuthProps>> = ({
    context,
    githubLabel,
    gitlabLabel,
    googleLabel,
    onClick,
    withCenteredText,
    ctaClassName,
    iconClassName,
    redirect,
}) => {
    // Since this component is only intended for use on Sourcegraph.com, it's OK to hardcode
    // GitHub and GitLab auth providers here as they are the only ones used on Sourcegraph.com.
    // In the future if this page is intended for use in Sourcegraph Sever, this would need to be generalized
    // for other auth providers such SAML, OpenID, Okta, Azure AD, etc.

    const githubProvider = context.authProviders.find(provider =>
        provider.authenticationURL.startsWith('/.auth/github/login?pc=https%3A%2F%2Fgithub.com%2F')
    )
    const gitlabProvider = context.authProviders.find(provider =>
        provider.authenticationURL.startsWith('/.auth/gitlab/login?pc=https%3A%2F%2Fgitlab.com%2F')
    )
    const googleProvider = context.authProviders.find(provider =>
        provider.authenticationURL.startsWith('/.auth/openidconnect/login?pc=google')
    )

    return (
        <>
            {githubProvider && (
                <Link
                    // Use absolute URL to force full-page reload (because the auth routes are
                    // handled by the backend router, not the frontend router).
                    to={
                        `${window.location.origin}${githubProvider.authenticationURL}` +
                        (redirect ? `&redirect=${redirect}` : '')
                    }
                    className={classNames(
                        'text-decoration-none',
                        withCenteredText && 'd-flex justify-content-center',
                        styles.signUpButton,
                        styles.githubButton,
                        ctaClassName
                    )}
                    onClick={() => onClick('github')}
                >
                    <Icon
                        className={classNames('mr-2', iconClassName)}
                        svgPath={mdiGithub}
                        inline={false}
                        aria-hidden={true}
                    />{' '}
                    {githubLabel}
                </Link>
            )}

            {gitlabProvider && (
                <Link
                    // Use absolute URL to force full-page reload (because the auth routes are
                    // handled by the backend router, not the frontend router).
                    to={
                        `${window.location.origin}${gitlabProvider.authenticationURL}` +
                        (redirect ? `&redirect=${redirect}` : '')
                    }
                    className={classNames(
                        'text-decoration-none',
                        withCenteredText && 'd-flex justify-content-center',
                        styles.signUpButton,
                        styles.gitlabButton,
                        ctaClassName
                    )}
                    onClick={() => onClick('gitlab')}
                >
                    <GitlabColorIcon className={classNames('mr-2', iconClassName)} /> {gitlabLabel}
                </Link>
            )}

            {googleProvider && (
                <Link
                    // Use absolute URL to force full-page reload (because the auth routes are
                    // handled by the backend router, not the frontend router).
                    to={
                        `${window.location.origin}${googleProvider.authenticationURL}` +
                        (redirect ? `&redirect=${redirect}` : '')
                    }
                    className={classNames(
                        'text-decoration-none',
                        withCenteredText && 'd-flex justify-content-center',
                        styles.signUpButton,
                        styles.gitlabButton,
                        ctaClassName
                    )}
                    onClick={() => onClick('openidconnect')}
                >
                    <GoogleIcon className={classNames('mr-2', iconClassName)} /> {googleLabel}
                </Link>
            )}
        </>
    )
}
