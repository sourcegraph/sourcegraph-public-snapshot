import * as H from 'history'
import CloudUploadIcon from 'mdi-react/CloudUploadIcon'
import FirefoxIcon from 'mdi-react/FirefoxIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { Logo1, Logo2 } from './logos'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    authenticatedUser: GQL.IUser | null
    isLightTheme: boolean
    location: H.Location
    history: H.History
}

/**
 * The welcome main page, which describes Sourcegraph functionality and other general information.
 */
export class WelcomeMainPage extends React.Component<Props> {
    public render(): JSX.Element | null {
        return (
            <div className="welcome-main-page">
                <section className="hero-section">
                    <div className="container hero-container mt-5 pt-3">
                        <div className="row justify-content-md-center">
                            <div className="col-md-7 col-lg-6 mr-lg-4 mb-4">
                                <img
                                    className="welcome-main-page__logo-mark mb-1"
                                    src="/.assets/img/sourcegraph-mark.svg"
                                />
                                <h2 className="welcome-main-page__header mt-2">
                                    <span className="font-weight-normal">
                                        Search,&nbsp;navigate, and review&nbsp;code.
                                    </span>{' '}
                                    Find&nbsp;answers.
                                </h2>
                                <p>Sourcegraph is a web-based code search and navigation tool for dev teams.</p>
                                <ul className="pl-3">
                                    <li>
                                        <strong>Code search:</strong> fast, cross-repository, works on any commit (no
                                        indexing delay), with support for regexps, punctuation, diffs, and{' '}
                                        <a href="https://docs.sourcegraph.com/user/search/queries" target="_blank">
                                            filters
                                        </a>
                                    </li>
                                    <li>
                                        <strong>Code intelligence:</strong> go-to-definition and find-references for{' '}
                                        <a
                                            href="https://sourcegraph.com/extensions?query=category%3A%22Programming+languages%22"
                                            target="_blank"
                                        >
                                            top languages
                                        </a>{' '}
                                        (via{' '}
                                        <a href="https://docs.sourcegraph.com/extensions" target="_blank">
                                            extensions
                                        </a>
                                        )
                                    </li>
                                    <li>
                                        <strong>Deep integrations</strong> with GitHub, GitHub Enterprise, GitLab,
                                        Bitbucket, Phabricator, etc.
                                    </li>
                                    <li>
                                        <a href="https://github.com/sourcegraph/sourcegraph" target="_blank">
                                            Open-source
                                        </a>
                                        , self-hosted, and free (paid{' '}
                                        <a href="https://about.sourcegraph.com/pricing" target="_blank">
                                            Enterprise
                                        </a>{' '}
                                        upgrade available)
                                    </li>
                                </ul>
                                <p className="mb-1">
                                    <a href="https://docs.sourcegraph.com/user/tour" target="_blank">
                                        Learn how it's used
                                    </a>{' '}
                                    to build better software faster at:
                                </p>
                                <div className="welcome-main-page__customer-logos d-flex align-items-center pl-2">
                                    <Logo1
                                        className="welcome-main-page__customer-logo welcome-main-page__customer-logo-1 mr-3"
                                        isLightTheme={this.props.isLightTheme}
                                    />
                                    <Logo2
                                        className="welcome-main-page__customer-logo welcome-main-page__customer-logo-2 mr-4"
                                        isLightTheme={this.props.isLightTheme}
                                    />
                                    <span className="small text-muted">
                                        &hellip;and thousands of other organizations.
                                    </span>
                                </div>
                            </div>
                            <div className="col-md-5 col-lg-4 mb-4">
                                <div className="mt-3">
                                    <a
                                        className="btn btn-primary btn-lg font-weight-bold mb-1 d-inline-flex align-items-center text-nowrap flex-wrap justify-content-center"
                                        href="https://docs.sourcegraph.com/#quickstart"
                                    >
                                        <CloudUploadIcon className="icon-inline mr-2" /> Deploy self-hosted Sourcegraph
                                    </a>
                                    <small className="text-muted d-block">
                                        For use with your organization's private code. Runs securely on your infra (in a
                                        single Docker container or on a cluster).
                                    </small>
                                </div>
                                <div className="mt-4">
                                    <a
                                        className="btn btn-secondary mb-1 d-inline-flex align-items-center"
                                        href="https://docs.sourcegraph.com/integration/browser_extension"
                                    >
                                        Install browser extension
                                    </a>
                                    <small className="text-muted d-block">
                                        Adds go-to-definition and find-references to GitHub and other code hosts. To use
                                        with private code, connect it to your self-hosted Sourcegraph instance.
                                    </small>
                                </div>
                                {!!this.props.authenticatedUser && ( // TODO!(sqs): single negation
                                    <div className="mt-4">
                                        {/* TODO!(sqs): make this different if signed in */}
                                        <Link to="/sign-up" target="_blank" className="welcome-main-page__sign-up">
                                            Sign up on Sourcegraph.com
                                        </Link>
                                        <small className="text-muted d-block">
                                            A public Sourcegraph instance for public code only.
                                        </small>
                                    </div>
                                )}
                            </div>
                        </div>
                        <div className="row justify-content-md-center mt-3">
                            <iframe
                                className="welcome-main-page__demo col-md-12 col-lg-10"
                                src="https://www.youtube.com/embed/Pfy2CjeJ2N4"
                                frameBorder="0"
                                allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture"
                                allowFullScreen={true}
                            />
                        </div>
                    </div>
                </section>
            </div>
        )
    }
}
