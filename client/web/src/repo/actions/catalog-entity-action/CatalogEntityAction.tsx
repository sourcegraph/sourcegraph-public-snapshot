import classNames from 'classnames'
import React, { useRef, useState } from 'react'

import { useQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { FileSpec, RevisionSpec } from '@sourcegraph/shared/src/util/url'

import { CatalogEntityIcon } from '../../../enterprise/catalog/components/CatalogEntityIcon'
import { Popover } from '../../../enterprise/insights/components/popover/Popover'
import {
    RepositoryFields,
    TreeEntryCatalogEntityResult,
    TreeEntryCatalogEntityVariables,
    TreeEntryCatalogEntityFields,
} from '../../../graphql-operations'
import { RepoHeaderActionButtonLink } from '../../components/RepoHeaderActions'
import { RepoHeaderContext } from '../../RepoHeader'

import styles from './CatalogEntityAction.module.scss'
import { TREE_ENTRY_CATALOG_ENTITY } from './gql'
import { CatalogEntityStateIndicator } from '../../../enterprise/catalog/pages/overview/components/entity-list/EntityList'

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

    const entity = catalogEntities[0]

    if (props.actionType === 'dropdown') {
        return (
            <RepoHeaderActionButtonLink to={entity.url} className="btn" file={true}>
                <CatalogEntityIcon entity={entity} className="icon-inline mr-1" /> {entity.name} (in catalog)
            </RepoHeaderActionButtonLink>
        )
    }

    return (
        <CatalogEntityActionPopoverButton
            entity={entity}
            buttonClassName={classNames('btn btn-icon small border border-primary', styles.btn)}
        />
    )
}

const CatalogEntityActionPopoverButton: React.FunctionComponent<{
    entity: TreeEntryCatalogEntityFields['catalogEntities'][0]
    buttonClassName?: string
}> = ({ entity, buttonClassName }) => {
    const targetButtonReference = useRef<HTMLButtonElement>(null)
    const [isOpen, setIsOpen] = useState(false)

    return (
        <>
            <div ref={targetButtonReference}>
                <RepoHeaderActionButtonLink to={entity.url} className={buttonClassName}>
                    <CatalogEntityIcon entity={entity} className="icon-inline mr-1" /> {entity.name}
                </RepoHeaderActionButtonLink>
            </div>
            <Popover
                isOpen={isOpen}
                target={targetButtonReference}
                interaction="hover"
                onVisibilityChange={setIsOpen}
                className="p-2"
            >
                <h4>
                    <CatalogEntityIcon entity={entity} className="icon-inline mr-1" /> {entity.name}
                    <CatalogEntityStateIndicator entity={entity} className="ml-1" />
                </h4>
                {entity.description && <p>{entity.description}</p>}
            </Popover>
        </>
    )
}
