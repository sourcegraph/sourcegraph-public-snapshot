import React, { useCallback, useState } from 'react'

import * as H from 'history'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import { Observable } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { isErrorLike, asError, ErrorLike } from '@sourcegraph/common'
import * as GQL from '@sourcegraph/shared/src/schema'
import { Button, Card, CardHeader, Icon, LoadingSpinner, useEventObservable } from '@sourcegraph/wildcard'

import { FilteredConnection, FilteredConnectionQueryArguments } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { GitRefConnectionFields, GitRefFields, GitRefType, TreePageRepositoryFields } from '../../graphql-operations'
import { queryGitBranches } from '../branches/RepositoryBranchesOverviewPage'
import { GitReferenceNode, queryGitReferences } from '../GitReference'

interface Props {
    repo: TreePageRepositoryFields
    location?: H.Location
    history?: H.History
}

interface OverviewTabProps {
    repo: TreePageRepositoryFields
    setShowAll: (spec: boolean) => void
}

interface Data {
    defaultBranch: GQL.IGitRef | null
    activeBranches: GQL.IGitRef[]
    hasMoreActiveBranches: boolean
}

/**
 * Renders pages related to repository branches.
 */
export const RepositoryBranchesTab: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    repo,
    history,
    location,
}) => {
    const [showAll, setShowAll] = useState(false)

    return (
        <div className="repository-branches-area container">
            <ul className="nav my-3">
                <li className="nav-item">
                    <Button
                        onClick={() => setShowAll(false)}
                        type="button"
                        variant="link"
                        outline={!showAll}
                        disabled={!showAll}
                    >
                        Overview
                    </Button>
                </li>
                <li className="nav-item">
                    <Button
                        onClick={() => setShowAll(true)}
                        type="button"
                        variant="link"
                        outline={showAll}
                        disabled={showAll}
                    >
                        All branches
                    </Button>
                </li>
            </ul>
            {showAll ? (
                <RepositoryBranchesAllTab repo={repo} location={location} history={history} />
            ) : (
                <RepositoryBranchesOverviewTab repo={repo} setShowAll={setShowAll} />
            )}
        </div>
    )
}

export const RepositoryBranchesAllTab: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    repo,
    history,
    location,
}) => {
    const queryBranches = (args: FilteredConnectionQueryArguments): Observable<GitRefConnectionFields> =>
        queryGitReferences({ ...args, repo: repo.id, type: GitRefType.GIT_BRANCH })

    return (
        <div>
            <PageTitle title="All branches" />
            <FilteredConnection<GitRefFields>
                listClassName="list-group list-group-flush"
                noun="branch"
                pluralNoun="branches"
                queryConnection={queryBranches}
                nodeComponent={GitReferenceNode}
                defaultFirst={20}
                autoFocus={true}
                history={history}
                location={location}
            />
        </div>
    )
}

export const RepositoryBranchesOverviewTab: React.FunctionComponent<React.PropsWithChildren<OverviewTabProps>> = ({
    repo,
    setShowAll,
}) => {
    const [branches, setBranches] = useState<Data | undefined>(undefined)

    useEventObservable<void, Data | null | ErrorLike>(
        useCallback(
            (clicks: Observable<void>) =>
                clicks.pipe(
                    mapTo(true),
                    startWith(false),
                    switchMap(() => queryGitBranches({ repo: repo.id, first: 10 })),
                    map(branch => {
                        if (branch === null) {
                            return branch
                        }

                        if (branch) {
                            setBranches(branch)
                        }

                        return branch
                    }),
                    catchError((error): [ErrorLike] => [asError(error)])
                ),
            [repo.id]
        )
    )

    return (
        <div>
            <PageTitle title="Branches" />
            {branches === undefined ? (
                <LoadingSpinner className="mt-2" />
            ) : isErrorLike(branches) ? (
                <ErrorAlert className="mt-2" error={branches} />
            ) : (
                <div>
                    {branches.defaultBranch && (
                        <Card className="card">
                            <CardHeader>Default branch</CardHeader>
                            <ul className="list-group list-group-flush">
                                <GitReferenceNode node={branches.defaultBranch} />
                            </ul>
                        </Card>
                    )}
                    {branches.activeBranches.length > 0 && (
                        <Card className="card">
                            <CardHeader>Active branches</CardHeader>
                            <div className="list-group list-group-flush">
                                {branches.activeBranches.map((gitReference, index) => (
                                    <GitReferenceNode key={index} node={gitReference} />
                                ))}
                                {branches.hasMoreActiveBranches && (
                                    <Button
                                        type="button"
                                        variant="secondary"
                                        onClick={() => setShowAll(true)}
                                        className="list-group-item list-group-item-action py-2 d-flex"
                                    >
                                        View more branches
                                        <Icon role="img" as={ChevronRightIcon} aria-hidden={true} />
                                    </Button>
                                )}
                            </div>
                        </Card>
                    )}
                </div>
            )}
        </div>
    )
}
