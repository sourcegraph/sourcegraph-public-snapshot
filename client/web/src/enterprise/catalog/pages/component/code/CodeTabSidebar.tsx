import classNames from 'classnames'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import FileAlertIcon from 'mdi-react/FileAlertIcon'
import SlackIcon from 'mdi-react/SlackIcon'
import React from 'react'
import { useRouteMatch } from 'react-router'
import { Link } from 'react-router-dom'

import { ComponentKind } from '@sourcegraph/shared/src/schema'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import {
    ComponentOwnerFields,
    ComponentTagFields,
    RepositoryForTreeFields,
    SourceLocationSetContributorsFields,
    TreeEntryForTreeFields,
    TreeOrComponentSourceLocationSetFields,
} from '../../../../../graphql-operations'
import { PersonList } from '../../../components/person-list/PersonList'
import { SourceLocationSetTitle } from '../../../contributions/tree/SourceLocationSetTitle'
import { TreeOrComponentViewOptionsProps } from '../../../contributions/tree/TreeOrComponent'
import { ComponentOwnerSidebarItem } from '../meta/ComponentOwnerSidebarItem'
import { ComponentTagsSidebarItem } from '../meta/ComponentTagsSidebarItem'
import { SourceSetContributorsSidebarItem } from '../meta/SourceSetContributorsSidebarItem'

interface Props extends Pick<TreeOrComponentViewOptionsProps, 'treeOrComponentViewMode'> {
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
    sourceLocationSet: TreeOrComponentSourceLocationSetFields & SourceLocationSetContributorsFields
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
    sourceLocationSet,
    useHash,
    treeOrComponentViewMode,
    className,
}) => {
    const match = useRouteMatch()

    const pathSeparator = useHash ? '#' : '/'

    const featuredComponent = treeOrComponentViewMode === 'auto' ? component : null
    const description = featuredComponent?.description || (tree.isRoot && repository.description) || null

    const SECTION_CLASS_NAME = 'mb-3'

    return (
        <aside>
            {component !== null && treeOrComponentViewMode === 'auto' && (
                <h2 className="h5 mb-3">
                    <SourceLocationSetTitle
                        component={component}
                        tree={tree}
                        treeOrComponentViewMode={treeOrComponentViewMode}
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
            {sourceLocationSet.codeOwners && sourceLocationSet.codeOwners.edges.length > 0 && (
                <section className={SECTION_CLASS_NAME}>
                    <PersonList
                        title="Code owners"
                        titleLink={`${match.url}${pathSeparator}code-owners`}
                        titleCount={sourceLocationSet.codeOwners.totalCount}
                        listTag="ol"
                        orientation="summary"
                        items={sourceLocationSet.codeOwners.edges.map(codeOwner => ({
                            person: codeOwner.node,
                            text:
                                codeOwner.fileProportion >= 0.01
                                    ? `${(codeOwner.fileProportion * 100).toFixed(0)}%`
                                    : '<1%',
                            textTooltip: `Owns ${codeOwner.fileCount} ${pluralize('line', codeOwner.fileCount)}`,
                        }))}
                        className={className}
                    />
                </section>
            )}
            {sourceLocationSet.contributors && sourceLocationSet.contributors.edges.length > 0 && (
                <section className={SECTION_CLASS_NAME}>
                    <SourceSetContributorsSidebarItem
                        contributors={sourceLocationSet.contributors}
                        titleLink={`${match.url}${pathSeparator}contributors`}
                    />
                </section>
            )}
        </aside>
    )
}
