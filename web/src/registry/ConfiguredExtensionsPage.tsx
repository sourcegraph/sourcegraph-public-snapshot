import GearIcon from '@sourcegraph/icons/lib/Gear'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable, Subject } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { DismissibleAlert } from '../components/DismissibleAlert'
import { FilteredConnectionFilter } from '../components/FilteredConnection'
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
import { extensionIDPrefix, RegistryPublisher } from './extension'

interface ConfiguredExtensionsListProps extends ConfiguredExtensionNodeDisplayProps, RouteComponentProps<{}> {
    /** Show extensions configured for this subject. */
    subject: Pick<GQL.ExtensionConfigurationSubject, '__typename' | 'id' | 'settingsURL' | 'viewerCanAdminister'>

    /** Update the connection from the data source when this value changes. */
    updateOnChange?: any
}

/**
 * Displays a list of all extensions used by a configuration subject.
 */
class ConfiguredExtensionsList extends React.PureComponent<ConfiguredExtensionsListProps> {
    private static FILTERS: FilteredConnectionFilter[] = [
        {
            label: 'All',
            id: 'all',
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
                filters={ConfiguredExtensionsList.FILTERS}
                noSummaryIfAllNodesVisible={true}
                updateOnChange={this.props.updateOnChange}
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

interface ConfiguredExtensionsPageProps extends ConfiguredExtensionsListProps {
    /**
     * Also link to extensions published by this publisher (when the configuration subject also corresponds to a
     * publisher) if publisher.registryExtensions.url is set.
     */
    publisher?: {
        registryExtensions?: Pick<GQL.IRegistryExtensionConnection, 'url'>
    } & RegistryPublisher
}

/**
 * Displays a page listing all extensions used by a configuration subject.
 */
export class ConfiguredExtensionsPage extends React.PureComponent<ConfiguredExtensionsPageProps> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('ConfiguredExtensions')
    }

    public render(): JSX.Element | null {
        return (
            <div className="configured-extensions-page">
                <PageTitle title="Configured extensions" />
                <DismissibleAlert className="alert-info mb-3" partialStorageKey="configured-extensions-help0">
                    <span>
                        <strong>Experimental feature:</strong> See which extensions a user or organization uses.
                        Extensions add features to Sourcegraph and other connected tools (such as editors, code hosts,
                        and code review tools).
                    </span>
                </DismissibleAlert>
                <div className="d-flex justify-content-between align-items-center mb-3">
                    <h2 className="mr-sm-2 mb-0">Configured extensions</h2>
                    <div>
                        {this.props.publisher &&
                            this.props.publisher.registryExtensions &&
                            this.props.publisher.registryExtensions.url && (
                                <Link className="btn btn-outline-link" to={this.props.publisher.registryExtensions.url}>
                                    Extensions published by {extensionIDPrefix(this.props.publisher)}
                                </Link>
                            )}{' '}
                        {this.props.subject &&
                            this.props.subject.settingsURL &&
                            this.props.subject.viewerCanAdminister && (
                                <Link className="btn btn-primary" to={this.props.subject.settingsURL}>
                                    <GearIcon className="icon-inline" /> Configure extensions
                                </Link>
                            )}
                    </div>
                </div>
                <ConfiguredExtensionsList {...this.props} />
            </div>
        )
    }
}
