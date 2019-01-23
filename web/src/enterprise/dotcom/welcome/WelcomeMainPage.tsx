import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
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
            <div className="welcome-area">
                <section className="hero-section">
                    <div className="container hero-container">
                        <div className="row">
                            <div className="col-lg-6">
                                <h2>Sourcegraph</h2>
                                <h1>Open. For business.</h1>
                                <p>
                                    Sourcegraph is a free, open-source, self-hosted code search and intelligence server
                                    that helps developers find, review, understand, and debug code. Use it with any Git
                                    code host for teams of any size. Start using it now, then install the Sourcegraph
                                    Docker image on your private code.
                                </p>
                                <p>Upgraded features available for enterprise users.</p>
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
                            <div className="col-lg-6">
                                <div
                                    className="bg-secondary border d-flex justify-content-center align-items-center"
                                    // tslint:disable-next-line:jsx-ban-props
                                    style={{ width: '360px', height: '240px', textTransform: 'uppercase' }}
                                >
                                    demo video placeholder
                                </div>
                            </div>
                        </div>
                    </div>
                </section>
                <section>
                    <h2>Powering developers at</h2>
                    <div className="row">
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
                </section>
                <section>
                    <h2>Advanced code search</h2>
                    <h1>Find. Then replace.</h1>
                    <p>
                        Search in files and diffs on your private code using simple terms, regular expressions, and
                        other filters.
                    </p>
                    <p>
                        Syncs repositories with your code host and supports searching any commit/branch, with no
                        indexing delay.
                    </p>
                    <Link className="btn btn-secondary" to="/welcome/search">
                        Explore code search
                    </Link>
                </section>
                <section>
                    <h2>Enhanced code browsing and intelligence</h2>
                    <h1>Mine your language.</h1>
                    <p>
                        Solve problems before they exist, commit by commit. Code intelligence makes browsing code
                        easier, with IDE-like hovers, go-to-definition, and find-references on your code, powered by
                        Sourcegraph extensions and language servers based on the open-source Language Server Protocol.
                    </p>
                    <p>It even works in code review diffs on GitHub and GitLab with our browser extensions.</p>
                    <Link className="btn btn-secondary" to="/welcome/code-intelligence">
                        Explore code intelligence
                    </Link>
                </section>
                <section>
                    <h2>Integrations</h2>
                    <h1>Get it. Together.</h1>
                    <p>
                        Connect your Sourcegraph instance with your existing tools. Get code intelligence while browsing
                        code on the web, and code search from your editor.
                    </p>
                    <Link className="btn btn-secondary" to="/welcome/integrations">
                        Explore integrations
                    </Link>
                </section>
                <div className="row">
                    <section className="col-lg-6 col-md-12">
                        <h2>Deploy Sourcegraph</h2>
                        <h1>Free. For all.</h1>
                        <p>
                            The pace at which humans can write code is the only thing that stands between us and flying
                            cars, a habitat on Mars, and a cure for cancer. That's why developers can get started and
                            deploy Sourcegraph for free, and contribute to our code on GitHub.
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
                    </section>
                    <section className="col-lg-6 col-md-12">
                        <h2>Sourcegraph pricing</h2>
                        <h1>Size. Up.</h1>
                        <p>
                            When you grow to hundreds or thousands of users and repositories, scale up instantly, and
                            protect your uptime with Sourcegraph on Kubernetes, external backups, and custom support
                            agreements. Start with Sourcegraph Core for free and scale with your deployment.
                        </p>
                        <a className="btn btn-secondary" href="https://about.sourcegraph.com/pricing/">
                            Sourcegraph pricing
                        </a>
                    </section>
                </div>
                <section>
                    <h2>Open. For business.</h2>
                    <h1>Sourcegraph is open source.</h1>
                    <p>
                        We opened up Sourcegraph to bring code search and intelligence to more developers and developer
                        ecosystems—and to help us realize the{' '}
                        <a href="https://about.sourcegraph.com/plan/">Sourcegraph master plan</a>. We're also excited
                        about what this means for Sourcegraph as a company. All of our customers, many with hundreds or
                        thousands of developers using Sourcegraph internally every day, started out with a single
                        developer spinning up a Sourcegraph instance and sharing it with their team. Being open-source
                        makes it even easier to use Sourcegraph.
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
                </section>
            </div>
        )
    }
}
