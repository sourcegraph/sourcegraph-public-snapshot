import classNames from 'classnames'
import SearchIcon from 'mdi-react/SearchIcon'
import React from 'react'
import { Route, Switch, useLocation, useRouteMatch } from 'react-router'
import { Link } from 'react-router-dom'

import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { pluralize } from '@sourcegraph/shared/src/util/strings'

import {
    RepositoryForTreeFields,
    TreeEntryForTreeFields,
    SourceSetAtTreeFields,
} from '../../../../../graphql-operations'
import { SourceSetTitle } from '../../../contributions/tree/SourceSetTitle'
import { SourceSetDescendentComponents } from '../../source-set-at-tree/SourceSetDescendentComponents'
import { SourceSetAtTreeViewOptionsProps } from '../../source-set-at-tree/useSourceSetAtTreeViewOptions'
import { SourceSetReadme } from '../readme/ComponentReadme'

import { SourceSetCodeOwners } from './CodeOwners'
import { CodeTabSidebar } from './CodeTabSidebar'
import { SourceSetCommits } from './ComponentCommits'
import { LastCommit } from './LastCommit'
import { SourceSetBranches } from './SourceSetBranches'
import { SourceSetContributors } from './SourceSetContributors'
import { SourceSetSelectMenu } from './SourceSetSelectMenu'
import { SourceSetTreeEntries } from './SourceSetTreeEntries'

interface Props
    extends TelemetryProps,
        ExtensionsControllerProps,
        ThemeProps,
        SettingsCascadeProps,
        Pick<SourceSetAtTreeViewOptionsProps, 'sourceSetAtTreeViewMode' | 'sourceSetAtTreeViewModeURL'> {
    repository: RepositoryForTreeFields
    tree: TreeEntryForTreeFields
    component: React.ComponentPropsWithoutRef<typeof CodeTabSidebar>['component'] | null
    sourceSet: SourceSetAtTreeFields
    useHash?: boolean
    className?: string
}

export const CodeTab: React.FunctionComponent<Props> = ({
    repository,
    tree,
    component,
    sourceSet,
    sourceSetAtTreeViewMode,
    sourceSetAtTreeViewModeURL,
    useHash,
    className,
    ...props
}) => {
    const match = useRouteMatch()
    const location = useLocation()
    const pathSeparator = useHash ? '#' : '/'

    return (
        <div className={classNames('row flex-wrap-reverse ', className)}>
            <Switch
                /* TODO(sqs): hack to make the router work with hashes */
                location={useHash ? { ...location, pathname: location.pathname + location.hash } : undefined}
            >
                <Route path={match.url} exact={true}>
                    <div className="col-md-9">
                        <div className="pb-2 d-flex align-items-center">
                            <h2 className="d-flex align-items-center h6 mb-0">
                                <SourceSetTitle
                                    component={component}
                                    tree={tree}
                                    sourceSetAtTreeViewMode={sourceSetAtTreeViewMode}
                                />
                                {component !== null && sourceSetAtTreeViewMode === 'auto' && (
                                    <>
                                        <span className="text-muted mx-1">in</span>
                                        <SourceSetTitle
                                            component={null}
                                            tree={tree}
                                            sourceSetAtTreeViewMode={sourceSetAtTreeViewMode}
                                        />
                                    </>
                                )}
                            </h2>
                            {component && (
                                <SourceSetSelectMenu
                                    sourceSetAtTreeViewMode={sourceSetAtTreeViewMode}
                                    sourceSetAtTreeViewModeURL={sourceSetAtTreeViewModeURL}
                                    buttonClassName="px-2 py-1 text-muted"
                                />
                            )}
                            <Link to={`${match.url}${pathSeparator}branches`} className="ml-3">
                                {sourceSet.branches.totalCount}{' '}
                                {pluralize('branch', sourceSet.branches.totalCount, 'branches')}
                            </Link>
                            {tree.isRoot && (
                                <Link to={`${match.url}${pathSeparator}tags`} className="ml-3">
                                    Tags
                                </Link>
                            )}
                            <div className="flex-1" />
                            <Link
                                to="/search?q=TODO"
                                className="d-inline-flex align-items-center btn btn-sm btn-outline-secondary"
                            >
                                <SearchIcon className="icon-inline mr-1" /> Search...
                            </Link>
                        </div>
                        <div className="card mb-3">
                            {sourceSet.commitsForLastCommit?.nodes[0] && (
                                <LastCommit
                                    commit={sourceSet.commitsForLastCommit?.nodes[0]}
                                    after={
                                        <Link to={`${match.url}${pathSeparator}commits`} className="ml-3 text-nowrap">
                                            All commits
                                        </Link>
                                    }
                                    className="card-body border-bottom p-3"
                                />
                            )}
                            {/* TODO(sqs): if a component, show a UI indication to the effect of "Also includes sources from other paths: ..." */}
                            {(sourceSet.__typename === 'Component' || sourceSet.__typename === 'GitTree') && (
                                <SourceSetTreeEntries {...props} sourceSet={sourceSet} className="card-body" />
                            )}
                        </div>
                        <SourceSetDescendentComponents
                            descendentComponents={sourceSet.descendentComponents}
                            repoID={repository.id}
                            filePath={tree.path}
                        />
                        {sourceSet.readme && <SourceSetReadme readme={sourceSet.readme} />}
                    </div>
                    <div className="col-md-3">
                        <CodeTabSidebar
                            repository={repository}
                            tree={tree}
                            component={component}
                            sourceSet={sourceSet}
                            sourceSetAtTreeViewMode={sourceSetAtTreeViewMode}
                            useHash={useHash}
                        />
                    </div>
                </Route>
                <Route path={`${match.url}${pathSeparator}contributors`}>
                    <SourceSetContributors sourceSet={sourceSet.id} className="mb-3" />
                </Route>
                <Route path={`${match.url}${pathSeparator}code-owners`}>
                    <SourceSetCodeOwners sourceSet={sourceSet.id} className="mb-3" />
                </Route>
                <Route path={`${match.url}${pathSeparator}commits`}>
                    <SourceSetCommits sourceSet={sourceSet.id} className="mb-3 card w-100" />
                </Route>
                <Route path={`${match.url}${pathSeparator}branches`}>
                    <SourceSetBranches sourceSet={sourceSet.id} className="w-100" />
                </Route>
            </Switch>
        </div>
    )
}
