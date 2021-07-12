import classNames from 'classnames'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import GithubIcon from 'mdi-react/GithubIcon'
import React from 'react'
import { Link, useLocation } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { BrandLogo } from '../components/branding/BrandLogo'
import { SourcegraphContext } from '../jscontext'

import styles from './ExperimentalSignUpPage.module.scss'
import { SignUpArguments, SignUpForm } from './SignUpForm'

interface Props extends ThemeProps, TelemetryProps {
    source: string | null
    showEmailForm: boolean
    /** Called to perform the signup on the server. */
    onSignUp: (args: SignUpArguments) => Promise<void>
    context: Pick<SourcegraphContext, 'authProviders'>
}

const SourceToTitleMap = {
    Context: 'Easily search the code you care about.',
    Saved: 'Create a library of useful searches.',
    Monitor: 'Monitor code for changes.',
    Extend: 'Augment code and workflows via extensions.',
    SearchCTA: 'Add your public (and soon private) repositories.',
    Snippet: 'Easily search the code you care about.',
}

export type ExperimentalSignUpSource = keyof typeof SourceToTitleMap

export const ShowEmailFormQueryParameter = 'showEmail'

export const ExperimentalSignUpPage: React.FunctionComponent<Props> = ({
    isLightTheme,
    source,
    showEmailForm,
    onSignUp,
    context,
    telemetryService,
}) => {
    const location = useLocation()

    const queryWithUseEmailToggled = new URLSearchParams(location.search)
    if (showEmailForm) {
        queryWithUseEmailToggled.delete(ShowEmailFormQueryParameter)
    } else {
        queryWithUseEmailToggled.append(ShowEmailFormQueryParameter, 'true')
    }

    const assetsRoot = window.context?.assetsRoot || ''

    // Since this page is only intented for use on Sourcegraph.com, it's OK to hardcode
    // GitHub and GitLab auth providers here as they are the only ones used on Sourcegraph.com.
    // In the future if this page is intented for use in Sourcegraph Sever, this would need to be generisized
    // for other auth providers such SAML, OpenID, Okta, Azure AD, etc.
    const githubProvider = context.authProviders.find(provider =>
        provider.authenticationURL?.startsWith('/.auth/github/login?pc=https%3A%2F%2Fgithub.com%2F')
    )
    const gitlabProvider = context.authProviders.find(provider =>
        provider.authenticationURL?.startsWith('/.auth/gitlab/login?pc=https%3A%2F%2Fgitlab.com%2F')
    )

    const sourceIsValid = source && Object.keys(SourceToTitleMap).includes(source)
    const title = sourceIsValid ? SourceToTitleMap[source as ExperimentalSignUpSource] : SourceToTitleMap.Context // Use Context as default

    const logEvent = (): void => {
        if (sourceIsValid) {
            telemetryService.log(`SignUpPLG${source || ''}_2_ClickedSignUp`)
        }
    }

    return (
        <div className={styles.page}>
            <header>
                <div className="position-relative">
                    <div className={styles.headerBackground1} />
                    <div className={styles.headerBackground2} />
                    <div className={styles.headerBackground3} />

                    <div className={styles.limitWidth}>
                        <BrandLogo isLightTheme={isLightTheme} variant="logo" className={styles.logo} />
                    </div>
                </div>

                <div className={styles.limitWidth}>
                    <h2 className={styles.pageHeading}>{title}</h2>
                </div>
            </header>

            <div className={classNames(styles.contents, styles.limitWidth)}>
                <div className={styles.contentsLeft}>
                    With a Sourcegraph account, you can also:
                    <ul className={styles.featureList}>
                        <li>Search across all your public (and soon private) repositories</li>
                        <li>Monitor code for changes</li>
                        <li>Navigate through code with IDE like go to references and definition hovers</li>
                        <li>Integrate data, tooling, and code in a single location </li>
                    </ul>
                    <div className={styles.companiesHeader}>
                        Trusted by developers at the world's most innovative companies:
                    </div>
                    <img
                        src={`${assetsRoot}/img/customer-logos-${isLightTheme ? 'light' : 'dark'}.svg`}
                        alt="Cloudflare, Uber, SoFi, Dropbox, Plaid, Toast"
                    />
                </div>

                <div className={styles.signUpWrapper}>
                    <h2>Create a free account</h2>
                    {!showEmailForm ? (
                        <>
                            {githubProvider && (
                                <a
                                    href={githubProvider.authenticationURL}
                                    className={classNames(styles.signUpButton, styles.githubButton)}
                                    onClick={logEvent}
                                >
                                    <GithubIcon className="mr-3" /> Continue with GitHub
                                </a>
                            )}
                            {gitlabProvider && (
                                <a
                                    href={gitlabProvider.authenticationURL}
                                    className={classNames(styles.signUpButton, styles.gitlabButton)}
                                    onClick={logEvent}
                                >
                                    <GitlabColorIcon className="mr-3" /> Continue with GitLab
                                </a>
                            )}

                            <div className="mb-4">
                                Or,{' '}
                                <Link to={`${location.pathname}?${queryWithUseEmailToggled.toString()}`}>
                                    continue with email
                                </Link>
                            </div>
                        </>
                    ) : (
                        <>
                            <small className="d-block mt-3">
                                <Link
                                    className="d-flex align-items-center"
                                    to={`${location.pathname}?${queryWithUseEmailToggled.toString()}`}
                                >
                                    <ChevronLeftIcon className={classNames('icon-inline', styles.backIcon)} />
                                    Go back
                                </Link>
                            </small>

                            <SignUpForm
                                onSignUp={args => {
                                    logEvent()
                                    return onSignUp(args)
                                }}
                                context={{ authProviders: [], sourcegraphDotComMode: true }}
                                buttonLabel="Sign up"
                                experimental={true}
                                className="my-3"
                            />
                        </>
                    )}

                    <small className="text-muted">
                        By registering, you agree to our{' '}
                        <a href="https://about.sourcegraph.com/terms" target="_blank" rel="noopener">
                            Terms of Service
                        </a>{' '}
                        and{' '}
                        <a href="https://about.sourcegraph.com/privacy" target="_blank" rel="noopener">
                            Privacy Policy
                        </a>
                        .
                    </small>

                    <hr className={styles.separator} />

                    <div>
                        Already have an account? <Link to={`/sign-in${location.search}`}>Log in</Link>
                    </div>
                </div>
            </div>
        </div>
    )
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
