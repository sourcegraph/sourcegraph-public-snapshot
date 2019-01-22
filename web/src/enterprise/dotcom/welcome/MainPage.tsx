import * as H from 'history'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { HeroPage } from '../../../components/HeroPage'
import { eventLogger } from '../../../tracking/eventLogger'
import { CodeIntellifyBlob } from './CodeIntellifyBlob'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    authenticatedUser: GQL.IUser | null
    location: H.Location
    history: H.History
}

interface State {
    // modalXXopen sets a state that the modal is open before animations or closed after animation
    // modalXXclosing sets a state that starts the closing process
    modalIntelligenceOpen: boolean
    modalIntelligenceClosing: boolean
    modalIntegrationsOpen: boolean
    modalIntegrationsClosing: boolean
    // determine what button in modal is active
    activeButton?: string
    // determine what section inside a modal is active
    activesection?: string
    // animateModalXX starts the animation process after opening.
    animateModalIntelligence: boolean
    animateModalIntegrations: boolean
    // Manual click state is to determine if animation should be stopped
    manualClick?: boolean
}
const heroEyebrow = 'Sourcegraph'
const heroTitle = 'Search, navigate, and review code. Find answers.'
const heroCopyTop =
    'Sourcegraph is a free, open-source, self-hosted code search and navigation tool for developers. Use it with any Git code host for teams of any size.'
const heroCopyBottom = 'Upgraded features available for enterprise users.'

const intelligenceSections = [
    {
        title: 'Code browsing',
        paragraph:
            'View open source code, like gorilla/mux, on sourcegraph.com, or deploy your own instance to see public code alongside your private code. See how your codebase changes over time in by browsing through branches, commits, and diffs.',
    },
    {
        title: 'Advanced code intelligence',
        paragraph:
            'Code intelligence makes browsing code easier, with IDE-like hovers, go-to-definition, and find-references on your code, powered by language servers based on the open-source Language Server Protocol.',
    },
    {
        title: 'Hover tooltip',
        paragraph:
            'Use the hover tooltip to discover and understand your code faster. Click on a token and then go to its definition, other references, or implementations. Speed through reviews by understanding new code, changed code, and what it affects.',
    },
    {
        title: '',
        paragraph:
            'Code intelligence is powered by language servers based on the open-standard Language Server Protocol (published by Microsoft, with participation from Facebook, Google, Sourcegraph, GitHub, RedHat, Twitter, Salesforce, Eclipse, and others). Visit langserver.org to learn more about the Language Server Protocol, find the latest support for your favorite language, and get involved.',
    },
]

const integrationsSections = [
    {
        title: 'Connect across your development workflow.',
        paragraph:
            'Sourcegraph has powerful integrations for every step of development. From planning with code discussion, development with Sourcegraph and IDE extensions, to review in PRs and Issues. Use Sourcegraph integrations to get code intelligence at every step of your workflow.',
        buttons: [],
    },
    {
        title: 'Browser extensions',
        paragraph:
            'Code intelligence makes browsing code easier, with IDE-like hovers, go-to-definition, and find-references on your code, powered by language servers based on the open-source Language Server Protocol.',
        buttons: [
            {
                id: 'btn-chrome',
                text: 'Chrome',
                link: 'https://chrome.google.com/webstore/detail/sourcegraph/dgjhfomjieaadpoljlnidmbgkdffpack',
            },
        ],
    },

    {
        title: 'Code host integrations',
        paragraph:
            'The Sourcegraph browser extension will add go-to-definition, find-references, hover tooltips, and code search to all files and diffs on supported code hosts. The extension will also add code intelligence and code search to public repositories. ',
        buttons: [
            { id: 'btn-gitlab', text: 'GitLab', link: 'https://docs.sourcegraph.com/integration/browser_extension' },
            { id: 'btn-github', text: 'GitHub', link: 'https://docs.sourcegraph.com/integration/browser_extension' },
            {
                id: 'btn-phabricator',
                text: 'Phabricator',
                link: 'https://docs.sourcegraph.com/integration/browser_extension',
            },
        ],
    },
    {
        title: 'Editor extensions',
        paragraph:
            'Our editor plugins let you quickly jump to files and search code on your Sourcegraph instance from your editor. Seamlessly jump for development to review without missing a step.',
        buttons: [
            { id: 'btn-atom', text: 'Atom', link: 'https://atom.io/packages/sourcegraph' },
            { id: 'btn-intellij', text: 'IntelliJ', link: 'https://plugins.jetbrains.com/plugin/9682-sourcegraph' },
            {
                id: 'btn-sublime',
                text: 'Sublime',
                link: 'https://github.com/sourcegraph/sourcegraph-sublime',
            },
            {
                id: 'btn-vscode',
                text: 'Visual Studio Code',
                link: 'https://marketplace.visualstudio.com/items?itemName=sourcegraph.sourcegraph',
            },
        ],
    },
]

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

// Set the defauly hover token of the hero tooltip
const defaultTooltipModalPosition = { line: 248, character: 11 }

/**
 * The main page
 */
export class MainPage extends React.Component<Props, State> {
    public state: State = {
        modalIntelligenceOpen: false,
        modalIntelligenceClosing: false,
        modalIntegrationsOpen: false,
        modalIntegrationsClosing: false,
        manualClick: false,
        activesection: 'none',
        animateModalIntelligence: false,
        animateModalIntegrations: false,
    }

    private overlayPortal: HTMLElement | undefined

    public componentDidMount(): void {
        eventLogger.logViewEvent('Home')

        const portal = document.createElement('div')
        document.body.appendChild(portal)
        this.overlayPortal = portal
    }

    public componentWillUnmount(): void {
        const windowBody = document.body
        windowBody.classList.remove('modal-open')
    }

    public render(): JSX.Element | null {
        window.context.sourcegraphDotComMode = true // TODO!(sqs)
        if (!window.context.sourcegraphDotComMode) {
            return <HeroPage icon={MapSearchIcon} title="Page not found" />
        }
        return (
            <div className="main-page container-fluid px-0">
                <style>{inlineStyle}</style>
                <section className="hero-section">
                    <div className=" container hero-container">
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
                                    overlayPortal={this.overlayPortal}
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
                                <div className="col-3 logo">
                                    <figure className="logo-image logo1" />
                                </div>
                                <div className="col-3 logo">
                                    <figure className="logo-image logo2" />
                                </div>
                                <div className="col-3 logo">
                                    <figure className="logo-image logo3" />
                                </div>
                                <div className="col-3 logo">
                                    <figure className="logo-image logo4" />
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
                                    <button
                                        className="btn btn-secondary"
                                        id="sampleButton"
                                        onClick={this.activateModal('intelligence')}
                                    >
                                        Explore code intelligence
                                    </button>
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
                                    <button
                                        className="btn btn-secondary"
                                        id="sampleButton"
                                        onClick={this.activateModal('integrations')}
                                    >
                                        Explore integrations
                                    </button>
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
                                <div className="close--icon">
                                    <CloseIcon className="material-icons" />
                                </div>
                            </div>
                            <div className="col-xs-12 col-sm-12 col-m-2 col-lg-2 item footer__extend features">
                                <h3>Features</h3>
                                <input type="checkbox" />
                                <ul>
                                    <li>
                                        <a onClick={this.activateModal('search')}>Code search</a>
                                    </li>
                                    <li>
                                        <a onClick={this.activateModal('intelligence')}>Code intelligence</a>
                                    </li>
                                    <li>
                                        <a onClick={this.activateModal('integrations')}>Integrations</a>
                                    </li>
                                    <li>
                                        <a href="//about.sourcegraph.com/pricing">Enterprise</a>
                                    </li>
                                </ul>
                                <div className="close--icon">
                                    <CloseIcon className="material-icons" />
                                </div>
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
                                <div className="close--icon">
                                    <CloseIcon className="material-icons" />
                                </div>
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
                <div
                    className={`modal-intelligence ${this.state.modalIntelligenceOpen ? 'modal-open' : 'modal-close'} ${
                        this.state.modalIntelligenceClosing ? 'modal-closing' : ''
                    }`}
                >
                    <div className="container">
                        <button className="btn-close-top" onClick={this.closeModal('intelligence')}>
                            <CloseIcon className="material-icons" />
                        </button>
                        <div className="row copy-section">
                            <div className="col-12 modal-header">
                                <h2>Enhanced code browsing and intelligence</h2>
                                <h1>Mine your language.</h1>
                            </div>
                        </div>
                        <div className="row intelligence-row">
                            <div className="col-lg-6 col-md-12 col-sm-12 modal-header copy-section">
                                {intelligenceSections.map(({ title, paragraph }, i) => (
                                    <div
                                        key={`search-sections-${i}`}
                                        className={`row modal-copy-row ${
                                            this.state.activesection === `${i}` || this.state.activesection === '99'
                                                ? 'activesec'
                                                : ''
                                        }`}
                                    >
                                        <div className="col-12">
                                            <h3>{title}</h3>
                                            <p>{paragraph}</p>
                                        </div>
                                    </div>
                                ))}
                            </div>
                            <div className="col-6 modal-code-intellify small-hidden">
                                <CodeIntellifyBlob
                                    {...this.props}
                                    startLine={236}
                                    endLine={284}
                                    parentElement={'.modal-code-intellify'}
                                    className={'code-intellify-container-modal'}
                                    overlayPortal={this.overlayPortal}
                                    tooltipClass={'modal-tooltip'}
                                    defaultHoverPosition={defaultTooltipModalPosition}
                                />
                            </div>
                        </div>
                        <div className="row action-row">
                            <div className="col-12 action-col">
                                <p className="action-text">
                                    Get started with Sourcegraph for free, and get get cross-repository code
                                    intelligence, advanced code search, and extensive integrations.
                                </p>

                                <a className="action-link" href="https://docs.sourcegraph.com/#quickstart">
                                    Deploy Sourcegraph
                                    <ChevronRightIcon className="material-icons" />
                                </a>
                            </div>
                            <div className="col-12 action-col">
                                <p className="action-text">
                                    Explore how Sourcegraph's code intelligence can augment and add to your workflow,
                                    prepare you for code review, and speed through development.
                                </p>
                                <a
                                    className="action-link"
                                    href="//about.sourcegraph.com/docs/code-intelligence"
                                    target="_blank"
                                >
                                    Code intelligence documentation
                                    <ChevronRightIcon className="material-icons" />
                                </a>
                            </div>
                        </div>
                        <button className="btn-close-bottom" onClick={this.closeModal('intelligence')}>
                            Close <CloseIcon className="material-icons" />
                        </button>
                    </div>
                </div>
                <div
                    className={`modal-integrations ${this.state.modalIntegrationsOpen ? 'modal-open' : 'modal-close'} ${
                        this.state.modalIntegrationsClosing ? 'modal-closing' : ''
                    }`}
                >
                    <div className="container">
                        <button className="btn-close-top" onClick={this.closeModal('integrations')}>
                            <CloseIcon className="material-icons" />
                        </button>
                        <div className="row copy-section">
                            <div className="col-12 modal-header">
                                <h2>Integrations</h2>
                                <h1>Get it. Together.</h1>
                            </div>
                        </div>
                        <div className="row intelligence-row">
                            <div className="col-12 modal-header copy-section">
                                {integrationsSections.map(({ title, paragraph, buttons }, i) => (
                                    <div
                                        key={`search-sections-${i}`}
                                        className={`row copy-section modal-copy-row activesec`}
                                    >
                                        <div className="col-12">
                                            <h3>{title}</h3>
                                            <p>{paragraph}</p>
                                            {buttons.map(({ text, id, link }, j) => (
                                                <a
                                                    key={`integrations-buttons-${j}`}
                                                    className={`btn btn-secondary btn-integrations  ${id}`}
                                                    href={`${link}`}
                                                >
                                                    <span className="logo-icon" />
                                                    {text}
                                                </a>
                                            ))}
                                        </div>
                                    </div>
                                ))}
                            </div>
                            <div className="col-6 modal-code-intellify" />
                        </div>
                        <div className="row action-row">
                            <div className="col-12 action-col">
                                <p className="action-text">
                                    Get started with Sourcegraph for free, and get get cross-repository code
                                    intelligence, advanced code search, and extensive integrations.
                                </p>

                                <a className="action-link" href="https://docs.sourcegraph.com/#quickstart">
                                    Deploy Sourcegraph
                                    <ChevronRightIcon className="material-icons" />
                                </a>
                            </div>
                            <div className="col-12 action-col">
                                <p className="action-text">
                                    Explore all of Sourcegraph's integrations and see how you can get cross-repository
                                    code intelligence on your favorite code host and editor.
                                </p>
                                <a
                                    className="action-link"
                                    href="//about.sourcegraph.com/docs/integrations"
                                    target="_blank"
                                >
                                    Integrations documentation
                                    <ChevronRightIcon className="material-icons" />
                                </a>
                            </div>
                        </div>

                        <button className="btn-close-bottom" onClick={this.closeModal('integrations')}>
                            Close <CloseIcon className="material-icons" />
                        </button>
                    </div>
                </div>
            </div>
        )
    }

    private activateModal = (section: string) => () => {
        const windowBody = document.body
        windowBody.classList.add('modal-open')

        if (section === 'intelligence') {
            this.setState(state => ({ modalIntelligenceOpen: !state.modalIntelligenceOpen }))
            this.setState(state => ({ animateModalIntelligence: !state.animateModalIntelligence }))
            this.setState(() => ({ activesection: '99' }))
        } else if (section === 'integrations') {
            this.setState(state => ({ modalIntegrationsOpen: !state.modalIntegrationsOpen }))
            this.setState(state => ({ animateModalIntegrations: !state.animateModalIntegrations }))
            this.setState(() => ({ activesection: '99' }))
        }
    }

    private closeModal = (modalName: string) => () => {
        const windowBody = document.body
        windowBody.classList.remove('modal-open')

        if (modalName === 'intelligence') {
            this.setState(state => ({
                modalIntelligenceClosing: !state.modalIntelligenceClosing,
                animateModalIntelligence: false,
            }))
            // RESET DID CLOSE
            setTimeout(() => {
                this.setState(state => ({
                    modalIntelligenceOpen: !state.modalIntelligenceOpen,
                    modalIntelligenceClosing: !state.modalIntelligenceClosing,
                }))
            }, 400)
        } else if (modalName === 'integrations') {
            this.setState(state => ({
                modalIntegrationsClosing: !state.modalIntegrationsClosing,
                animateModalIntegrations: false,
            }))
            // RESET DID CLOSE
            setTimeout(() => {
                this.setState(state => ({
                    modalIntegrationsOpen: !state.modalIntegrationsOpen,
                    modalIntegrationsClosing: !state.modalIntegrationsClosing,
                }))
            }, 400)
        }
    }
}
