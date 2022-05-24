import React from 'react'

import classNames from 'classnames'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import { useLocation } from 'react-router-dom'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Link, Icon, Typography } from '@sourcegraph/wildcard'

import { BrandLogo } from '../components/branding/BrandLogo'
import { VSCodeIcon } from '../components/CtaIcons'
import { AuthProvider, SourcegraphContext } from '../jscontext'

import { ExternalsAuth } from './ExternalsAuth'
import { SignUpArguments, SignUpForm } from './SignUpForm'

import styles from './VsCodeSignUpPage.module.scss'

export const ShowEmailFormQueryParameter = 'showEmail'
interface Props extends ThemeProps, TelemetryProps {
    source: string | null
    showEmailForm: boolean
    /** Called to perform the signup on the server. */
    onSignUp: (args: SignUpArguments) => Promise<void>
    context: Pick<SourcegraphContext, 'authProviders' | 'experimentalFeatures'>
}

/**
 * Sign up page specifically from users via our VS Code integration
 */
export const VsCodeSignUpPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    isLightTheme,
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

    const logEvent = (type: AuthProvider['serviceType']): void => {
        const eventType = type === 'builtin' ? 'form' : type
        telemetryService.log(
            'SignupInitiated',
            { type: eventType, source: 'vs-code' },
            { type: eventType, source: 'vs-code' }
        )
    }

    const signUpForm = (
        <SignUpForm
            onSignUp={args => {
                logEvent('builtin')
                return onSignUp(args)
            }}
            context={{
                authProviders: [],
                sourcegraphDotComMode: true,
                experimentalFeatures: context.experimentalFeatures,
            }}
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
                    <Icon className={styles.backIcon} as={ChevronLeftIcon} />
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
                    <Typography.H2
                        className={classNames('d-flex', 'align-items-center', 'mb-3', 'mt-1', styles.pageHeading)}
                    >
                        <div className={classNames(styles.iconCirlce, 'mr-3')}>
                            <VSCodeIcon />
                        </div>{' '}
                        <strong className="mr-1">Unlock the full potential of the Sourcegraph extension</strong>
                    </Typography.H2>
                    With a Sourcegraph account, you can:
                    <ul className={styles.featureList}>
                        <li>Search all of your code from your code host, even without downloading it locally</li>
                        <li>Reference and re-use code from all your projects without leaving VS Code</li>
                        <li>Create code monitors to alert you to changes in code</li>
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
                    {' '}
                    <Typography.H2>Create a free account</Typography.H2>
                    {renderAuthMethod()}
                    <small className="text-muted">
                        By registering, you agree to our{' '}
                        <Link to="https://about.sourcegraph.com/terms" target="_blank" rel="noopener">
                            Terms of Service
                        </Link>{' '}
                        and{' '}
                        <Link to="https://about.sourcegraph.com/privacy" target="_blank" rel="noopener">
                            Privacy Policy
                        </Link>
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
