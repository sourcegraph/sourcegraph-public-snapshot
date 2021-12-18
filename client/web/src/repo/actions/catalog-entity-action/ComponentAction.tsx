import classNames from 'classnames'
import React, { useRef, useState } from 'react'

import { useQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { FileSpec, RevisionSpec } from '@sourcegraph/shared/src/util/url'

import { ComponentIcon } from '../../../enterprise/catalog/components/ComponentIcon'
import { OverviewStatusContexts } from '../../../enterprise/catalog/pages/entity-detail/global/OverviewStatusContexts'
import { ComponentStateIndicator } from '../../../enterprise/catalog/pages/overview/components/entity-state-indicator/EntityStateIndicator'
import { positionBottomRight } from '../../../enterprise/insights/components/context-menu/utils'
import { Popover } from '../../../enterprise/insights/components/popover/Popover'
import {
    RepositoryFields,
    TreeEntryComponentResult,
    TreeEntryComponentVariables,
    TreeEntryComponentFields,
} from '../../../graphql-operations'
import { RepoHeaderActionButtonLink } from '../../components/RepoHeaderActions'
import { RepoHeaderContext } from '../../RepoHeader'

import styles from './ComponentAction.module.scss'
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
export const ComponentAction: React.FunctionComponent<Props & RepoHeaderContext> = props => {
    const { data, error, loading } = useQuery<TreeEntryComponentResult, TreeEntryComponentVariables>(
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

    const components = (data && data.node?.__typename === 'Repository' && data.node.commit?.blob?.components) || null

    if (!components || components.length === 0) {
        return null
    }

    const entity = components[0]

    if (props.actionType === 'dropdown') {
        return (
            <RepoHeaderActionButtonLink to={entity.url} className="btn" file={true}>
                <ComponentIcon component={entity} className="icon-inline mr-1" /> {entity.name} (in catalog)
            </RepoHeaderActionButtonLink>
        )
    }

    return (
        <ComponentActionPopoverButton
            entity={entity}
            buttonClassName={classNames('btn btn-icon small border border-primary', styles.btn)}
        />
    )
}

const ComponentActionPopoverButton: React.FunctionComponent<{
    entity: TreeEntryComponentFields['components'][0]
    buttonClassName?: string
}> = ({ entity, buttonClassName }) => {
    const targetButtonReference = useRef<HTMLButtonElement>(null)
    const [isOpen, setIsOpen] = useState(false)

    return (
        <>
            <div ref={targetButtonReference}>
                <RepoHeaderActionButtonLink to={entity.url} className={buttonClassName}>
                    <ComponentIcon component={entity} className="icon-inline mr-1" /> {entity.name}
                </RepoHeaderActionButtonLink>
            </div>
            <Popover
                isOpen={isOpen}
                target={targetButtonReference}
                interaction="hover"
                onVisibilityChange={setIsOpen}
                position={positionBottomRight}
                className="p-3"
                style={{ maxWidth: '50vw' }}
            >
                <h3>
                    <ComponentIcon component={entity} className="icon-inline mr-1" /> {entity.name}
                    <ComponentStateIndicator entity={entity} className="ml-1" />
                </h3>
                {entity.description && <p>{entity.description}</p>}
                <OverviewStatusContexts entity={entity} itemClassName="mb-3" />
            </Popover>
        </>
    )
}
