import PuzzleIcon from '@sourcegraph/icons/lib/Puzzle'
import * as React from 'react'
import { gql } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { LinkOrSpan } from '../components/LinkOrSpan'
import { pluralize } from '../util/strings'
import { extensionIDPrefix } from './extension'

export const registryPublisherFragment = gql`
    fragment RegistryPublisherFields on RegistryPublisher {
        __typename
        ... on User {
            id
            username
            registryExtensions {
                totalCount
                url
            }
        }
        ... on Org {
            id
            name
            registryExtensions {
                totalCount
                url
            }
        }
    }
`

export const RegistryPublisherNode: React.SFC<{
    node: GQL.RegistryPublisher
}> = ({ node }) => (
    <LinkOrSpan
        to={node.registryExtensions.url}
        className="list-group-item d-flex justify-content-between align-items-center"
    >
        <span>{extensionIDPrefix(node)}</span>
        <small
            title={`${node.registryExtensions.totalCount} ${pluralize(
                'published extension',
                node.registryExtensions.totalCount
            )}`}
        >
            <strong>{node.registryExtensions.totalCount}</strong>
            <PuzzleIcon className="icon-inline" />
        </small>
    </LinkOrSpan>
)
