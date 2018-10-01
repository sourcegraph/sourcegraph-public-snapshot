import * as H from 'history'
import { throttle } from 'lodash'
import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'
import { parseSearchURLQuery } from '..'
import * as GQL from '../../backend/graphqlschema'
import { Form } from '../../components/Form'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { limitString } from '../../util'
import { queryIndexOfScope, submitSearch } from '../helpers'
import { QueryInputForModal } from './QueryInputForModal'

import { ExtensionsDocumentsProps } from '../../extensions/environment/ExtensionsEnvironment'
import { ExtensionsControllerProps } from '../../extensions/ExtensionsClientCommonContext'
import { CodeIntellifyBlob } from './CodeIntellifyBlob'
import { SearchButton } from './SearchButton'

interface Props extends ExtensionsControllerProps, ExtensionsDocumentsProps {
    user: GQL.IUser | null
    location: H.Location
    history: H.History
    isLightTheme: boolean
    onThemeChange: () => void
    onMainPage: (mainPage: boolean) => void
    showHelpPopover: boolean
    onHelpPopoverToggle: (visible?: boolean) => void
    handleScroll: () => void
}

interface State {
    /** The query value entered by the user in the query input */
    userQuery: string
    modalSearchOpen: boolean
    modalIntelligenceOpen: boolean
    modalSearchClosing: boolean
    modalIntelligenceClosing: boolean
    /** Unique ID for keeping selected button state. */
    activeButton?: string
    activeSection?: string
    animateModalSearch: boolean
    animateModalIntelligence: boolean
    manualClick?: boolean
    bgScrollStyle: string
}
const heroEyebrow = 'Sourcegraph'
const heroTitle = 'Open. For business.'
const heroCopyTop =
    'Sourcegraph is a free, open-source, self-hosted code search and intelligence server that helps developers find, review, understand, and debug code. Use it with any Git code host for teams of any size. Try it out on then install the Sourcegraph docker image on your private code. Upgraded features start at $4/user/month.'
const heroCopyBottom =
    'Get started with powerful search, advanced code intelligence, and vast integrations into your code host and browser.'

const searchSections = [
    {
        title: 'Powerful, Flexible Queries',
        paragraph:
            'Sourcegraph code search performs full-text searches and supports both regular expression and exact queries. By default, Sourcegraph searches across all your repositories. Our search query syntax allows for advanced queries, such as searching over any branch or commit, narrowing searches by programming language or file pattern, and more.',
        buttons: [
            { query: 'github.com', text: 'GitHub DEMO' },
            { query: 'repogroup:goteam file:\\.go$', text: 'Go Code by Go Team' },
            {
                query: 'repogroup:ethereum file:\\.(txt|md)$ file:(test|spec) ',
                text: 'Core Etherium Test Files',
                onClick: 'buttonGoLang',
            },
            { query: 'repogroup:angular file:\\.JSON$', text: 'Angular JSON Files' },
        ],
    },
    {
        title: 'Commit Diff Search',
        paragraph:
            'Search over commit diffs using type:diff to see how your codebase has changed over time. This is often used to find changes to particular functions, classes, or areas of the codebase when debugging.<br><br>You can also search within commit diffs on multiple branches by specifying the branches in a repo: field after the @ sign. After the @, separate Git refs with :, specify Git ref globs by prefixing them with *, and exclude commits reachable from a ref by prefixing it with ^.',
        buttons: [
            { query: 'repo:^github\\.com/apple/swift$@master-next type:diff', text: 'Swift Diff on Master-Next' },
            { query: 'repogroup:goteam file:\\.go$', text: 'ReactJS Test File Diff' },
            {
                query: 'repogroup:ethereum file:\\.(txt|md)$ file:(test|spec) ',
                text: 'Core Etherium Test Files',
                onClick: 'buttonGoLang',
            },
            { query: 'repogroup:angular file:\\.JSON$', text: 'Angular JSON Files' },
        ],
    },
    {
        title: 'Commit Message Search',
        paragraph:
            'Searching over commit messages is supported in Sourcegraph by adding type:commit to your search query. Separately, you can also use the message:"any string" token to filter type:diff searches for a given commit message.',
        buttons: [
            { query: 'repogroup:goteam', text: 'Go Team Code' },
            { query: 'repogroup:goteam file:\\.go$', text: 'Go Code by Go Team' },
            {
                query: 'repogroup:ethereum file:\\.(txt|md)$ file:(test|spec) ',
                text: 'Core Etherium Test Files',
                onClick: 'buttonGoLang',
            },
            { query: 'repogroup:angular file:\\.JSON$', text: 'Angular JSON Files' },
        ],
    },
    {
        title: 'Symbol Search',
        paragraph:
            'Searching for symbols makes it easier to find specific functions, variables and more. Use the type:symbol filter to search for symbol results. Symbol results also appear in typeahead suggestions, so you can jump directly to symbols by name.',
        buttons: [
            { query: 'repogroup:goteam', text: 'Go Team Code' },
            { query: 'repogroup:goteam file:\\.go$', text: 'Go Code by Go Team' },
            {
                query: 'repogroup:ethereum file:\\.(txt|md)$ file:(test|spec) ',
                text: 'Core Etherium Test Files',
                onClick: 'buttonGoLang',
            },
            { query: 'repogroup:angular file:\\.JSON$', text: 'Angular JSON Files' },
        ],
    },
]
const intelligenceSections = [
    {
        title: 'Code browsing',
        paragraph:
            'View open source code, like gorilla/mux, on sourcegraph.com, or deploy your own instance to see public code alongside your private code. See how your codebase changes over time in functions, classes, or areas of the codebase when debugging.',
    },
    {
        title: 'Advanced code intelligence',
        paragraph:
            'Code intelligence makes browsing code easier, with IDE-like hovers, go-to-definition, and find-references on your code, powered by language servers based on the open-source Language Server Protocol.',
    },
    {
        title: 'Hover tooltip',
        paragraph:
            "Use the hovertooltip to discover and understand your code faster. Click on a token and then go to it's definition, other refrences, or implementations. Speed through reviews by understanding new code and changed code and what it affects.",
    },
    {
        title: '',
        paragraph:
            'Code intelligence is powered by language servers based on the open-standard Language Server Protocol (published by Microsoft, with participation from Facebook, Google, Sourcegraph, GitHub, RedHat, Twitter, Salesforce, Eclipse, and others). Visit langserver.org to learn more about the Language Server Protocol, find the latest support for your favorite language, and get involved.',
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
        overflow: scroll !important;
        max-height: 200px;
    }
    .modal-tooltip {
        z-index: 9999 !important;
        opacity: 1 !important;
        visbility: visbily !important;
    }
`
/**
 * The search page
 */
export class MainPage extends React.Component<Props, State> {
    private static HIDE_REPOGROUP_SAMPLE_STORAGE_KEY = 'MainPage/hideRepogroupSample'

    private overlayPortal: HTMLElement | undefined

    constructor(props: Props) {
        super(props)

        const searchOptions = parseSearchURLQuery(props.location.search)

        this.state = {
            userQuery: (searchOptions && searchOptions.query) || '',
            modalSearchOpen: false,
            modalSearchClosing: false,
            modalIntelligenceOpen: false,
            modalIntelligenceClosing: false,
            manualClick: false,
            activeSection: 'fadwsghjk',
            animateModalSearch: false,
            animateModalIntelligence: false,
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
        this.props.onMainPage(true)

        window.addEventListener('scroll', throttle(this.handleScroll, 50))
    }

    public componentWillUnmount(): void {
        this.props.onMainPage(false)
    }

    private handleScroll = (event: any): void => {
        // const scrollPos = window.scrollY
        // const sqrtPos = Math.sqrt(scrollPos) * 2
        // let opacityChange = sqrtPos / 50
        // if (opacityChange <= 0.5) {
        //     opacityChange = 0.5
        // } else if (opacityChange >= 1) {
        //     opacityChange = 1
        // } else if (isNaN(opacityChange)) {
        //     opacityChange = 0.5
        // }
        // console.log(scrollPos)
        // console.log(sqrtPos)
        // console.log(opacityChange)
        // this.setState({ bgScrollStyle })
    }
    public render(): JSX.Element | null {
        return (
            <div className="main-page">
                <style>{inlineStyle}</style>
                <style>{this.state.bgScrollStyle}</style>

                <PageTitle title={this.getPageTitle()} />
                <section className="hero-section">
                    <div className="hero-section__bg" />
                    <div className=" container hero-container">
                        <div className="row">
                            <div className="col-6">
                                <h2>{heroEyebrow}</h2>
                                <h1>{heroTitle}</h1>
                                <p>{heroCopyTop}</p>
                                <button className="btn btn-primary">Deploy Sourcegraph</button>
                                <button className="btn btn-secondary">Sourcegraph GitHub</button>
                            </div>
                            <div className="col-6">
                                <CodeIntellifyBlob
                                    {...this.props}
                                    startLine={236}
                                    endLine={250}
                                    parentElement={'.hero-section'}
                                    containerClass={'code-intellify-container'}
                                    overlayPortal={this.overlayPortal}
                                    tooltipClass={'hero-tooltip'}
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
                                        className="btn btn-secondary"
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
                                        find-references on your code, powered by language servers based on the
                                        open-source Language Server Protocol.
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
                                    <h1>Get it. Together</h1>
                                    <p>
                                        Connect your Sourcegraph instance with your existing tools. Get code
                                        intelligence while browsing code on the web, and code search from your editor.
                                    </p>
                                    <button className="btn btn-secondary">Explore Integrations</button>
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
                                <button className="btn btn-primary">Deploy Sourcegraph</button>
                                <button className="btn btn-secondary">Sourcegraph GitHub</button>
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
                                <button className="btn btn-secondary">Sourcegraph Pricing</button>
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
                                        <a href="/blog">Blog</a>
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
                                        <a href="/plan">Master Plan</a>
                                    </li>
                                    <li>
                                        <a href="/about">About</a>
                                    </li>
                                    <li>
                                        <a href="/contact">Contact</a>
                                    </li>
                                    <li>
                                        <a href="/jobs">Careers</a>
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
                                        <a href="/#search-datacenter">Code Search</a>
                                    </li>
                                    <li>
                                        <a href="/#intelligence-ntegrations">Code Intelligence</a>
                                    </li>
                                    <li>
                                        <a href="/#search-datacenter">Data Center</a>
                                    </li>
                                    <li>
                                        <a href="/#intelligence-integrations">Integrations</a>
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
                                        <a href="/docs">Documentation</a>
                                    </li>
                                    <li>
                                        <a href="/changelog">Changelog</a>
                                    </li>
                                    <li>
                                        <a href="/pricing">Pricing</a>
                                    </li>
                                    <li>
                                        <a href="/security">Security</a>
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
                                    <a href="/terms">Terms</a>
                                    <a href="/privacy">Privacy</a>
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
                    <div className="container">
                        <div className="row copy-section">
                            <div className="col-12 modal-header">
                                <h2>Advanced Code Search</h2>
                                <h1>Find. Then replace.</h1>
                            </div>
                        </div>
                        <div
                            className={`search-row copy-section ${
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
                                    this.state.activeSection === `${i}` || this.state.activeSection === '99'
                                        ? 'activeSec'
                                        : ''
                                }`}
                            >
                                <div className="col-12">
                                    <h3>{title}</h3>
                                    <p>{paragraph}</p>
                                    {buttons.map(({ onClick, text, query }, j) => (
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
                        <button className="btn" onClick={this.closeModal('search')}>
                            Close
                        </button>
                    </div>
                </div>
                <div
                    className={`modal-intelligence ${this.state.modalIntelligenceOpen ? 'modal-open' : 'modal-close'} ${
                        this.state.modalIntelligenceClosing ? 'modal-closing' : ''
                    }`}
                >
                    <div className="container">
                        <div className="row copy-section">
                            <div className="col-12 modal-header">
                                <h2>Enhanced Code Browsing and Intelligence</h2>
                                <h1>Mine your language.</h1>
                            </div>
                        </div>
                        <div className="row intelligence-row">
                            <div className="col-6 modal-header copy-section">
                                {intelligenceSections.map(({ title, paragraph }, i) => (
                                    <div
                                        key={`search-sections-${i}`}
                                        className={`row modal-copy-row ${
                                            this.state.activeSection === `${i}` || this.state.activeSection === '99'
                                                ? 'activeSec'
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
                            <div className="col-6 modal-code-intellify">
                                <CodeIntellifyBlob
                                    {...this.props}
                                    startLine={236}
                                    endLine={288}
                                    parentElement={'.modal-code-intellify'}
                                    containerClass={'code-intellify-container-modal'}
                                    overlayPortal={this.overlayPortal}
                                    tooltipClass={'modal-tooltip'}
                                />
                            </div>
                        </div>
                        <button className="btn" onClick={this.closeModal('intelligence')}>
                            Close
                        </button>
                    </div>
                </div>
            </div>
        )
    }

    private handleQueryChange = (activeButton: string, passedQ: string) => () => {
        if (passedQ === '') {
            this.setState({ activeButton, manualClick: true, activeSection: '99' })
        } else {
            this.setState({ activeButton, userQuery: passedQ, manualClick: true, activeSection: '99' })
        }
    }

    private activateModal = (section: string) => () => {
        const windowBody = document.querySelector('body')
        if (windowBody) {
            windowBody.classList.add('modal-open')
        }

        if (section === 'intelligence') {
            this.setState(state => ({ modalIntelligenceOpen: !state.modalIntelligenceOpen }))
            this.setState(state => ({ animateModalIntelligence: !state.animateModalIntelligence }))
            this.setState(state => ({ activeSection: '99' }))
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
                        activeSection: '0',
                        userQuery: 'repogroup:goteam file:\\.go$',
                    }))
                }
            }, 1450)
            setTimeout(() => {
                if (this.state.manualClick === false) {
                    this.setState(state => ({
                        activeButton: '1-0',
                        activeSection: '1',

                        userQuery: 'repo:^github\\.com/apple/swift$@master-next type:diff',
                    }))
                }
            }, 7450)
            setTimeout(() => {
                if (this.state.manualClick === false) {
                    this.setState(state => ({
                        activeButton: '2-3',
                        activeSection: '2',
                        userQuery: 'repogroup:goteam file:\\.go$',
                    }))
                }
            }, 15550)
            setTimeout(() => {
                if (this.state.manualClick === false) {
                    this.setState(state => ({
                        activeButton: '3-2',
                        activeSection: '3',
                        userQuery: 'repogroup:goteam file:\\.go$',
                    }))
                }
            }, 21450)
            setTimeout(() => {
                if (this.state.manualClick === false) {
                    this.setState(state => ({
                        activeButton: '0-0',
                        activeSection: '99',
                        userQuery: 'repogroup:goteam file:\\.go$',
                    }))
                }
            }, 28450)
        }
    }

    private closeModal = (modalName: string) => () => {
        const windowBody = document.querySelector('body')
        if (windowBody) {
            windowBody.classList.remove('modal-open')
        }
        if (modalName === 'search') {
            this.setState(state => ({ modalSearchClosing: !state.modalSearchClosing }))
            this.setState(state => ({ animateModalSearch: false }))
            // RESET DID CLOSE
            setTimeout(() => {
                this.setState(state => ({ modalSearchOpen: !state.modalSearchOpen }))
                this.setState(state => ({ modalSearchClosing: !state.modalSearchClosing }))
            }, 400)
        } else if (modalName === 'intelligence') {
            this.setState(state => ({ modalIntelligenceClosing: !state.modalIntelligenceClosing }))
            this.setState(state => ({ animateModalIntelligence: false }))
            // RESET DID CLOSE
            setTimeout(() => {
                this.setState(state => ({ modalIntelligenceOpen: !state.modalIntelligenceOpen }))
                this.setState(state => ({ modalIntelligenceClosing: !state.modalIntelligenceClosing }))
            }, 400)
        } else if (modalName === 'integrations') {
            // Add int code here
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
        submitSearch(this.props.history, { query: this.state.userQuery }, 'home')
    }

    private getPageTitle(): string | undefined {
        const options = parseSearchURLQuery(this.props.location.search)
        if (options && options.query) {
            return `${limitString(this.state.userQuery, 25, true)}`
        }
        return undefined
    }
}
