import React from 'react'
import { Link } from 'react-router-dom'

export const WelcomeAreaFooter: React.FunctionComponent<{ isLightTheme: boolean }> = ({ isLightTheme }) => (
    <>
        <div className="row mt-5 pt-4 border-top">
            <div className="col-sm-4 col-m-4 col-lg-4">
                <img
                    className="mb-2"
                    src={
                        isLightTheme
                            ? 'https://about.sourcegraph.com/sourcegraph/logo.svg'
                            : 'https://about.sourcegraph.com/sourcegraph/logo--light.svg'
                    }
                />
                <p>
                    <a href="mailto:hi@sourcegraph.com" target="_blank">
                        hi@sourcegraph.com
                    </a>
                    <br />
                    142 Minna St, 2nd Floor
                    <br />
                    San Francisco, CA 94105 (USA)
                </p>
            </div>
            <div className="col-xs-12 col-sm-12 col-md-2 col-lg-2">
                <h3 className="mb-0">Features</h3>
                <ul className="list-unstyled">
                    <li>
                        <Link to="/welcome/search">Code search</Link>
                    </li>
                    <li>
                        <Link to="/welcome/code-intelligence">Code intelligence</Link>
                    </li>
                    <li>
                        <Link to="/welcome/integrations">Integrations</Link>
                    </li>
                    <li>
                        <a href="https://about.sourcegraph.com/pricing" target="_blank">
                            Enterprise
                        </a>
                    </li>
                </ul>
            </div>
            <div className="col-xs-12 col-sm-12 col-md-2 col-lg-2">
                <h3 className="mb-0">Community</h3>
                <ul className="list-unstyled">
                    <li>
                        <a href="https://github.com/sourcegraph/sourcegraph" target="_blank">
                            GitHub
                        </a>
                    </li>
                    <li>
                        <a href="https://about.sourcegraph.com/blog" target="_blank">
                            Blog
                        </a>
                    </li>
                    <li>
                        <a href="https://twitter.com/srcgraph" target="_blank">
                            Twitter
                        </a>
                    </li>
                    <li>
                        <a href="https://www.linkedin.com/company/4803356/" target="_blank">
                            LinkedIn
                        </a>
                    </li>
                </ul>
            </div>
            <div className="col-xs-12 col-sm-12 col-md-2 col-lg-2">
                <h3 className="mb-0">Company</h3>
                <ul className="list-unstyled">
                    <li>
                        <a href="https://about.sourcegraph.com/plan" target="_blank">
                            Master plan
                        </a>
                    </li>
                    <li>
                        <a href="https://about.sourcegraph.com/about" target="_blank">
                            About
                        </a>
                    </li>
                    <li>
                        <a href="https://about.sourcegraph.com/contact" target="_blank">
                            Contact
                        </a>
                    </li>
                    <li>
                        <a href="https://about.sourcegraph.com/jobs" target="_blank">
                            Careers
                        </a>
                    </li>
                </ul>
            </div>
            <div className="col-xs-12 col-sm-12 col-md-2 col-lg-2">
                <h3 className="mb-0">Resources</h3>
                <ul className="list-unstyled">
                    <li>
                        <a href="https://docs.sourcegraph.com" target="_blank">
                            Documentation
                        </a>
                    </li>
                    <li>
                        <a
                            href="https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/CHANGELOG.md"
                            target="_blank"
                        >
                            Changelog
                        </a>
                    </li>
                    <li>
                        <a href="https://about.sourcegraph.com/pricing" target="_blank">
                            Pricing
                        </a>
                    </li>
                    <li>
                        <a href="https://about.sourcegraph.com/security" target="_blank">
                            Security
                        </a>
                    </li>
                </ul>
            </div>
        </div>
        <p className="text-muted mt-3">
            <a href="https://about.sourcegraph.com/terms" target="_blank">
                Terms
            </a>{' '}
            -{' '}
            <a href="https://about.sourcegraph.com/privacy" target="_blank">
                Privacy
            </a>{' '}
            - Copyright © 2018 Sourcegraph, Inc.
        </p>
    </>
)
