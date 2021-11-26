import classNames from 'classnames'
import React, { useRef, useState } from 'react'

import { useQuery, gql } from '@sourcegraph/http-client'
import { FileSpec, RevisionSpec } from '@sourcegraph/shared/src/util/url'

import { ComponentIcon } from '../../../enterprise/catalog/components/ComponentIcon'
import { positionBottomRight } from '../../../enterprise/insights/components/context-menu/utils'
import { Popover } from '../../../enterprise/insights/components/popover/Popover'
import {
    ComponentsForTreeEntryHeaderActionFields,
    ComponentsForTreeEntryHeaderActionResult,
    ComponentsForTreeEntryHeaderActionVariables,
    RepositoryFields,
} from '../../../graphql-operations'
import { RepoHeaderActionButtonLink } from '../../components/RepoHeaderActions'
import { RepoHeaderContext } from '../../RepoHeader'

import styles from './ComponentRepoHeaderAction.module.scss'

// TODO(sqs): LICENSE move to enterprise/

// TODO(sqs): should this show up when there is no repository rev?

interface Props extends Partial<RevisionSpec>, Partial<FileSpec> {
    repo: Pick<RepositoryFields, 'id' | 'name'>

    actionType?: 'nav' | 'dropdown'
}

const COMPONENTS_FOR_TREE_ENTRY_HEADER_ACTION = gql`
    query ComponentsForTreeEntryHeaderAction($repository: ID!, $rev: String!, $path: String!) {
        node(id: $repository) {
            __typename
            ... on Repository {
                id
                commit(rev: $rev) {
                    id
                    treeEntry(path: $path) {
                        id
                        ...ComponentsForTreeEntryHeaderActionFields
                    }
                }
                components(path: $path, primary: true, recursive: false) {
                    id
                }
            }
        }
    }

    fragment ComponentsForTreeEntryHeaderActionFields on TreeEntry {
        components {
            __typename
            id
            name
            kind
            description
            url
        }
    }
`

/**
 * A repository header action that displays the catalog entity associated with the current file
 * path.
 */
export const ComponentRepoHeaderAction: React.FunctionComponent<Props & RepoHeaderContext> = props => {
    const { data, error, loading } = useQuery<
        ComponentsForTreeEntryHeaderActionResult,
        ComponentsForTreeEntryHeaderActionVariables
    >(COMPONENTS_FOR_TREE_ENTRY_HEADER_ACTION, {
        variables: { repository: props.repo.id, rev: props.revision || 'HEAD', path: props.filePath || '' },
        fetchPolicy: 'cache-first',
    })

    if (error) {
        throw error
    }
    if (loading && !data) {
        return null
    }

    const repository = data && data.node?.__typename === 'Repository' ? data.node : null
    if (!repository) {
        return null
    }

    // Don't show if there is a primary component for this path, since it will be shown more
    // prominently in the main content area.
    if (repository.components.length > 0) {
        return null
    }

    const components = repository.commit?.treeEntry?.components

    if (!components || components.length === 0) {
        return null
    }

    const component = components[0]

    if (props.actionType === 'dropdown') {
        return (
            <RepoHeaderActionButtonLink to={component.url} className="btn" file={true}>
                <ComponentIcon component={component} className="icon-inline mr-1" /> {component.name}
            </RepoHeaderActionButtonLink>
        )
    }

    return (
        <ComponentActionPopoverButton
            component={component}
            buttonClassName={classNames('btn btn-icon small border border-secondary', styles.btn)}
        />
    )
}

const ComponentActionPopoverButton: React.FunctionComponent<{
    component: ComponentsForTreeEntryHeaderActionFields['components'][0]
    buttonClassName?: string
}> = ({ component, buttonClassName }) => {
    const targetButtonReference = useRef<HTMLDivElement>(null)
    const [isOpen, setIsOpen] = useState(false)

    return (
        <>
            <div ref={targetButtonReference}>
                <RepoHeaderActionButtonLink to={component.url} className={buttonClassName}>
                    <ComponentIcon component={component} className="icon-inline mr-1" /> {component.name}
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
                    <ComponentIcon component={component} className="icon-inline mr-1" /> {component.name}
                </h3>
                {component.description && <p className="mb-0">{component.description}</p>}
            </Popover>
        </>
    )
}
