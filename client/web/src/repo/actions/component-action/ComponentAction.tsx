import classNames from 'classnames'
import React, { useRef, useState } from 'react'

import { useQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { FileSpec, RevisionSpec } from '@sourcegraph/shared/src/util/url'

import { ComponentIcon } from '../../../enterprise/catalog/components/ComponentIcon'
import { OverviewStatusContexts } from '../../../enterprise/catalog/pages/component/OverviewStatusContexts'
import { ComponentStateIndicator } from '../../../enterprise/catalog/pages/overview/components/entity-state-indicator/ComponentStateIndicator'
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

import styles from './ComponentAction.module.scss'
import { COMPONENTS_FOR_TREE_ENTRY_HEADER_ACTION } from './gql'

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
    const { data, error, loading } = useQuery<
        ComponentsForTreeEntryHeaderActionResult,
        ComponentsForTreeEntryHeaderActionVariables
    >(COMPONENTS_FOR_TREE_ENTRY_HEADER_ACTION, {
        variables: { repository: props.repo.id, rev: props.revision || 'HEAD', path: props.filePath || '' },
        fetchPolicy: 'cache-and-network',
    })

    if (error) {
        throw error
    }
    if (loading && !data) {
        return null
    }

    const components =
        (data && data.node?.__typename === 'Repository' && data.node.commit?.treeEntry?.components) || null

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
            buttonClassName={classNames('btn btn-icon small border border-primary', styles.btn)}
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
                    <ComponentStateIndicator component={component} className="ml-1" />
                </h3>
                {component.description && <p>{component.description}</p>}
                <OverviewStatusContexts component={component} itemClassName="mb-3" />
            </Popover>
        </>
    )
}
