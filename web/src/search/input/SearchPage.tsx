import * as H from 'history'
import LinkIcon from 'mdi-react/LinkIcon'
import * as React from 'react'
import { parseSearchURLQuery } from '..'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import * as GQL from '../../../../shared/src/graphql/schema'
import { isSettingsValid, SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { Form } from '../../components/Form'
import { PageTitle } from '../../components/PageTitle'
import { Notices } from '../../global/Notices'
import { QuickLink, Settings } from '../../schema/settings.schema'
import { ThemePreferenceProps, ThemeProps } from '../../theme'
import { eventLogger } from '../../tracking/eventLogger'
import { limitString } from '../../util'
import { queryIndexOfScope, submitSearch } from '../helpers'
import { QueryBuilder } from './QueryBuilder'
import { QueryInput } from './QueryInput'
import { SearchButton } from './SearchButton'
import { ISearchScope, SearchFilterChips } from './SearchFilterChips'

interface Props extends SettingsCascadeProps, ThemeProps, ThemePreferenceProps, ActivationProps {
    authenticatedUser: GQL.IUser | null
    location: H.Location
    history: H.History
    isSourcegraphDotCom: boolean
}

interface State {
    /** The query value entered by the user in the query input */
    userQuery: string
    /** The query that results from combining all values in the query builder form. */
    builderQuery: string
}

/**
 * The search page
 */
export class SearchPage extends React.Component<Props, State> {
    private static HIDE_REPOGROUP_SAMPLE_STORAGE_KEY = 'SearchPage/hideRepogroupSample'

    constructor(props: Props) {
        super(props)

        const query = parseSearchURLQuery(props.location.search)
        this.state = {
            userQuery: query || '',
            builderQuery: '',
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('Home')
        if (
            window.context.sourcegraphDotComMode &&
            !localStorage.getItem(SearchPage.HIDE_REPOGROUP_SAMPLE_STORAGE_KEY) &&
            !this.state.userQuery
        ) {
            this.setState({ userQuery: 'repogroup:sample' })
        }
    }

    public render(): JSX.Element | null {
        let logoUrl =
            `${window.context.assetsRoot}/img/sourcegraph` +
            (this.props.isLightTheme ? '-light' : '') +
            '-head-logo.svg'
        const { branding } = window.context
        if (branding) {
            if (this.props.isLightTheme) {
                if (branding.light && branding.light.logo) {
                    logoUrl = branding.light.logo
                }
            } else if (branding.dark && branding.dark.logo) {
                logoUrl = branding.dark.logo
            }
        }
        const hasScopes = this.getScopes().length > 0
        return (
            <div className="search-page">
                <PageTitle title={this.getPageTitle()} />
                <img className="search-page__logo" src={logoUrl} />
                <Form className="search search-page__container" onSubmit={this.onSubmit}>
                    <div className="search-page__input-container">
                        <QueryInput
                            {...this.props}
                            value={this.state.userQuery}
                            onChange={this.onUserQueryChange}
                            autoFocus={'cursor-at-end'}
                            hasGlobalQueryBehavior={true}
                        />
                        <SearchButton />
                    </div>
                    {hasScopes ? (
                        <>
                            <div className="search-page__input-sub-container">
                                <SearchFilterChips
                                    location={this.props.location}
                                    history={this.props.history}
                                    query={this.state.userQuery}
                                    authenticatedUser={this.props.authenticatedUser}
                                    settingsCascade={this.props.settingsCascade}
                                    isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                                />
                            </div>
                            {this.renderQuickLinks()}
                            <QueryBuilder
                                onFieldsQueryChange={this.onBuilderQueryChange}
                                isSourcegraphDotCom={window.context.sourcegraphDotComMode}
                            />
                        </>
                    ) : (
                        <>
                            <QueryBuilder
                                onFieldsQueryChange={this.onBuilderQueryChange}
                                isSourcegraphDotCom={window.context.sourcegraphDotComMode}
                            />
                            {this.renderQuickLinks()}
                            <div className="search-page__input-sub-container">
                                <SearchFilterChips
                                    location={this.props.location}
                                    history={this.props.history}
                                    query={this.state.userQuery}
                                    authenticatedUser={this.props.authenticatedUser}
                                    settingsCascade={this.props.settingsCascade}
                                    isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                                />
                            </div>
                        </>
                    )}
                    <Notices className="my-3" location="home" settingsCascade={this.props.settingsCascade} />
                </Form>
            </div>
        )
    }

    private renderQuickLinks(): JSX.Element | null {
        const quickLinks = this.getQuickLinks()
        return quickLinks.length > 0 ? (
            <div className="search-page__input-sub-container">
                {quickLinks.map((quickLink, i) => (
                    <small className="text-nowrap mr-2 search-page__quicklink">
                        <a href={quickLink.url} data-tooltip={quickLink.description} key={i}>
                            <LinkIcon className="icon-inline pr-1" />
                            {quickLink.name}
                        </a>
                    </small>
                ))}
            </div>
        ) : null
    }

    private onUserQueryChange = (userQuery: string) => {
        this.setState({ userQuery })

        if (window.context.sourcegraphDotComMode) {
            if (queryIndexOfScope(userQuery, 'repogroup:sample') !== -1) {
                localStorage.removeItem(SearchPage.HIDE_REPOGROUP_SAMPLE_STORAGE_KEY)
            } else {
                localStorage.setItem(SearchPage.HIDE_REPOGROUP_SAMPLE_STORAGE_KEY, 'true')
            }
        }
    }

    private onBuilderQueryChange = (builderQuery: string) => {
        this.setState({ builderQuery })
    }

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        const query = [this.state.builderQuery, this.state.userQuery].filter(s => !!s).join(' ')
        submitSearch(this.props.history, query, 'home', this.props.activation)
    }

    private getPageTitle(): string | undefined {
        const query = parseSearchURLQuery(this.props.location.search)
        if (query) {
            return `${limitString(this.state.userQuery, 25, true)}`
        }
        return undefined
    }

    private getScopes(): ISearchScope[] {
        return (
            (isSettingsValid<Settings>(this.props.settingsCascade) &&
                this.props.settingsCascade.final['search.scopes']) ||
            []
        )
    }

    private getQuickLinks(): QuickLink[] {
        return (
            (isSettingsValid<Settings>(this.props.settingsCascade) && this.props.settingsCascade.final.quicklinks) || []
        )
    }
}
