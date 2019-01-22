import * as H from 'history'
import CloseIcon from 'mdi-react/CloseIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { HeroPage } from '../../../components/HeroPage'
import { CodeIntellifyBlob } from './CodeIntellifyBlob'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    location: H.Location
    history: H.History
}

const heroEyebrow = 'Sourcegraph'
const heroTitle = 'Search, navigate, and review code. Find answers.'
const heroCopyTop =
    'Sourcegraph is a free, open-source, self-hosted code search and navigation tool for developers. Use it with any Git code host for teams of any size.'
const heroCopyBottom = 'Upgraded features available for enterprise users.'

const inlineStyle = `
    .hero-tooltip {
        z-index: 1;
        position: fixed !important;
        transform: translateY(44px);
    }
    .modal-tooltip {
        z-index: 9999 !important;
        opacity: 1 !important;
        visbility: visbily !important;
    }
`
// Set the defauly hover token of the hero tooltip
const defaultTooltipHeroPosition = { line: 244, character: 11 }

/**
 * The welcome main page, which describes Sourcegraph functionality and other general information.
 */
export class WelcomeMainPage extends React.Component<Props> {
    public render(): JSX.Element | null {
        window.context.sourcegraphDotComMode = true // TODO!(sqs)
        if (!window.context.sourcegraphDotComMode) {
            return <HeroPage icon={MapSearchIcon} title="Page not found" />
        }
        return (
            <div className="welcome-area">
                <style>{inlineStyle}</style>
                <section className="hero-section">
                    <div className="container hero-container">
                        <div className="row">
                            <div className="col-lg-6 col-md-12 col-sm-12">
                                <h2>{heroEyebrow}</h2>
                                <h1 className="mt-2">{heroTitle}</h1>
                                <p>{heroCopyTop}</p>
                                <p>{heroCopyBottom}</p>
                                <a className="btn btn-primary" href="https://docs.sourcegraph.com/#quickstart">
                                    Deploy Sourcegraph
                                </a>
                                <a
                                    className="btn btn-secondary"
                                    href="https://github.com/sourcegraph/sourcegraph"
                                    target="_blank"
                                >
                                    Sourcegraph on GitHub
                                </a>
                            </div>
                            <div className="col-6 small-hidden">
                                <CodeIntellifyBlob
                                    {...this.props}
                                    startLine={236}
                                    endLine={250}
                                    parentElement={'.hero-section'}
                                    className={'code-intellify-container'}
                                    tooltipClass={'hero-tooltip'}
                                    defaultHoverPosition={defaultTooltipHeroPosition}
                                />
                            </div>
                        </div>
                    </div>
                </section>
                <div className="about-section-container">
                    <section className="company-logos">
                        <div className="container">
                            <div className="row">
                                <div className="col-12">
                                    <h2>Powering developers at</h2>
                                </div>
                                <div className="col-3 welcome-company-logo">
                                    <figure className="welcome-company-logo--image welcome-company-logo__1" />
                                </div>
                                <div className="col-3 welcome-company-logo">
                                    <figure className="welcome-company-logo--image welcome-company-logo__2" />
                                </div>
                                <div className="col-3 welcome-company-logo">
                                    <figure className="welcome-company-logo--image welcome-company-logo__3" />
                                </div>
                                <div className="col-3 welcome-company-logo">
                                    <figure className="welcome-company-logo--image welcome-company-logo__4" />
                                </div>
                            </div>
                        </div>
                    </section>
                    <section className="about-section code-search">
                        <div className="container">
                            <div className="row">
                                <div className="col-12">
                                    <h2>Advanced code search</h2>
                                    <h1>Find. Then replace.</h1>
                                    <p>
                                        Search in files and diffs on your private code using simple terms, regular
                                        expressions, and other filters.
                                    </p>
                                    <p>
                                        Syncs repositories with your code host and supports searching any commit/branch,
                                        with no indexing delay.
                                    </p>
                                    <Link className="btn btn-secondary" to="/welcome/search">
                                        Explore code search
                                    </Link>
                                </div>
                            </div>
                        </div>
                    </section>
                    <section className="about-section code-intelligence">
                        <div className="container">
                            <div className="row">
                                <div className="col-12">
                                    <h2>Enhanced code browsing and intelligence</h2>
                                    <h1>Mine your language.</h1>
                                    <p>
                                        Solve problems before they exist, commit by commit. Code intelligence makes
                                        browsing code easier, with IDE-like hovers, go-to-definition, and
                                        find-references on your code, powered by Sourcegraph extensions and language
                                        servers based on the open-source Language Server Protocol.
                                    </p>
                                    <p>
                                        It even works in code review diffs on GitHub and GitLab with our browser
                                        extensions.
                                    </p>
                                    <Link className="btn btn-secondary" to="/welcome/code-intelligence">
                                        Explore code intelligence
                                    </Link>
                                </div>
                            </div>
                        </div>
                    </section>
                    <section className="about-section integrations">
                        <div className="container">
                            <div className="row">
                                <div className="col-12">
                                    <h2>Integrations</h2>
                                    <h1>Get it. Together.</h1>
                                    <p>
                                        Connect your Sourcegraph instance with your existing tools. Get code
                                        intelligence while browsing code on the web, and code search from your editor.
                                    </p>
                                    <Link className="btn btn-secondary" to="/welcome/integrations">
                                        Explore integrations
                                    </Link>
                                </div>
                            </div>
                        </div>
                    </section>
                </div>
                <section className="up-next-section">
                    <div className="container">
                        <div className="row">
                            <div className="col-lg-6 col-md-12">
                                <h2>Deploy Sourcegraph</h2>
                                <h1>Free. For all.</h1>
                                <p>
                                    The pace at which humans can write code is the only thing that stands between us and
                                    flying cars, a habitat on Mars, and a cure for cancer. That's why developers can get
                                    started and deploy Sourcegraph for free, and contribute to our code on GitHub.
                                </p>
                                <a className="btn btn-primary" href="https://docs.sourcegraph.com/#quickstart">
                                    Deploy Sourcegraph
                                </a>
                                <a
                                    className="btn btn-secondary"
                                    href="https://github.com/sourcegraph/sourcegraph/"
                                    target="_blank"
                                >
                                    Sourcegraph on GitHub
                                </a>
                            </div>
                            <div className="col-lg-6 col-md-12">
                                <h2>Sourcegraph pricing</h2>
                                <h1>Size. Up.</h1>
                                <p>
                                    When you grow to hundreds or thousands of users and repositories, scale up
                                    instantly, and protect your uptime with Sourcegraph on Kubernetes, external backups,
                                    and custom support agreements. Start with Sourcegraph core for free and scale with
                                    your deployment.
                                </p>
                                <a className="btn btn-secondary" href="//about.sourcegraph.com/pricing/">
                                    Sourcegraph pricing
                                </a>
                            </div>
                        </div>
                    </div>
                </section>
                <section className="up-next-section blog-callout">
                    <div className="container">
                        <div className="row">
                            <div className="col-12">
                                <h2>Open. For business.</h2>
                                <h1>Sourcegraph is open source.</h1>
                                <p>
                                    We opened up Sourcegraph to bring code search and intelligence to more developers
                                    and developer ecosystems—and to help us realize the{' '}
                                    <a href="//about.sourcegraph.com/plan/">Sourcegraph master plan</a>. We're also
                                    excited about what this means for Sourcegraph as a company. All of our customers,
                                    many with hundreds or thousands of developers using Sourcegraph internally every
                                    day, started out with a single developer spinning up a Sourcegraph instance and
                                    sharing it with their team. Being open-source makes it even easier to use
                                    Sourcegraph.
                                </p>
                                <a
                                    className="btn btn-primary"
                                    href="https://about.sourcegraph.com/blog/sourcegraph-is-now-open-source/"
                                >
                                    Release announcement
                                </a>
                                <a className="btn btn-secondary" href="https://github.com/sourcegraph/sourcegraph/">
                                    Sourcegraph on GitHub
                                </a>
                            </div>
                        </div>
                    </div>
                </section>
                <section className="footer-section">
                    <div className="container">
                        <div className="row">
                            <div className="col-sm-4 col-m-4 col-lg-4 item logo__section">
                                <img
                                    className="footer__logo"
                                    src="https://about.sourcegraph.com/sourcegraph/logo--light.svg"
                                />
                                <div className="footer__contact">
                                    <p>
                                        <a className="mail__contect" href="mailto:hi@sourcegraph.com" target="_blank">
                                            hi@sourcegraph.com
                                        </a>
                                    </p>
                                    <p className="addr__contact">
                                        142 Minna St, 2nd Floor
                                        <br />
                                        San Francisco CA, 94105
                                    </p>
                                </div>
                            </div>
                            <div className="col-xs-12 col-sm-12 col-m-2 col-lg-2 item footer__extend community">
                                <h3>Community</h3>
                                <input type="checkbox" />
                                <ul>
                                    <li>
                                        <a href="//about.sourcegraph.com/blog">Blog</a>
                                    </li>
                                    <li>
                                        <a href="https://github.com/sourcegraph" target="_blank">
                                            GitHub
                                        </a>
                                    </li>
                                    <li>
                                        <a href="https://www.linkedin.com/company/4803356/" target="_blank">
                                            LinkedIn
                                        </a>
                                    </li>
                                    <li>
                                        <a href="https://twitter.com/srcgraph" target="_blank">
                                            Twitter
                                        </a>
                                    </li>
                                </ul>
                                <div className="close--icon">
                                    <CloseIcon className="material-icons" />
                                </div>
                            </div>
                            <div className="col-xs-12 col-sm-12 col-m-2 col-lg-2 item footer__extend company">
                                <h3>Company</h3>
                                <input type="checkbox" />
                                <ul>
                                    <li>
                                        <a href="//about.sourcegraph.com/plan">Master Plan</a>
                                    </li>
                                    <li>
                                        <a href="//about.sourcegraph.com/about">About</a>
                                    </li>
                                    <li>
                                        <a href="//about.sourcegraph.com/contact">Contact</a>
                                    </li>
                                    <li>
                                        <a href="//about.sourcegraph.com/jobs">Careers</a>
                                    </li>
                                </ul>
                            </div>
                            <div className="col-xs-12 col-sm-12 col-m-2 col-lg-2 item footer__extend features">
                                <h3>Features</h3>
                                <input type="checkbox" />
                                <ul>
                                    <li>
                                        <Link to="/welcome/search">Code Search</Link>
                                    </li>
                                    <li>
                                        <Link to="/welcome/code-intelligence">Code intelligence</Link>
                                    </li>
                                    <li>
                                        <Link to="/welcome/integrations">Integrations</Link>
                                    </li>
                                    <li>
                                        <a href="//about.sourcegraph.com/pricing">Enterprise</a>
                                    </li>
                                </ul>
                            </div>
                            <div className="col-xs-12 col-sm-12 col-m-2 col-lg-2 item footer__extend resources">
                                <h3>Resources</h3>
                                <input type="checkbox" />
                                <ul>
                                    <li>
                                        <a href="https://docs.sourcegraph.com">Documentation</a>
                                    </li>
                                    <li>
                                        <a href="//sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/CHANGELOG.md">
                                            Changelog
                                        </a>
                                    </li>
                                    <li>
                                        <a href="//about.sourcegraph.com/pricing">Pricing</a>
                                    </li>
                                    <li>
                                        <a href="//about.sourcegraph.com/security">Security</a>
                                    </li>
                                </ul>
                            </div>
                        </div>
                        <div className="small__contact">
                            <div className="footer__contact">
                                <p>
                                    <a className="mail__contect" href="mailto:hi@sourcegraph.com" target="_blank">
                                        hi@sourcegraph.com
                                    </a>
                                </p>
                                <p className="addr__contact">
                                    142 Minna St, 2nd Floor
                                    <br />
                                    San Francisco CA, 94105
                                </p>
                            </div>
                        </div>
                    </div>
                    <div className="container">
                        <div className="row">
                            <div className="col-lg-12 copyright__container item">
                                <span>
                                    <p className="copyright">Copyright © 2018 Sourcegraph, Inc.</p>
                                </span>
                                <span className="terms">
                                    <a href="//about.sourcegraph.com/terms">Terms</a>
                                    <a href="//about.sourcegraph.com/privacy">Privacy</a>
                                </span>
                            </div>
                        </div>
                    </div>
                </section>
            </div>
        )
    }
}
