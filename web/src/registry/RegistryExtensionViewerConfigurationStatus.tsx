import CheckmarkIcon from '@sourcegraph/icons/lib/Checkmark'
import GearIcon from '@sourcegraph/icons/lib/Gear'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, startWith, switchMap } from 'rxjs/operators'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { LinkOrSpan } from '../components/LinkOrSpan'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../util/errors'
import { registryExtensionConfigurationSubjectFragment } from './RegistryExtensionConfigurationSubjectsList'
import { RegistryExtensionConfigureButton } from './RegistryExtensionConfigureButton'

const registryExtensionViewerConfigurationSubjectFragment = gql`
    fragment RegistryExtensionConfigurationSubjectEdgeFields on ExtensionConfigurationSubjectEdge {
        node {
            ...RegistryExtensionConfigurationSubjectFields
        }
        extension {
            id
            extensionID
        }
        isEnabled
        url
    }
    ${registryExtensionConfigurationSubjectFragment}
`

const RegistryExtensionViewerConfigurationSubjectEdge: React.SFC<{
    edge: GQL.IExtensionConfigurationSubjectEdge
    className: string
}> = ({ edge, className }) => {
    let label: React.ReactFragment
    let noun: string
    switch (edge.node.__typename) {
        case 'Site':
            label = 'All users (global settings)'
            noun = 'global settings'
            break
        case 'Org':
            label = edge.node.name
            noun = `organization settings for ${edge.node.name}`
            break
        case 'User':
            label = edge.node.username
            noun = `user settings for ${edge.node.username}`
            break
        default:
            label = `Unknown`
            noun = 'settings'
            break
    }

    const color = edge.isEnabled ? 'success' : 'secondary'

    return (
        <LinkOrSpan
            to={edge.node.viewerCanAdminister ? edge.url : null}
            className={`list-group-item ${className} d-flex justify-content-between align-items-center list-group-item-${color} border-${color}`}
            title={`Extension is ${edge.isEnabled ? 'enabled' : 'disabled'} in ${noun}`}
        >
            <span>
                <strong>{label}</strong>{' '}
                {edge.isEnabled ? (
                    <CheckmarkIcon className="icon-inline" />
                ) : (
                    <span className="text-muted">(disabled)</span>
                )}
            </span>
            {edge.node.viewerCanAdminister && <GearIcon className="icon-inline" />}
        </LinkOrSpan>
    )
}

interface Props extends RouteComponentProps<{}> {
    authenticatedUser: Pick<GQL.IUser, 'id'> | null

    extension: Pick<GQL.IRegistryExtension, 'id' | 'viewerHasEnabled' | 'viewerCanConfigure'>

    nonEmptyClassName: string

    onUpdate: () => void
}

const LOADING: 'loading' = 'loading'

interface State {
    /** The configuration subjects, 'loading', or an error. */
    subjectsOrError: typeof LOADING | GQL.IExtensionConfigurationSubjectEdge[] | ErrorLike
}

/**
 * Displays the viewer's ancestor configuration subjects for whom an extension is enabled.
 *
 * Unlike the similarly named RegistryExtensionConfigurationSubjectsList, this is for showing the viewer "why is
 * this extension enabled for me?", not "which other organizations use this extension?".
 */
export class RegistryExtensionViewerConfigurationStatus extends React.PureComponent<Props, State> {
    public state: State = {
        subjectsOrError: LOADING,
    }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const extensionChanges = this.componentUpdates.pipe(
            map(({ extension }) => extension),
            // Updating when viewerHasEnabled changes makes it so that clicking "Enable/disable extension" in
            // the header immediately updates this list.
            distinctUntilChanged((a, b) => a.id === b.id && a.viewerHasEnabled === b.viewerHasEnabled)
        )

        this.subscriptions.add(
            extensionChanges
                .pipe(
                    switchMap(extension =>
                        queryRegistryExtensionViewerConfigurationSubjects({ extension: extension.id }).pipe(
                            catchError(error => [asError(error)]),
                            map(result => ({ subjectsOrError: result })),
                            startWith<Pick<State, 'subjectsOrError'>>({ subjectsOrError: LOADING })
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate as State), err => console.error(err))
        )

        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(nextProps: Props): void {
        this.componentUpdates.next(nextProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const itemClassName = 'py-2'

        if (
            (this.props.extension.viewerHasEnabled || this.props.extension.viewerCanConfigure) &&
            this.state.subjectsOrError === LOADING
        ) {
            // Reserve vertical space for a common case where we know we (usually) need 1 row.
            return (
                <div className={`list-group ${this.props.nonEmptyClassName}`}>
                    <div className={`list-group-item ${itemClassName}`}>&nbsp;</div>
                </div>
            )
        }

        return this.state.subjectsOrError === LOADING ? null : isErrorLike(this.state.subjectsOrError) ? (
            <div className={`alert alert-danger ${this.props.nonEmptyClassName}`}>
                {upperFirst(this.state.subjectsOrError.message)}
            </div>
        ) : (
            <div className={`list-group ${this.props.nonEmptyClassName}`}>
                {this.state.subjectsOrError.length > 0
                    ? this.state.subjectsOrError.map((edge, i) => (
                          <RegistryExtensionViewerConfigurationSubjectEdge
                              key={i}
                              edge={edge}
                              className={itemClassName}
                          />
                      ))
                    : this.props.authenticatedUser && (
                          <RegistryExtensionConfigureButton
                              extensionGQLID={this.props.extension.id}
                              extensionID={undefined}
                              subject={this.props.authenticatedUser.id}
                              viewerCanConfigure={this.props.extension.viewerCanConfigure}
                              isEnabled={this.props.extension.viewerHasEnabled}
                              onDidUpdate={this.props.onUpdate}
                              buttonClassName="w-100"
                          />
                      )}
            </div>
        )
    }
}

function queryRegistryExtensionViewerConfigurationSubjects(args: {
    extension: GQL.ID
}): Observable<GQL.IExtensionConfigurationSubjectEdge[]> {
    return queryGraphQL(
        gql`
            query RegistryExtensionConfigurationSubjects($extension: ID!) {
                node(id: $extension) {
                    ... on RegistryExtension {
                        extensionConfigurationSubjects(users: true, viewer: true) {
                            edges {
                                ...RegistryExtensionConfigurationSubjectEdgeFields
                            }
                        }
                    }
                }
            }
            ${registryExtensionViewerConfigurationSubjectFragment}
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.node || errors) {
                throw createAggregateError(errors)
            }
            const node = data.node as GQL.IRegistryExtension
            if (!node.extensionConfigurationSubjects) {
                throw createAggregateError(errors)
            }
            return node.extensionConfigurationSubjects.edges
        })
    )
}
