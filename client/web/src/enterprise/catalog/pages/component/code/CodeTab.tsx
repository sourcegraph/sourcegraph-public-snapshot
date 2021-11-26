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
    ComponentStateDetailFields,
    RepositoryForTreeFields,
    TreeEntryForTreeFields,
    TreeOrComponentSourceLocationSetFields,
} from '../../../../../graphql-operations'
import { RepoRevisionContainerBreadcrumb } from '../../../../../repo/RepoRevisionContainer'
import { SourceLocationSetTitle } from '../../../contributions/tree/SourceLocationSetTitle'
import { TreeOrComponentViewOptionsProps } from '../../../contributions/tree/TreeOrComponent'
import { SourceLocationSetCommits } from '../ComponentCommits'
import { ComponentSourceLocations } from '../ComponentSourceLocations'
import { ComponentReadme, LastCommit } from '../overview/ComponentOverviewMain'
import { SourceLocationSetContributors } from '../SourceLocationSetContributors'
import { SourceLocationSetTreeEntries } from '../SourceLocationSetTreeEntries'

import { SourceLocationSetCodeOwners } from './CodeOwners'
import { CodeTabSidebar } from './CodeTabSidebar'
import { SourceLocationSetBranches } from './SourceLocationSetBranches'
import { SourceLocationSetSelectMenu } from './SourceLocationSetSelectMenu'

interface Props
    extends TelemetryProps,
        ExtensionsControllerProps,
        ThemeProps,
        SettingsCascadeProps,
        Pick<TreeOrComponentViewOptionsProps, 'treeOrComponentViewMode' | 'treeOrComponentViewModeURL'> {
    repository: RepositoryForTreeFields
    tree: TreeEntryForTreeFields
    component: ComponentStateDetailFields | null
    sourceLocationSet: TreeOrComponentSourceLocationSetFields
    isTree?: boolean
    useHash?: boolean
    className?: string
}

export const CodeTab: React.FunctionComponent<Props> = ({
    repository,
    tree,
    component,
    sourceLocationSet,
    treeOrComponentViewMode,
    treeOrComponentViewModeURL,
    isTree,
    useHash,
    className,
    ...props
}) => {
    const match = useRouteMatch()
    const location = useLocation()
    const pathSeparator = useHash ? '#' : '/'

    return (
        <div className={classNames('row flex-wrap-reverse', className)}>
            <Switch
                /* TODO(sqs): hack to make the router work with hashes */
                location={useHash ? { ...location, pathname: location.pathname + location.hash } : undefined}
            >
                <Route path={match.url} exact={true}>
                    <div className="col-md-9">
                        {!isTree && (
                            <>
                                <h4 className="sr-only">Sources</h4>
                                <ComponentSourceLocations component={component} className="mb-3" />
                            </>
                        )}
                        <div className="pb-2 d-flex align-items-center">
                            <h2 className="d-flex align-items-center h6 mb-0">
                                <SourceLocationSetTitle
                                    component={component}
                                    tree={tree}
                                    treeOrComponentViewMode={treeOrComponentViewMode}
                                />
                                {component !== null && treeOrComponentViewMode === 'auto' && (
                                    <>
                                        <span className="text-muted mx-1">in</span>
                                        <SourceLocationSetTitle
                                            component={null}
                                            tree={tree}
                                            treeOrComponentViewMode={treeOrComponentViewMode}
                                        />
                                    </>
                                )}
                            </h2>
                            {component && (
                                <SourceLocationSetSelectMenu
                                    treeOrComponentViewMode={treeOrComponentViewMode}
                                    treeOrComponentViewModeURL={treeOrComponentViewModeURL}
                                    buttonClassName="px-2 py-1 text-muted"
                                />
                            )}
                            {false && (
                                <div>
                                    <RepoRevisionContainerBreadcrumb
                                        // TODO(sqs): actually use the passed-down values for these
                                        revision="main"
                                        resolvedRevisionOrError={{
                                            commitID: '4aaa77a86c83a5802b355c64679223e327280bbf',
                                            defaultBranch: 'main',
                                            rootTreeURL: 'TODO(sqs)',
                                        }}
                                        repo={repository}
                                        small={false}
                                    />
                                </div>
                            )}
                            <Link to={`${match.url}${pathSeparator}branches`} className="ml-3">
                                {sourceLocationSet.branches.totalCount}{' '}
                                {pluralize('branch', sourceLocationSet.branches.totalCount, 'branches')}
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
                            {sourceLocationSet.commits?.nodes[0] && (
                                <LastCommit
                                    commit={sourceLocationSet.commits?.nodes[0]}
                                    after={
                                        <Link to={`${match.url}${pathSeparator}commits`} className="ml-3 text-nowrap">
                                            All commits
                                        </Link>
                                    }
                                    className="card-body border-bottom p-3"
                                />
                            )}
                            {/* TODO(sqs): if a component, show a UI indication to the effect of "Also includes sources from other paths: ..." */}
                            {(sourceLocationSet.__typename === 'Component' ||
                                sourceLocationSet.__typename === 'GitTree') && (
                                <SourceLocationSetTreeEntries
                                    {...props}
                                    sourceLocationSet={sourceLocationSet}
                                    className="card-body"
                                />
                            )}
                        </div>
                        {sourceLocationSet.readme && <ComponentReadme readme={sourceLocationSet.readme} />}
                    </div>
                    <div className="col-md-3">
                        <CodeTabSidebar
                            repository={repository}
                            tree={tree}
                            component={component}
                            sourceLocationSet={sourceLocationSet}
                            treeOrComponentViewMode={treeOrComponentViewMode}
                            isTree={isTree}
                            useHash={useHash}
                        />
                    </div>
                </Route>
                <Route path={`${match.url}${pathSeparator}contributors`}>
                    <SourceLocationSetContributors sourceLocationSet={sourceLocationSet.id} className="mb-3" />
                </Route>
                <Route path={`${match.url}${pathSeparator}code-owners`}>
                    <SourceLocationSetCodeOwners sourceLocationSet={sourceLocationSet.id} className="mb-3" />
                </Route>
                <Route path={`${match.url}${pathSeparator}commits`}>
                    <SourceLocationSetCommits sourceLocationSet={sourceLocationSet.id} className="mb-3 card w-100" />
                </Route>
                <Route path={`${match.url}${pathSeparator}branches`}>
                    <SourceLocationSetBranches sourceLocationSet={sourceLocationSet.id} className="w-100" />
                </Route>
            </Switch>
        </div>
    )
}
