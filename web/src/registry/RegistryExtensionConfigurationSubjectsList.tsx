import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { FilteredConnection } from '../components/FilteredConnection'
import { createAggregateError } from '../util/errors'

export const registryExtensionConfigurationSubjectFragment = gql`
    fragment RegistryExtensionConfigurationSubjectFields on ExtensionConfigurationSubject {
        __typename
        settingsURL
        viewerCanAdminister
        ... on User {
            id
            username
            displayName
        }
        ... on Org {
            id
            name
            displayName
        }
        ... on Site {
            id
        }
    }
`

const RegistryExtensionConfigurationSubjectNode: React.SFC<{
    node: GQL.ExtensionConfigurationSubject
}> = ({ node }) => {
    let label: React.ReactFragment
    switch (node.__typename) {
        case 'Site':
            label = (
                <>
                    <strong>Global settings</strong> (all users)
                </>
            )
            break
        case 'Org':
            label = node.name
            break
        case 'User':
            label = node.username
            break
        default:
            label = `Unknown`
            break
    }

    return (
        <Link to={node.settingsURL} className="list-group-item px-0 py-2">
            {label}
        </Link>
    )
}

interface Props extends RouteComponentProps<{}> {
    extension: Pick<GQL.IRegistryExtension, 'id' | 'viewerHasEnabled'>
}

class FilteredRegistryExtensionConfigurationSubjectConnection extends FilteredConnection<
    GQL.ExtensionConfigurationSubject
> {}

/**
 * Displays the configuration subjects for whom an extension is enabled.
 */
export class RegistryExtensionConfigurationSubjectsList extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        return (
            <FilteredRegistryExtensionConfigurationSubjectConnection
                listClassName="list-group list-group-flush"
                listComponent="div"
                noun="organization using this extension"
                pluralNoun="organizations using this extension"
                queryConnection={this.queryRegistryExtensionConfigurationSubjects}
                // Updating when viewerHasEnabled changes makes it so that clicking "Enable/disable extension" in
                // the header immediately updates this list.
                updateOnChange={`${this.props.extension.id}:${this.props.extension.viewerHasEnabled}`}
                nodeComponent={RegistryExtensionConfigurationSubjectNode}
                hideSearch={true}
                history={this.props.history}
                location={this.props.location}
            />
        )
    }

    private queryRegistryExtensionConfigurationSubjects = (args: {
        first?: number
    }): Observable<GQL.IExtensionConfigurationSubjectConnection> =>
        queryGraphQL(
            gql`
                query RegistryExtensionConfigurationSubjects($extension: ID!, $first: Int) {
                    node(id: $extension) {
                        ... on RegistryExtension {
                            extensionConfigurationSubjects(first: $first) {
                                nodes {
                                    ...RegistryExtensionConfigurationSubjectFields
                                }
                                totalCount
                            }
                        }
                    }
                }
                ${registryExtensionConfigurationSubjectFragment}
            `,
            { ...args, extension: this.props.extension.id }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.node || errors) {
                    throw createAggregateError(errors)
                }
                const node = data.node as GQL.IRegistryExtension
                if (!node.extensionConfigurationSubjects) {
                    throw createAggregateError(errors)
                }
                return node.extensionConfigurationSubjects
            })
        )
}
