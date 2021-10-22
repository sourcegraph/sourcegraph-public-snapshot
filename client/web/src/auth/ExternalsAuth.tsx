import classNames from 'classnames'
import GithubIcon from 'mdi-react/GithubIcon'
import React from 'react'

import { AuthProvider, SourcegraphContext } from '../jscontext'

import styles from './CloudSignUpPage.module.scss'
import { maybeAddPostSignUpRedirect } from './SignInSignUpCommon'

interface ExternalsAuthProps {
    context: Pick<SourcegraphContext, 'authProviders' | 'experimentalFeatures'>
    githubLabel: string
    gitlabLabel: string
    onClick: (type: AuthProvider['serviceType']) => void
    withCenteredText?: boolean
}

const GitlabColorIcon: React.FunctionComponent<{ className?: string }> = ({ className }) => (
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

const ExternalsAuth: React.FunctionComponent<ExternalsAuthProps> = ({
    context,
    githubLabel,
    gitlabLabel,
    onClick,
    withCenteredText,
}) => {
    // Since this component is only intended for use on Sourcegraph.com, it's OK to hardcode
    // GitHub and GitLab auth providers here as they are the only ones used on Sourcegraph.com.
    // In the future if this page is intended for use in Sourcegraph Sever, this would need to be generalized
    // for other auth providers such SAML, OpenID, Okta, Azure AD, etc.

    const githubProvider = context.authProviders.find(provider =>
        provider.authenticationURL?.startsWith('/.auth/github/login?pc=https%3A%2F%2Fgithub.com%2F')
    )
    const gitlabProvider = context.authProviders.find(provider =>
        provider.authenticationURL?.startsWith('/.auth/gitlab/login?pc=https%3A%2F%2Fgitlab.com%2F')
    )

    return (
        <>
            {githubProvider && (
                <a
                    href={maybeAddPostSignUpRedirect(githubProvider.authenticationURL)}
                    className={classNames(
                        'text-decoration-none',
                        withCenteredText && 'd-flex justify-content-center',
                        styles.signUpButton,
                        styles.githubButton
                    )}
                    onClick={() => onClick('github')}
                >
                    <GithubIcon className="mr-3" /> {githubLabel}
                </a>
            )}

            {gitlabProvider && (
                <a
                    href={maybeAddPostSignUpRedirect(gitlabProvider.authenticationURL)}
                    className={classNames(
                        'text-decoration-none',
                        withCenteredText && 'd-flex justify-content-center',
                        styles.signUpButton,
                        styles.gitlabButton
                    )}
                    onClick={() => onClick('gitlab')}
                >
                    <GitlabColorIcon className="mr-3" /> {gitlabLabel}
                </a>
            )}
        </>
    )
}

export { ExternalsAuth }
