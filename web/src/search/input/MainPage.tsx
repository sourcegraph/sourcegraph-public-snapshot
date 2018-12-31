import * as H from 'history'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import * as React from 'react'
import { parseSearchURLQuery } from '..'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { Form } from '../../components/Form'
import { HeroPage } from '../../components/HeroPage'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { limitString } from '../../util'
import { queryIndexOfScope, submitSearch } from '../helpers'
import { CodeIntellifyBlob } from './CodeIntellifyBlob'
import { QueryInputForModal } from './QueryInputForModal'
import { SearchButton } from './SearchButton'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    authenticatedUser: GQL.IUser | null
    location: H.Location
    history: H.History
    isLightTheme: boolean
    onThemeChange: () => void
    onMainPage: (mainPage: boolean) => void
}

interface State {
    /** The query value entered by the user in the query input */
    userQuery: string
    // modalXXopen sets a state that the modal is open before animations or closed after animation
    // modalXXclosing sets a state that starts the closing process
    modalSearchOpen: boolean
    modalIntelligenceOpen: boolean
    modalSearchClosing: boolean
    modalIntelligenceClosing: boolean
    modalIntegrationsOpen: boolean
    modalIntegrationsClosing: boolean
    // determine what button in modal is active
    activeButton?: string
    // determine what section inside a modal is active
    activesection?: string
    // animateModalXX starts the animation process after opening.
    animateModalSearch: boolean
    animateModalIntelligence: boolean
    animateModalIntegrations: boolean
    // Manual click state is to determine if animation should be stopped
    manualClick?: boolean
    bgScrollStyle: string
}
const heroEyebrow = 'Sourcegraph'
const heroTitle = 'Open. For business.'
const heroCopyTop =
    'Sourcegraph is a free, open-source, self-hosted code search and intelligence server that helps developers find, review, understand, and debug code. Use it with any Git code host for teams of any size. Start using it now, then install the Sourcegraph Docker image on your private code.'
const heroCopyBottom = 'Upgraded features start at $4/user/month.'

const searchSections = [
    {
        title: 'Powerful, Flexible Queries',
        paragraph:
            'Sourcegraph code search performs full-text searches and supports both regular expression and exact queries. By default, Sourcegraph searches across all your repositories. Our search query syntax allows for advanced queries, such as searching over any branch or commit, narrowing searches by programming language or file pattern, and more.',
        buttons: [
            { query: 'repo:^gitlab.com/ ', text: 'Code on GitLab' },
            { query: 'repogroup:goteam file:\\.go$', text: 'Go Code by Go Team' },
            {
                query: 'repogroup:ethereum file:\\.(txt|md)$ file:(test|spec) ',
                text: 'Core Etherium Test Files',
            },
            { query: 'repogroup:angular file:\\.JSON$', text: 'Angular JSON Files' },
        ],
    },
    {
        title: 'Commit Diff Search',
        paragraph:
            'Search over commit diffs using type:diff to see how your codebase has changed over time. This is often used to find changes to particular functions, classes, or areas of the codebase when debugging. You can also search within commit diffs on multiple branches by specifying the branches in a repo: field after the @ sign.',
        buttons: [
            { query: 'repo:^github\\.com/apple/swift$@master-next type:diff', text: 'Swift Diff on Master-Next' },
            { query: 'repo:^github.com/facebook/react$ file:(test|spec)  type:diff', text: 'ReactJS Test File Diff' },
            {
                query: 'repo:^github.com/golang/oauth2$ type:diff',
                text: 'Go oAuth2 Diff',
            },
            {
                query: 'repo:^github.com/kubernetes/kubernetes$ type:diff statefulset',
                text: 'Kubernetes Statefulset Diff',
            },
        ],
    },
    {
        title: 'Commit Message Search',
        paragraph:
            'Searching over commit messages is supported in Sourcegraph by adding type:commit to your search query. Separately, you can also use the message:"any string" token to filter type:diff searches for a given commit message.',
        buttons: [
            { query: 'type:commit  repogroup:angular author:google.com>$ ', text: 'Angular Commits by Googlers' },
            { query: 'repogroup:npm type:commit security', text: 'NPM Commits mentioning Security' },
            {
                query: 'repogroup:ethereum type:commit money loss',
                text: "Ethereum Commits mentioning 'Money Loss'",
            },
            { query: 'repo:^github.com/sourcegraph/sourcegraph type:commit', text: 'Sourcegraph Commits' },
        ],
    },
    {
        title: 'Symbol Search',
        paragraph:
            'Searching for symbols makes it easier to find specific functions, variables and more. Use the type:symbol filter to search for symbol results. Symbol results also appear in typeahead suggestions, so you can jump directly to symbols by name.',
        buttons: [
            { query: 'repogroup:goteam type:symbol httpRouter', text: 'Go Code with httpRouter' },
            { query: 'repo:^github.com/apple/swift$ type:symbol main', text: "Calls to 'main' in Swift" },
            {
                query: 'repo:golang/go$ type:symbol sprintF',
                text: 'Golang sprintF',
            },
            { query: 'repo:golang/go$ type:symbol newDecoder', text: 'Golang newDecoder' },
        ],
    },
]
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
        title: 'IDE extensions',
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
                text: 'VS Code',
                link: 'https://marketplace.visualstudio.com/items?itemName=sourcegraph.sourcegraph',
            },
        ],
    },
]

const inlineStyle = `
    .layout {
        display: block !important;
    }
    * {
        overflow: visible !important;
    }
    .hero-tooltip {
        z-index: 1;
        position: fixed !important;
        transform: translateY(44px);
    }
    .hover-overlay__contents {
        overflow-y: auto !important;
        max-height: 200px;
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
    private static HIDE_REPOGROUP_SAMPLE_STORAGE_KEY = 'MainPage/hideRepogroupSample'

    private overlayPortal: HTMLElement | undefined

    constructor(props: Props) {
        super(props)

        const query = parseSearchURLQuery(props.location.search)

        this.state = {
            userQuery: query || '',
            modalSearchOpen: false,
            modalSearchClosing: false,
            modalIntelligenceOpen: false,
            modalIntelligenceClosing: false,
            modalIntegrationsOpen: false,
            modalIntegrationsClosing: false,
            manualClick: false,
            activesection: 'none',
            animateModalSearch: false,
            animateModalIntelligence: false,
            animateModalIntegrations: false,
            bgScrollStyle: `
            .feature-card {
                transform: translateY(-0px);
                opacity: .5;
            }
        `,
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('Home')
        if (
            window.context.sourcegraphDotComMode &&
            !localStorage.getItem(MainPage.HIDE_REPOGROUP_SAMPLE_STORAGE_KEY) &&
            !this.state.userQuery
        ) {
            this.setState({ userQuery: 'repogroup:sample' })
        }

        const portal = document.createElement('div')
        document.body.appendChild(portal)
        this.overlayPortal = portal

        // communicate onMainPage to SourcegraphWebApp for dark theme look
        if (window.context.sourcegraphDotComMode) {
            this.props.onMainPage(true)
        }

        // Add class to body to prevent global element styles from affecting other pages
        const windowBody = document.body
        windowBody.classList.add('main-page')
    }

    public componentWillUnmount(): void {
        this.props.onMainPage(false)
        const windowBody = document.body
        windowBody.classList.remove('main-page')
        windowBody.classList.remove('modal-open')
    }
    public render(): JSX.Element | null {
        if (!window.context.sourcegraphDotComMode) {
            return <HeroPage icon={MapSearchIcon} title="Page not found" />
        }
        return (
            <div className="main-page">
                <style>{inlineStyle}</style>
                <style>{this.state.bgScrollStyle}</style>

                <PageTitle title={this.getPageTitle()} />
                <section className="hero-section">
                    <div className="hero-section__bg" />
                    <div className=" container hero-container">
                        <div className="row">
                            <div className="col-lg-6 col-md-12 col-sm-12">
                                <h2>{heroEyebrow}</h2>
                                <h1>{heroTitle}</h1>
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
                    <section className="about-section code-search">
                        <div className="container">
                            <div className="row">
                                <div className="col-12">
                                    <h2>Advanced Code Search</h2>
                                    <h1>Find. Then replace.</h1>
                                    <p>
                                        Search in files and diffs on your private code using simple terms, regular
                                        expressions, and other filters.
                                    </p>
                                    <p>
                                        Syncs repositories with your code host and supports searching any commit/branch,
                                        with no indexing delay.
                                    </p>
                                    <button
                                        className="btn btn-secondary "
                                        id="sampleButton"
                                        onClick={this.activateModal('search')}
                                    >
                                        Explore Code Search
                                    </button>
                                </div>
                            </div>
                        </div>
                    </section>
                    <section className="about-section code-intelligence">
                        <div className="container">
                            <div className="row">
                                <div className="col-12">
                                    <h2>Enhanced Code Browsing and Intelligence</h2>
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
                                        Explore Code Intelligence
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
                                        Explore Integrations
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
                                <h2>Sourcegraph Pricing</h2>
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
                                        <a onClick={this.activateModal('search')}>Code Search</a>
                                    </li>
                                    <li>
                                        <a onClick={this.activateModal('intelligence')}>Code Intelligence</a>
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
                                        <a href="//about.sourcegraph.com/changelog">Changelog</a>
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
                    className={`modal-search ${this.state.modalSearchOpen ? 'modal-open' : 'modal-close'} ${
                        this.state.modalSearchClosing ? 'modal-closing' : ''
                    }`}
                >
                    <div className="container search-modal-container">
                        <button className="btn-close-top" onClick={this.closeModal('search')}>
                            <CloseIcon className="material-icons" />
                        </button>
                        <div className="row copy-section">
                            <div className="col-12 modal-header">
                                <h2>Advanced Code Search</h2>
                                <h1>Find. Then replace.</h1>
                            </div>
                        </div>
                        <div
                            className={`search-row copy-section  ${
                                this.state.animateModalSearch ? 'search-row-animate' : ''
                            }`}
                        >
                            <Form className="search main-page__container" onSubmit={this.onSubmit}>
                                <div
                                    className="main-page__input-container"
                                    onClick={this.handleQueryChange('0', '')}
                                    onKeyPress={this.handleQueryChange('0', '')}
                                >
                                    <QueryInputForModal
                                        {...this.props}
                                        value={this.state.userQuery}
                                        onChange={this.onUserQueryChange}
                                        autoFocus={'cursor-at-end'}
                                        hasGlobalQueryBehavior={true}
                                    />
                                    <SearchButton />
                                </div>
                            </Form>
                        </div>

                        {searchSections.map(({ title, paragraph, buttons }, i) => (
                            <div
                                key={`search-sections-${i}`}
                                className={`row copy-section modal-copy-row ${
                                    this.state.activesection === `${i}` || this.state.activesection === '99'
                                        ? 'activesec'
                                        : ''
                                }`}
                            >
                                <div className="col-12">
                                    <h3>{title}</h3>
                                    <p>{paragraph}</p>
                                    {buttons.map(({ text, query }, j) => (
                                        <button
                                            key={`search-buttons-${j}`}
                                            className={`btn btn-secondary ${
                                                this.state.activeButton === `${i}-${j}` ? 'active' : ''
                                            }`}
                                            onClick={this.handleQueryChange(`${i}-${j}`, query)}
                                        >
                                            {text}
                                        </button>
                                    ))}
                                </div>
                            </div>
                        ))}
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
                                    Explore the power and extensibility of Sourcegraph's search query syntax, and learn
                                    how to search across all of your public and private code.
                                </p>
                                <a
                                    className="action-link"
                                    href="//about.sourcegraph.com/docs/search/query-syntax"
                                    target="_blank"
                                >
                                    Search Documentation
                                    <ChevronRightIcon className="material-icons" />
                                </a>
                            </div>
                        </div>
                        <button className="btn-close-bottom" onClick={this.closeModal('search')}>
                            Close <CloseIcon className="material-icons" />
                        </button>
                    </div>
                </div>
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
                                <h2>Enhanced Code Browsing and Intelligence</h2>
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
                                    Code Intelligence Documentation
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
                                    Integrations Documentation
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

    private handleQueryChange = (activeButton: string, passedQ: string) => () => {
        if (passedQ === '') {
            this.setState({ activeButton, manualClick: true, activesection: '99' })
        } else {
            this.setState({ activeButton, userQuery: passedQ, manualClick: true, activesection: '99' })
        }
    }

    private activateModal = (section: string) => () => {
        const windowBody = document.body
        windowBody.classList.add('modal-open')

        if (section === 'intelligence') {
            this.setState(state => ({ modalIntelligenceOpen: !state.modalIntelligenceOpen }))
            this.setState(state => ({ animateModalIntelligence: !state.animateModalIntelligence }))
            this.setState(state => ({ activesection: '99' }))
        } else if (section === 'integrations') {
            this.setState(state => ({ modalIntegrationsOpen: !state.modalIntegrationsOpen }))
            this.setState(state => ({ animateModalIntegrations: !state.animateModalIntegrations }))
            this.setState(state => ({ activesection: '99' }))
        } else if (section === 'search') {
            this.setState(state => ({ modalSearchOpen: !state.modalSearchOpen }))
            this.setState(state => ({ animateModalSearch: !state.modalSearchOpen }))
            setTimeout(() => {
                if (this.state.animateModalSearch) {
                    this.setState(state => ({ animateModalSearch: false }))
                }
            }, 750)
            setTimeout(() => {
                if (this.state.manualClick === false) {
                    this.setState(state => ({
                        activeButton: '0-1',
                        activesection: '0',
                        userQuery: 'repogroup:goteam file:\\.go$',
                    }))
                }
            }, 1450)
            setTimeout(() => {
                if (this.state.manualClick === false) {
                    this.setState(state => ({
                        activeButton: '1-0',
                        activesection: '1',

                        userQuery: 'repo:^github\\.com/apple/swift$@master-next type:diff',
                    }))
                }
            }, 7450)
            setTimeout(() => {
                if (this.state.manualClick === false) {
                    this.setState(state => ({
                        activeButton: '2-3',
                        activesection: '2',
                        userQuery: 'repogroup:goteam file:\\.go$',
                    }))
                }
            }, 15550)
            setTimeout(() => {
                if (this.state.manualClick === false) {
                    this.setState(state => ({
                        activeButton: '3-2',
                        activesection: '3',
                        userQuery: 'repogroup:goteam file:\\.go$',
                    }))
                }
            }, 21450)
            setTimeout(() => {
                if (this.state.manualClick === false) {
                    this.setState(state => ({
                        activeButton: '0-0',
                        activesection: '99',
                        userQuery: 'repogroup:goteam file:\\.go$',
                    }))
                }
            }, 28450)
        }
    }

    private closeModal = (modalName: string) => () => {
        const windowBody = document.body
        windowBody.classList.remove('modal-open')

        if (modalName === 'search') {
            this.setState(state => ({ modalSearchClosing: !state.modalSearchClosing, animateModalSearch: false }))
            // RESET DID CLOSE
            setTimeout(() => {
                this.setState(state => ({ modalSearchOpen: !state.modalSearchOpen }))
                this.setState(state => ({ modalSearchClosing: !state.modalSearchClosing }))
            }, 400)
        } else if (modalName === 'intelligence') {
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

    private onUserQueryChange = (userQuery: string) => {
        this.setState({ userQuery })

        if (window.context.sourcegraphDotComMode) {
            if (queryIndexOfScope(userQuery, 'repogroup:sample') !== -1) {
                localStorage.removeItem(MainPage.HIDE_REPOGROUP_SAMPLE_STORAGE_KEY)
            } else {
                localStorage.setItem(MainPage.HIDE_REPOGROUP_SAMPLE_STORAGE_KEY, 'true')
            }
        }
    }

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        submitSearch(this.props.history, this.state.userQuery, 'home')
    }

    private getPageTitle(): string | undefined {
        const query = parseSearchURLQuery(this.props.location.search)
        if (query) {
            return `${limitString(this.state.userQuery, 25, true)}`
        }
        return undefined
    }
}
