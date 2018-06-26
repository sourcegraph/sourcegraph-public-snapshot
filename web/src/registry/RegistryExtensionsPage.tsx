import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable, Subject } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { FilteredConnection, FilteredConnectionFilter } from '../components/FilteredConnection'
import { RegistryExtensionNodeCard } from '../registry/RegistryExtensionNodeCard'
import { createAggregateError } from '../util/errors'
import { RegistryExtensionNodeRow } from './RegistryExtensionNodeRow'

export const registryExtensionFragment = gql`
    fragment RegistryExtensionFields on RegistryExtension {
        id
        publisher {
            __typename
            ... on User {
                id
                username
                displayName
                url
            }
            ... on Org {
                id
                name
                displayName
                url
            }
        }
        extensionID
        extensionIDWithoutRegistry
        name
        manifest {
            raw
            title
            description
        }
        createdAt
        updatedAt
        url
        remoteURL
        registryName
        isLocal
        users {
            totalCount
        }
        viewerHasEnabled
        viewerCanConfigure
        viewerCanAdminister
    }
`

export interface RegistryExtensionNodeDisplayProps {
    /** Which form of the extension ID to show (values correspond to GraphQL API fields of RegistryExtension). */
    showExtensionID?: 'extensionID' | 'extensionIDWithoutRegistry' | 'name'

    /** Whether to show the extension's source (local/remote registry). */
    showSource?: boolean

    /** Whether to show the action to enable an extension. */
    showUserActions?: boolean

    /** Whether to show the action to delete an extension from the registry. */
    showDeleteAction?: boolean

    /** Whether to show the action to edit an extension. */
    showEditAction?: boolean

    /** Whether to show the last-updated timestamp. */
    showTimestamp?: boolean
}

export interface RegistryExtensionNodeProps extends RegistryExtensionNodeDisplayProps {
    node: GQL.IRegistryExtension
    authenticatedUserID: GQL.ID | null
    onDidUpdate: () => void
}

class FilteredRegistryExtensionConnection extends FilteredConnection<
    GQL.IRegistryExtension,
    Pick<RegistryExtensionNodeProps, 'onDidUpdate' | 'authenticatedUserID'>
> {}

/** Ways to display the list of extensions. */
export enum ExtensionsListViewMode {
    Cards = 'cards',
    List = 'list',
}

interface RegistryExtensionsListProps extends RegistryExtensionNodeDisplayProps, RouteComponentProps<{}> {
    /** Only show extensions from this publisher (or all if null). */
    publisher: Pick<GQL.RegistryPublisher, '__typename' | 'id'> | null

    authenticatedUser: Pick<GQL.IUser, 'id'> | null

    /** How the list should be displayed. */
    mode: ExtensionsListViewMode

    /** User-selectable filters for the list. */
    filters?: FilteredConnectionFilter[]
}

/**
 * Displays registry extensions.
 */
export class RegistryExtensionsList extends React.PureComponent<RegistryExtensionsListProps> {
    public static FILTERS: FilteredConnectionFilter[] = [
        {
            label: 'All',
            id: 'all',
            tooltip: 'Show all extensions',
            args: { remote: true, local: true },
        },
        {
            label: 'Remote',
            id: 'remote',
            tooltip: 'Show only extensions from the remote registry',
            args: { remote: true, local: false },
        },
        {
            label: 'Local',
            id: 'local',
            tooltip: 'Show only extensions from the local registry',
            args: { remote: false, local: true },
        },
    ]

    private updates = new Subject<void>()

    public render(): JSX.Element | null {
        const nodeProps: Pick<
            RegistryExtensionNodeProps,
            'onDidUpdate' | 'authenticatedUserID' | keyof RegistryExtensionNodeDisplayProps
        > = {
            onDidUpdate: this.onDidUpdateRegistryExtension,
            authenticatedUserID: this.props.authenticatedUser && this.props.authenticatedUser.id,
            showExtensionID: this.props.showExtensionID,
            showSource: this.props.showSource,
            showUserActions: this.props.showUserActions,
            showDeleteAction: this.props.showDeleteAction,
            showEditAction: this.props.showEditAction,
            showTimestamp: this.props.showTimestamp,
        }

        return (
            <FilteredRegistryExtensionConnection
                className="registry-extensions-list"
                listClassName={
                    this.props.mode === ExtensionsListViewMode.Cards ? 'row mt-3' : 'list-group list-group-flush'
                }
                listComponent={this.props.mode === ExtensionsListViewMode.Cards ? 'div' : 'ul'}
                noun="registry extension"
                pluralNoun="registry extensions"
                queryConnection={this.queryRegistryExtensions}
                nodeComponent={
                    this.props.mode === ExtensionsListViewMode.Cards
                        ? RegistryExtensionNodeCard
                        : RegistryExtensionNodeRow
                }
                nodeComponentProps={nodeProps}
                updates={this.updates}
                filters={this.props.filters}
                hideSearch={false}
                noSummaryIfAllNodesVisible={true}
                history={this.props.history}
                location={this.props.location}
            />
        )
    }

    private queryRegistryExtensions = (args: {
        query?: string
        first?: number
        local?: boolean
        remote?: boolean
    }): Observable<GQL.IRegistryExtensionConnection> =>
        queryGraphQL(
            gql`
                query RegistryExtensions(
                    $first: Int
                    $publisher: ID
                    $query: String
                    $local: Boolean
                    $remote: Boolean
                ) {
                    extensionRegistry {
                        extensions(
                            first: $first
                            publisher: $publisher
                            query: $query
                            local: $local
                            remote: $remote
                        ) {
                            nodes {
                                ...RegistryExtensionFields
                            }
                            totalCount
                            pageInfo {
                                hasNextPage
                            }
                            error
                        }
                    }
                }
                ${registryExtensionFragment}
            `,
            {
                ...args,
                publisher: this.props.publisher && this.props.publisher.id,
                local: args.local === undefined ? true : args.local,
                remote: args.remote === undefined ? true : args.remote,
            }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.extensionRegistry || !data.extensionRegistry.extensions || errors) {
                    throw createAggregateError(errors)
                }
                return data.extensionRegistry.extensions
            })
        )

    private onDidUpdateRegistryExtension = () => this.updates.next()
}
