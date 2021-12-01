import classNames from 'classnames'
import React from 'react'

import { useQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { FileSpec, RevisionSpec } from '@sourcegraph/shared/src/util/url'

import { CatalogEntityIcon } from '../../../enterprise/catalog/components/CatalogEntityIcon'
import {
    RepositoryFields,
    TreeEntryCatalogEntityResult,
    TreeEntryCatalogEntityVariables,
} from '../../../graphql-operations'
import { RepoHeaderActionButtonLink } from '../../components/RepoHeaderActions'
import { RepoHeaderContext } from '../../RepoHeader'

import styles from './CatalogEntityAction.module.scss'
import { TREE_ENTRY_CATALOG_ENTITY } from './gql'

// TODO(sqs): LICENSE move to enterprise/

// TODO(sqs): should this show up when there is no repository rev?

interface Props extends Partial<RevisionSpec>, Partial<FileSpec> {
    repo: Pick<RepositoryFields, 'id' | 'name'>

    actionType?: 'nav' | 'dropdown'
}

/**
 * A repository header action that displays the catalog entity associated with the current file
 * path.
 */
export const CatalogEntityAction: React.FunctionComponent<Props & RepoHeaderContext> = props => {
    const { data, error, loading } = useQuery<TreeEntryCatalogEntityResult, TreeEntryCatalogEntityVariables>(
        TREE_ENTRY_CATALOG_ENTITY,
        {
            variables: { repository: props.repo.id, rev: props.revision || 'HEAD', path: props.filePath || '' },

            // Cache this data but always re-request it in the background when we revisit
            // this page to pick up newer changes.
            fetchPolicy: 'cache-and-network',

            // For subsequent requests while this page is open, make additional network
            // requests; this is necessary for `refetch` to actually use the network. (see
            // https://github.com/apollographql/apollo-client/issues/5515)
            nextFetchPolicy: 'network-only',
        }
    )

    if (error) {
        throw error
    }
    if (loading && !data) {
        return null
    }

    const catalogEntities =
        (data && data.node?.__typename === 'Repository' && data.node.commit?.blob?.catalogEntities) || null

    if (!catalogEntities || catalogEntities.length === 0) {
        return null
    }

    const catalogEntity = catalogEntities[0]

    const icon = <CatalogEntityIcon entity={catalogEntity} className="icon-inline mr-1" />

    if (props.actionType === 'dropdown') {
        return (
            <RepoHeaderActionButtonLink to={catalogEntity.url} className="btn" file={true}>
                {icon} {catalogEntity.name} (in catalog)
            </RepoHeaderActionButtonLink>
        )
    }

    return (
        <RepoHeaderActionButtonLink
            to={catalogEntity.url}
            className={classNames('btn btn-icon small border border-primary', styles.btn)}
        >
            {icon} {catalogEntity.name}
        </RepoHeaderActionButtonLink>
    )
}
