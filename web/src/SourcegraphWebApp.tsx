// This gets automatically expanded into
// imports that only pick what we need
import '@babel/polyfill'

// Polyfill URL because Chrome and Firefox are not spec-compliant
// Hostnames of URIs with custom schemes (e.g. git) are not parsed out
import { URL, URLSearchParams } from 'whatwg-url'
// The polyfill does not expose createObjectURL, which we need for creating data: URIs for Web
// Workers. So retain it.
URL.createObjectURL = window.URL.createObjectURL
Object.assign(window, { URL, URLSearchParams })

// Load only a subset of the highlight.js languages
import { registerLanguage } from 'highlight.js/lib/highlight'
registerLanguage('go', require('highlight.js/lib/languages/go'))
registerLanguage('javascript', require('highlight.js/lib/languages/javascript'))
registerLanguage('typescript', require('highlight.js/lib/languages/typescript'))
registerLanguage('java', require('highlight.js/lib/languages/java'))
registerLanguage('python', require('highlight.js/lib/languages/python'))
registerLanguage('php', require('highlight.js/lib/languages/php'))
registerLanguage('bash', require('highlight.js/lib/languages/bash'))
registerLanguage('clojure', require('highlight.js/lib/languages/clojure'))
registerLanguage('cpp', require('highlight.js/lib/languages/cpp'))
registerLanguage('cs', require('highlight.js/lib/languages/cs'))
registerLanguage('css', require('highlight.js/lib/languages/css'))
registerLanguage('dockerfile', require('highlight.js/lib/languages/dockerfile'))
registerLanguage('elixir', require('highlight.js/lib/languages/elixir'))
registerLanguage('haskell', require('highlight.js/lib/languages/haskell'))
registerLanguage('html', require('highlight.js/lib/languages/xml'))
registerLanguage('lua', require('highlight.js/lib/languages/lua'))
registerLanguage('ocaml', require('highlight.js/lib/languages/ocaml'))
registerLanguage('r', require('highlight.js/lib/languages/r'))
registerLanguage('ruby', require('highlight.js/lib/languages/ruby'))
registerLanguage('rust', require('highlight.js/lib/languages/rust'))
registerLanguage('swift', require('highlight.js/lib/languages/swift'))

import { Notifications } from '@sourcegraph/extensions-client-common/lib/app/notifications/Notifications'
import { createController as createExtensionsController } from '@sourcegraph/extensions-client-common/lib/client/controller'
import { ConfiguredExtension } from '@sourcegraph/extensions-client-common/lib/extensions/extension'
import {
    ConfigurationCascadeOrError,
    ConfigurationSubject,
    ConfiguredSubject,
    Settings,
} from '@sourcegraph/extensions-client-common/lib/settings'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import ServerIcon from '@sourcegraph/icons/lib/Server'
import * as React from 'react'
import { Route, RouteComponentProps } from 'react-router'
import { BrowserRouter } from 'react-router-dom'
import { Subscription } from 'rxjs'
import {
    Component as ExtensionsComponent,
    EMPTY_ENVIRONMENT as EXTENSIONS_EMPTY_ENVIRONMENT,
} from 'sourcegraph/module/environment/environment'
import { URI } from 'sourcegraph/module/types/textDocument'
import { currentUser } from './auth'
import * as GQL from './backend/graphqlschema'
import { FeedbackText } from './components/FeedbackText'
import { HeroPage } from './components/HeroPage'
import { Tooltip } from './components/tooltip/Tooltip'
import { ExtensionsEnvironmentProps, USE_PLATFORM } from './extensions/environment/ExtensionsEnvironment'
import {
    ConfigurationCascadeProps,
    createMessageTransports,
    ExtensionsControllerProps,
    ExtensionsProps,
} from './extensions/ExtensionsClientCommonContext'
import { createExtensionsContextController } from './extensions/ExtensionsClientCommonContext'
import { Layout, LayoutProps } from './Layout'
import { updateUserSessionStores } from './marketing/util'
import { eventLogger } from './tracking/eventLogger'
import { isErrorLike } from './util/errors'

interface AppState
    extends ConfigurationCascadeProps,
        ExtensionsProps,
        ExtensionsEnvironmentProps,
        ExtensionsControllerProps {
    error?: Error
    user?: GQL.IUser | null

    /**
     * Whether the light theme is enabled or not
     */
    isLightTheme: boolean

    /**
     * The current search query in the navbar.
     */
    navbarSearchQuery: string

    /** Whether the help popover is shown. */
    showHelpPopover: boolean

    /** Whether the history popover is shown. */
    showHistoryPopover: boolean
}

const LIGHT_THEME_LOCAL_STORAGE_KEY = 'light-theme'

/** A fallback configuration subject that can be constructed synchronously at initialization time. */
const SITE_SUBJECT_NO_ADMIN: Pick<GQL.IConfigurationSubject, 'id' | 'viewerCanAdminister'> = {
    id: window.context.siteGQLID,
    viewerCanAdminister: false,
}

/**
 * The root component
 */
export class SourcegraphWebApp extends React.Component<{}, AppState> {
    constructor(props: {}) {
        super(props)
        const extensions = createExtensionsContextController()
        this.state = {
            isLightTheme: localStorage.getItem(LIGHT_THEME_LOCAL_STORAGE_KEY) !== 'false',
            navbarSearchQuery: '',
            showHelpPopover: false,
            showHistoryPopover: false,
            configurationCascade: { subjects: null, merged: null },
            extensions,
            extensionsEnvironment: EXTENSIONS_EMPTY_ENVIRONMENT,
            extensionsController: createExtensionsController(extensions.context, createMessageTransports),
        }
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        updateUserSessionStores()

        document.body.classList.add('theme')
        this.subscriptions.add(
            currentUser.subscribe(user => this.setState({ user }), () => this.setState({ user: null }))
        )

        if (USE_PLATFORM) {
            this.subscriptions.add(this.state.extensionsController)

            this.subscriptions.add(
                this.state.extensions.context.configurationCascade.subscribe(
                    v => this.onConfigurationCascadeChange(v),
                    err => console.error(err)
                )
            )

            // Keep the Sourcegraph extensions controller's extensions up-to-date.
            //
            // TODO(sqs): handle loading and errors
            this.subscriptions.add(
                this.state.extensions.viewerConfiguredExtensions.subscribe(
                    extensions => this.onViewerConfiguredExtensionsChange(extensions),
                    err => console.error(err)
                )
            )
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
        document.body.classList.remove('theme')
        document.body.classList.remove('theme-light')
        document.body.classList.remove('theme-dark')
    }

    public componentDidUpdate(): void {
        localStorage.setItem(LIGHT_THEME_LOCAL_STORAGE_KEY, this.state.isLightTheme + '')
        document.body.classList.toggle('theme-light', this.state.isLightTheme)
        document.body.classList.toggle('theme-dark', !this.state.isLightTheme)
    }

    public render(): React.ReactFragment | null {
        if (this.state.error) {
            return <HeroPage icon={ErrorIcon} title={'Something happened'} subtitle={this.state.error.message} />
        }

        if (window.pageError && window.pageError.statusCode !== 404) {
            const statusCode = window.pageError.statusCode
            const statusText = window.pageError.statusText
            const errorMessage = window.pageError.error
            const errorID = window.pageError.errorID

            let subtitle: JSX.Element | undefined
            if (errorID) {
                subtitle = <FeedbackText headerText="Sorry, there's been a problem." />
            }
            if (errorMessage) {
                subtitle = (
                    <div className="app__error">
                        {subtitle}
                        {subtitle && <hr />}
                        <pre>{errorMessage}</pre>
                    </div>
                )
            } else {
                subtitle = <div className="app__error">{subtitle}</div>
            }
            return <HeroPage icon={ServerIcon} title={`${statusCode}: ${statusText}`} subtitle={subtitle} />
        }

        if (this.state.user === undefined) {
            return null
        }

        return [
            <BrowserRouter key={0}>
                <Route path="/" render={this.renderLayout} />
            </BrowserRouter>,
            <Tooltip key={1} />,
            USE_PLATFORM ? <Notifications key={2} extensionsController={this.state.extensionsController} /> : null,
        ]
    }

    private renderLayout = (props: RouteComponentProps<any>) => {
        let viewerSubject: LayoutProps['viewerSubject']
        if (this.state.user) {
            viewerSubject = this.state.user
        } else if (
            this.state.configurationCascade &&
            !isErrorLike(this.state.configurationCascade) &&
            this.state.configurationCascade.subjects &&
            !isErrorLike(this.state.configurationCascade.subjects) &&
            this.state.configurationCascade.subjects.length > 0
        ) {
            viewerSubject = this.state.configurationCascade.subjects[0].subject
        } else {
            viewerSubject = SITE_SUBJECT_NO_ADMIN
        }

        return (
            <Layout
                {...props}
                /* Checked for undefined in render() above */
                user={this.state.user as GQL.IUser | null}
                viewerSubject={viewerSubject}
                isLightTheme={this.state.isLightTheme}
                onThemeChange={this.onThemeChange}
                navbarSearchQuery={this.state.navbarSearchQuery}
                onNavbarQueryChange={this.onNavbarQueryChange}
                showHelpPopover={this.state.showHelpPopover}
                showHistoryPopover={this.state.showHistoryPopover}
                onHelpPopoverToggle={this.onHelpPopoverToggle}
                onHistoryPopoverToggle={this.onHistoryPopoverToggle}
                configurationCascade={this.state.configurationCascade}
                extensions={this.state.extensions}
                extensionsEnvironment={this.state.extensionsEnvironment}
                extensionsOnComponentChange={this.extensionsOnComponentChange}
                extensionsOnRootChange={this.extensionsOnRootChange}
                extensionsController={this.state.extensionsController}
            />
        )
    }

    private onThemeChange = () => {
        this.setState(
            state => ({ isLightTheme: !state.isLightTheme }),
            () => {
                eventLogger.log(this.state.isLightTheme ? 'LightThemeClicked' : 'DarkThemeClicked')
            }
        )
    }

    private onNavbarQueryChange = (navbarSearchQuery: string) => {
        this.setState({ navbarSearchQuery })
    }

    private onHelpPopoverToggle = (visible?: boolean): void => {
        eventLogger.log('HelpPopoverToggled')
        this.setState(prevState => ({
            // If visible is any non-boolean type (e.g., MouseEvent), treat it as undefined. This lets callers use
            // onHelpPopoverToggle directly in an event handler without wrapping it in an another function.
            showHelpPopover: visible !== true && visible !== false ? !prevState.showHelpPopover : visible,
        }))
    }

    private onHistoryPopoverToggle = (visible?: boolean): void => {
        eventLogger.log('HistoryPopoverToggled')
        this.setState(prevState => ({
            showHistoryPopover: visible !== true && visible !== false ? !prevState.showHistoryPopover : visible,
        }))
    }

    private onConfigurationCascadeChange(
        configurationCascade: ConfigurationCascadeOrError<ConfigurationSubject, Settings>
    ): void {
        this.setState(
            prevState => {
                const update: Pick<AppState, 'configurationCascade' | 'extensionsEnvironment'> = {
                    configurationCascade,
                    extensionsEnvironment: prevState.extensionsEnvironment,
                }
                if (
                    configurationCascade.subjects !== null &&
                    !isErrorLike(configurationCascade.subjects) &&
                    configurationCascade.merged !== null &&
                    !isErrorLike(configurationCascade.merged)
                ) {
                    // Only update Sourcegraph extensions environment configuration if the configuration was
                    // successfully parsed.
                    //
                    // TODO(sqs): Think through how this error should be handled.
                    update.extensionsEnvironment = {
                        ...prevState.extensionsEnvironment,
                        configuration: {
                            subjects: configurationCascade.subjects.filter(
                                (subject): subject is ConfiguredSubject<ConfigurationSubject, Settings> =>
                                    subject.settings !== null && !isErrorLike(subject.settings)
                            ),
                            merged: configurationCascade.merged,
                        },
                    }
                }
                return update
            },
            () => this.state.extensionsController.setEnvironment(this.state.extensionsEnvironment)
        )
    }

    private onViewerConfiguredExtensionsChange(viewerConfiguredExtensions: ConfiguredExtension[]): void {
        this.setState(
            prevState => ({
                extensionsEnvironment: {
                    ...prevState.extensionsEnvironment,
                    extensions: viewerConfiguredExtensions,
                },
            }),
            () => this.state.extensionsController.setEnvironment(this.state.extensionsEnvironment)
        )
    }

    private extensionsOnComponentChange = (component: ExtensionsComponent | null): void => {
        this.setState(
            prevState => ({ extensionsEnvironment: { ...prevState.extensionsEnvironment, component } }),
            () => this.state.extensionsController.setEnvironment(this.state.extensionsEnvironment)
        )
    }

    private extensionsOnRootChange = (root: URI | null): void => {
        this.setState(
            prevState => ({ extensionsEnvironment: { ...prevState.extensionsEnvironment, root } }),
            () => this.state.extensionsController.setEnvironment(this.state.extensionsEnvironment)
        )
    }
}
