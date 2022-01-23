import classNames from 'classnames'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import FileAlertIcon from 'mdi-react/FileAlertIcon'
import SlackIcon from 'mdi-react/SlackIcon'
import React from 'react'
import { useRouteMatch } from 'react-router'
import { Link } from 'react-router-dom'

import { ComponentKind } from '@sourcegraph/shared/src/schema'

import {
    ComponentOwnerFields,
    ComponentTagFields,
    RepositoryForTreeFields,
    SourceSetContributorsFields,
    TreeEntryForTreeFields,
    SourceSetAtTreeFields,
} from '../../../../../../graphql-operations'
import { SourceSetTitle } from '../../../../contributions/tree/SourceSetTitle'
import { SourceSetAtTreeViewOptionsProps } from '../../../contributions/tree/useSourceSetAtTreeViewOptions'
import { ComponentOwnerSidebarItem } from './sidebar/ComponentOwnerSidebarItem'
import { ComponentTagsSidebarItem } from './sidebar/ComponentTagsSidebarItem'
import { SourceSetCodeOwnersSidebarItem } from './sidebar/SourceSetCodeOwnersSidebarItem'
import { SourceSetContributorsSidebarItem } from './sidebar/SourceSetContributorsSidebarItem'

interface Props extends Pick<SourceSetAtTreeViewOptionsProps, 'sourceSetAtTreeViewMode'> {
    repository: RepositoryForTreeFields
    tree: TreeEntryForTreeFields
    component:
        | null
        | (ComponentOwnerFields & {
              __typename: 'Component'
              name: string
              kind: ComponentKind
              description: string | null
              tags: ComponentTagFields[]
          })
    sourceSet: SourceSetAtTreeFields & SourceSetContributorsFields
    useHash?: boolean
    className?: string
}

/**
 * The code tab sidebar is shown on the default tab when browsing the tree. If the tree is also a
 * component, it shows some component information (but only a subset of what's available in the
 * "Component" tab).
 */
export const CodeTabSidebar: React.FunctionComponent<Props> = ({
    tree,
    repository,
    component,
    sourceSet,
    useHash,
    sourceSetAtTreeViewMode,
    className,
}) => {
    const match = useRouteMatch()

    const pathSeparator = useHash ? '#' : '/'

    const featuredComponent = sourceSetAtTreeViewMode === 'auto' ? component : null
    const description = featuredComponent?.description || (tree.isRoot && repository.description) || null

    const SECTION_CLASS_NAME = 'mb-3'

    return (
        <aside>
            {component !== null && sourceSetAtTreeViewMode === 'auto' && (
                <h2 className="h5 mb-3">
                    <SourceSetTitle
                        component={component}
                        tree={tree}
                        sourceSetAtTreeViewMode={sourceSetAtTreeViewMode}
                    />
                </h2>
            )}
            {description && (
                <section className={SECTION_CLASS_NAME}>
                    <p className="mb-0">{description}</p>
                </section>
            )}
            {featuredComponent && (
                <section
                    className={classNames(SECTION_CLASS_NAME, 'd-flex flex-wrap align-items-center small')}
                    style={{ gap: 'calc(0.5*var(--spacer))' }}
                >
                    <Link to="#" className="d-flex align-items-center text-body text-nowrap">
                        <FileAlertIcon className="icon-inline mr-1" />
                        Runbook
                    </Link>
                    <Link to="#" className="d-flex align-items-center text-body text-nowrap">
                        <AlertCircleOutlineIcon className="icon-inline mr-1" />
                        Issues
                    </Link>
                    <Link to="#" className="d-flex align-items-center text-body text-nowrap">
                        <SlackIcon className="icon-inline mr-1" />
                        #dev-frontend
                    </Link>
                </section>
            )}
            {featuredComponent?.tags && featuredComponent.tags.length > 0 && (
                <section className={SECTION_CLASS_NAME}>
                    <ComponentTagsSidebarItem tags={featuredComponent.tags} />
                </section>
            )}
            {featuredComponent && (
                <section className={SECTION_CLASS_NAME}>
                    <h4 className="font-weight-bold">Owner</h4>
                    <ComponentOwnerSidebarItem owner={featuredComponent.owner} />
                </section>
            )}
            {sourceSet.codeOwners && sourceSet.codeOwners.edges.length > 0 && (
                <section className={SECTION_CLASS_NAME}>
                    <SourceSetCodeOwnersSidebarItem
                        codeOwners={sourceSet.codeOwners}
                        titleLink={`${match.url}${pathSeparator}code-owners`}
                    />
                </section>
            )}
            {sourceSet.contributors && sourceSet.contributors.edges.length > 0 && (
                <section className={SECTION_CLASS_NAME}>
                    <SourceSetContributorsSidebarItem
                        contributors={sourceSet.contributors}
                        titleLink={`${match.url}${pathSeparator}contributors`}
                    />
                </section>
            )}
        </aside>
    )
}
