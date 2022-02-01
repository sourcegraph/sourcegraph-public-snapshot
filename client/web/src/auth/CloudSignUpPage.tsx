import classNames from 'classnames'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import React from 'react'
import { useLocation } from 'react-router-dom'

import { useQuery } from '@sourcegraph/http-client'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { ProductStatusBadge, Link } from '@sourcegraph/wildcard'

import { BrandLogo } from '../components/branding/BrandLogo'
import { FeatureFlagProps } from '../featureFlags/featureFlags'
import { UserAreaUserProfileResult, UserAreaUserProfileVariables } from '../graphql-operations'
import { AuthProvider, SourcegraphContext } from '../jscontext'
import { USER_AREA_USER_PROFILE } from '../user/area/UserArea'
import { UserAvatar } from '../user/UserAvatar'

import styles from './CloudSignUpPage.module.scss'
import { ExternalsAuth } from './ExternalsAuth'
import { SignUpArguments, SignUpForm } from './SignUpForm'

interface Props extends ThemeProps, TelemetryProps, FeatureFlagProps {
    source: string | null
    showEmailForm: boolean
    /** Called to perform the signup on the server. */
    onSignUp: (args: SignUpArguments) => Promise<void>
    context: Pick<SourcegraphContext, 'authProviders' | 'experimentalFeatures'>
}

const SourceToTitleMap = {
    Context: 'Easily search the code you care about.',
    Saved: 'Create a library of useful searches.',
    Monitor: 'Monitor code for changes.',
    Extend: 'Augment code and workflows via extensions.',
    SearchCTA: 'Easily search the code you care about.',
    HomepageCTA: 'Easily search the code you care about.',
    Snippet: 'Easily search the code you care about.',
}

export type CloudSignUpSource = keyof typeof SourceToTitleMap

export const ShowEmailFormQueryParameter = 'showEmail'

/**
 * Sign up page specifically for Sourcegraph.com
 */
export const CloudSignUpPage: React.FunctionComponent<Props> = ({
    isLightTheme,
    source,
    showEmailForm,
    onSignUp,
    context,
    telemetryService,
    featureFlags,
}) => {
    const location = useLocation()

    const queryWithUseEmailToggled = new URLSearchParams(location.search)
    if (showEmailForm) {
        queryWithUseEmailToggled.delete(ShowEmailFormQueryParameter)
    } else {
        queryWithUseEmailToggled.append(ShowEmailFormQueryParameter, 'true')
    }

    const assetsRoot = window.context?.assetsRoot || ''
    const sourceIsValid = source && Object.keys(SourceToTitleMap).includes(source)
    const defaultTitle = SourceToTitleMap.Context
    const title = sourceIsValid ? SourceToTitleMap[source as CloudSignUpSource] : defaultTitle

    const invitedBy = queryWithUseEmailToggled.get('invitedBy')
    const { data } = useQuery<UserAreaUserProfileResult, UserAreaUserProfileVariables>(USER_AREA_USER_PROFILE, {
        variables: { username: invitedBy || '', siteAdmin: false },
        skip: !invitedBy,
    })
    const invitedByUser = data?.user

    const logEvent = (type: AuthProvider['serviceType']): void => {
        const eventType = type === 'builtin' ? 'form' : type
        telemetryService.log('SignupInitiated', { type: eventType }, { type: eventType })
    }

    const signUpForm = (
        <SignUpForm
            featureFlags={featureFlags}
            onSignUp={args => {
                logEvent('builtin')
                return onSignUp(args)
            }}
            context={{ authProviders: [], sourcegraphDotComMode: true }}
            buttonLabel="Sign up"
            experimental={true}
            className="my-3"
        />
    )

    const renderCodeHostAuth = (): JSX.Element => (
        <>
            <ExternalsAuth
                context={context}
                githubLabel="Continue with GitHub"
                gitlabLabel="Continue with GitLab"
                onClick={logEvent}
            />

            <div className="mb-4">
                Or, <Link to={`${location.pathname}?${queryWithUseEmailToggled.toString()}`}>continue with email</Link>
            </div>
        </>
    )

    const renderEmailAuthForm = (): JSX.Element => (
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

            {signUpForm}
        </>
    )

    const renderAuthMethod = (): JSX.Element => (showEmailForm ? renderEmailAuthForm() : renderCodeHostAuth())

    return (
        <div className={styles.page}>
            <header className="position-relative">
                <div className={styles.headerBackground1} />
                <div className={styles.headerBackground2} />
            </header>
            <div className={classNames('d-flex', 'justify-content-center', 'mb-5', styles.leftOrRightContainer)}>
                <div className={styles.leftOrRight}>
                    <BrandLogo isLightTheme={isLightTheme} variant="logo" className={styles.logo} />
                    <h2
                        className={classNames(
                            'd-flex',
                            'align-items-center',
                            'mb-4',
                            'mt-1',
                            invitedBy ? styles.pageHeadingInvitedBy : styles.pageHeading
                        )}
                    >
                        {invitedByUser ? (
                            <>
                                <UserAvatar
                                    className={classNames('icon-inline', 'mr-3', styles.avatar)}
                                    user={invitedByUser}
                                />
                                <strong className="mr-1">{invitedBy}</strong> has invited you to join Sourcegraph
                            </>
                        ) : (
                            title
                        )}
                    </h2>

                    {invitedBy ? 'With a Sourcegraph account, you can:' : 'With a Sourcegraph account, you can also:'}
                    <ul className={styles.featureList}>
                        <li>
                            <div className="d-flex align-items-center">
                                <ProductStatusBadge status="beta" className="text-uppercase mr-1" /> Search across all
                                your public and private repositories
                            </div>
                        </li>
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
                        className={styles.customerLogos}
                    />
                </div>

                <div className={classNames(styles.leftOrRight, styles.signUpWrapper)}>
                    <h2>Create a free account</h2>
                    {renderAuthMethod()}

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
