import classNames from 'classnames'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import FileAlertIcon from 'mdi-react/FileAlertIcon'
import SlackIcon from 'mdi-react/SlackIcon'
import React from 'react'
import { useRouteMatch } from 'react-router'
import { Link } from 'react-router-dom'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { pluralize } from '@sourcegraph/shared/src/util/strings'
import { Badge } from '@sourcegraph/wildcard'

import {
    ComponentDetailFields,
    RepositoryForTreeFields,
    TreeEntryForTreeFields,
    TreeOrComponentSourceLocationSetFields,
} from '../../../../../graphql-operations'
import { UserAvatar } from '../../../../../user/UserAvatar'
import { SourceLocationSetTitle } from '../../../contributions/tree/SourceLocationSetTitle'
import { TreeOrComponentViewOptionsProps } from '../../../contributions/tree/TreeOrComponent'
import {
    ComponentLabelsSidebarItem,
    ComponentOwnerSidebarItem,
    ComponentTagsSidebarItem,
} from '../overview/OverviewTab'
import { PersonList } from '../../../components/person-list/PersonList'

interface Props
    extends TelemetryProps,
        ExtensionsControllerProps,
        ThemeProps,
        SettingsCascadeProps,
        Pick<TreeOrComponentViewOptionsProps, 'treeOrComponentViewMode'> {
    repository: RepositoryForTreeFields
    tree: TreeEntryForTreeFields
    component: ComponentDetailFields | null
    sourceLocationSet: TreeOrComponentSourceLocationSetFields
    isTree?: boolean
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
    isTree,
    useHash,
    treeOrComponentViewMode,
    className,
    ...props
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
                    <ComponentTagsSidebarItem component={featuredComponent} />
                </section>
            )}
            {featuredComponent && (
                <section className={SECTION_CLASS_NAME}>
                    <h4 className="font-weight-bold">Owner</h4>
                    <ComponentOwnerSidebarItem component={featuredComponent} isTree={isTree} />
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
                    {/* TODO(sqs): For this, could show a visualization horizontal bar where width = % of person's contributions, bg color is recency of last contribution, and text overlay is the person's name */}
                    <PersonList
                        title="Contributors"
                        titleLink={`${match.url}${pathSeparator}contributors`}
                        titleCount={sourceLocationSet.contributors.totalCount}
                        listTag="ol"
                        orientation="summary"
                        items={sourceLocationSet.contributors.edges.map(contributor => ({
                            person: contributor.person,
                            text:
                                contributor.authoredLineProportion >= 0.01
                                    ? `${(contributor.authoredLineProportion * 100).toFixed(0)}%`
                                    : '<1%',
                            textTooltip: `${contributor.authoredLineCount} ${pluralize(
                                'line',
                                contributor.authoredLineCount
                            )}`,
                            date: contributor.lastCommit.author.date,
                        }))}
                        className={className}
                    />
                </section>
            )}
            <hr className="my-3 d-none" />
            <section className={classNames('d-none', SECTION_CLASS_NAME)}>
                <h4 className="font-weight-bold">
                    Depends on{' '}
                    <Badge variant="secondary" small={true} pill={true} className="ml-1">
                        171
                    </Badge>
                </h4>
                <ul className="list-inline">
                    {'x'
                        .repeat(8)
                        .split(/x/g)
                        .map((_value, index) => (
                            <li key={index} className="list-inline-item mb-1 mr-1">
                                <UserAvatar size={19} user={{ displayName: `user ${index}`, avatarURL: null }} />
                            </li>
                        ))}
                </ul>
            </section>
            <section className={classNames('d-none', SECTION_CLASS_NAME)}>
                <h4 className="font-weight-bold">
                    Used by{' '}
                    {/*                     <Badge variant="secondary" small={true} pill={true} className="ml-1">
                        21
                    </Badge> */}
                </h4>
                <ul className="list-inline">
                    <li className="list-inline-item mb-1 mr-1">
                        <Badge variant="secondary" small={true} pill={true} className="ml-1">
                            28
                        </Badge>{' '}
                        components
                    </li>
                    <li className="list-inline-item mb-1 mr-1">
                        <Badge variant="secondary" small={true} pill={true} className="ml-1">
                            11
                        </Badge>{' '}
                        people
                    </li>
                    {/* {'x'
                        .repeat(8)
                        .split(/x/g)
                        .map((_value, index) => (
                            <li key={index} className="list-inline-item mb-1 mr-1">
                                <UserAvatar size={19} user={{ displayName: `user ${index}`, avatarURL: null }} />
                            </li>
                        ))} */}
                </ul>
            </section>
            <hr className="my-3 d-none" />
            {false && featuredComponent?.labels && featuredComponent.labels.length > 0 && (
                <section className={classNames(SECTION_CLASS_NAME)}>
                    <ComponentLabelsSidebarItem component={featuredComponent} />
                </section>
            )}
        </aside>
    )
}
