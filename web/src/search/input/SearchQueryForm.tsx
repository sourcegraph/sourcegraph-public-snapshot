import * as H from 'history'
import * as React from 'react'
import { parseSearchURLQuery } from '..'
import * as GQL from '../../../../shared/src/graphql/schema'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { Form } from '../../components/Form'
import { queryIndexOfScope, submitSearch } from '../helpers'
import { QueryBuilder } from './QueryBuilder'
import { QueryInput } from './QueryInput'
import { SearchButton } from './SearchButton'
import { SearchFilterChips } from './SearchFilterChips'

interface State {
    /** The query value entered by the user in the query input */
    userQuery: string
    /** The query that results from combining all values in the query builder form. */
    builderQuery: string
}

interface Props extends SettingsCascadeProps {
    authenticatedUser: GQL.IUser | null
    location: H.Location
    history: H.History
    isLightTheme: boolean
    onThemeChange: () => void
}

export class SearchQueryForm extends React.Component<Props, State> {
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
        if (
            window.context.sourcegraphDotComMode &&
            !localStorage.getItem(SearchQueryForm.HIDE_REPOGROUP_SAMPLE_STORAGE_KEY) &&
            !this.state.userQuery
        ) {
            this.setState({ userQuery: 'repogroup:sample' })
        }
    }

    public render(): JSX.Element | null {
        return (
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
                <div className="search-page__input-sub-container">
                    <SearchFilterChips
                        location={this.props.location}
                        history={this.props.history}
                        query={this.state.userQuery}
                        authenticatedUser={this.props.authenticatedUser}
                        settingsCascade={this.props.settingsCascade}
                    />
                </div>
                <QueryBuilder
                    onFieldsQueryChange={this.onBuilderQueryChange}
                    isDotCom={window.context.sourcegraphDotComMode}
                />
            </Form>
        )
    }

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        const query = [this.state.builderQuery, this.state.userQuery].filter(s => !!s).join(' ')
        submitSearch(this.props.history, query, 'home')
    }

    private onUserQueryChange = (userQuery: string) => {
        this.setState({ userQuery })

        if (window.context.sourcegraphDotComMode) {
            if (queryIndexOfScope(userQuery, 'repogroup:sample') !== -1) {
                localStorage.removeItem(SearchQueryForm.HIDE_REPOGROUP_SAMPLE_STORAGE_KEY)
            } else {
                localStorage.setItem(SearchQueryForm.HIDE_REPOGROUP_SAMPLE_STORAGE_KEY, 'true')
            }
        }
    }

    private onBuilderQueryChange = (builderQuery: string) => {
        this.setState({ builderQuery })

        if (window.context.sourcegraphDotComMode) {
            if (queryIndexOfScope(builderQuery, 'repogroup:sample') !== -1) {
                localStorage.removeItem(SearchQueryForm.HIDE_REPOGROUP_SAMPLE_STORAGE_KEY)
            } else {
                localStorage.setItem(SearchQueryForm.HIDE_REPOGROUP_SAMPLE_STORAGE_KEY, 'true')
            }
        }
    }
}
