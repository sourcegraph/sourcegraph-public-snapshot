import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable, Subject } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { FilteredConnectionDisplayProps, FilteredConnectionFilter } from '../components/FilteredConnection'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { createAggregateError } from '../util/errors'
import {
    configuredExtensionFragment,
    ConfiguredExtensionNode,
    ConfiguredExtensionNodeDisplayProps,
    ConfiguredExtensionNodeProps,
    FilteredConfiguredExtensionConnection,
} from './ConfiguredExtensionNode'

export interface ConfiguredExtensionsListProps
    extends ConfiguredExtensionNodeDisplayProps,
        RouteComponentProps<{}>,
        Pick<FilteredConnectionDisplayProps, 'emptyElement' | 'onFilterSelect'> {
    /** Show extensions configured for this subject. */
    subject: Pick<GQL.ExtensionConfigurationSubject, '__typename' | 'id' | 'settingsURL' | 'viewerCanAdminister'>

    /** Update the connection from the data source when this value changes. */
    updateOnChange?: any
}

const FILTER_ALL_ID = 'all'

/**
 * Displays a list of all extensions used by a configuration subject.
 */
class ConfiguredExtensionsList extends React.PureComponent<ConfiguredExtensionsListProps> {
    private static FILTERS: FilteredConnectionFilter[] = [
        {
            label: 'All',
            id: FILTER_ALL_ID,
            tooltip: 'Show all extensions referenced in configuration (including disabled and invalid extensions)',
            args: { enabled: true, disabled: true, invalid: true },
        },
        {
            label: 'Enabled',
            id: 'enabled',
            tooltip: 'Show only extensions that are enabled',
            args: { enabled: true, disabled: false, invalid: true },
        },
        {
            label: 'Disabled',
            id: 'disabled',
            tooltip: 'Show only extensions that are disabled',
            args: { enabled: false, disabled: true, invalid: true },
        },
    ]

    private updates = new Subject<void>()

    public render(): JSX.Element | null {
        const nodeProps: Pick<
            ConfiguredExtensionNodeProps,
            'onDidUpdate' | 'settingsURL' | keyof ConfiguredExtensionNodeDisplayProps
        > = {
            settingsURL: this.props.subject.settingsURL,
            onDidUpdate: this.onDidUpdateConfiguredExtension,
            showUserActions: this.props.showUserActions,
        }

        return (
            <FilteredConfiguredExtensionConnection
                listClassName="list-group list-group-flush"
                noun="extension"
                pluralNoun="extensions"
                queryConnection={this.queryConfiguredExtensions}
                nodeComponent={ConfiguredExtensionNode}
                nodeComponentProps={nodeProps}
                updates={this.updates}
                hideSearch={true}
                filters={this.props.subject.viewerCanAdminister ? ConfiguredExtensionsList.FILTERS : []}
                noSummaryIfAllNodesVisible={true}
                updateOnChange={this.props.updateOnChange}
                emptyElement={this.props.emptyElement}
                onFilterSelect={this.props.onFilterSelect}
                compact={true}
                history={this.props.history}
                location={this.props.location}
            />
        )
    }

    private queryConfiguredExtensions = (
        args: GQL.IConfiguredExtensionsOnExtensionConfigurationSubjectArguments
    ): Observable<GQL.IConfiguredExtensionConnection> =>
        queryGraphQL(
            gql`
                query ConfiguredExtensions(
                    $subject: ID!
                    $first: Int
                    $enabled: Boolean
                    $disabled: Boolean
                    $invalid: Boolean
                ) {
                    extensionConfigurationSubject(id: $subject) {
                        configuredExtensions(first: $first, enabled: $enabled, disabled: $disabled, invalid: $invalid) {
                            nodes {
                                ...ConfiguredExtensionFields
                            }
                            totalCount
                            pageInfo {
                                hasNextPage
                            }
                        }
                    }
                }
                ${configuredExtensionFragment}
            `,
            {
                ...args,
                subject: this.props.subject && this.props.subject.id,
                enabled: args.enabled === undefined ? true : args.enabled,
                disabled: args.disabled === undefined ? false : args.disabled,
                invalid: args.invalid === undefined ? false : args.invalid,
            }
        ).pipe(
            map(({ data, errors }) => {
                if (
                    !data ||
                    !data.extensionConfigurationSubject ||
                    !data.extensionConfigurationSubject.configuredExtensions
                ) {
                    throw createAggregateError(errors)
                }
                return data.extensionConfigurationSubject.configuredExtensions
            })
        )

    private onDidUpdateConfiguredExtension = () => this.updates.next()
}

interface ConfiguredExtensionsPageProps extends ConfiguredExtensionsListProps {}

interface ConfiguredExtensionsPageState {
    activeFilter?: string
}

/**
 * Displays a page listing all extensions used by a configuration subject.
 */
export class ConfiguredExtensionsPage extends React.PureComponent<
    ConfiguredExtensionsPageProps,
    ConfiguredExtensionsPageState
> {
    public state: ConfiguredExtensionsPageState = {}

    public componentDidMount(): void {
        eventLogger.logViewEvent('ConfiguredExtensions')
    }

    public render(): JSX.Element | null {
        return (
            <div className="configured-extensions-page">
                <PageTitle title="Extensions used" />
                <ConfiguredExtensionsList
                    {...this.props}
                    onFilterSelect={this.onFilterSelect}
                    emptyElement={
                        this.props.subject.viewerCanAdminister && this.state.activeFilter === 'all' ? (
                            <div className="px-3 py-5 text-center bg-striped-secondary border">
                                <h4 className="text-muted mb-3">
                                    Enable extensions to add new features to Sourcegraph.
                                </h4>
                                <Link to="/registry" className="btn btn-primary">
                                    View available extensions in registry
                                </Link>
                            </div>
                        ) : (
                            undefined
                        )
                    }
                />
            </div>
        )
    }

    private onFilterSelect = (filterID: string | undefined): void => {
        this.setState({ activeFilter: filterID })
    }
}
