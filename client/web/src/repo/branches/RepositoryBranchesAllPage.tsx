import * as React from 'react'

import { RouteComponentProps } from 'react-router-dom'
import { Observable } from 'rxjs'

import { FilteredConnection, FilteredConnectionQueryArguments } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { GitRefType, GitRefConnectionFields, GitRefFields } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'
import { GitReferenceNode, queryGitReferences } from '../GitReference'

import { RepositoryBranchesAreaPageProps } from './RepositoryBranchesArea'

interface Props extends RepositoryBranchesAreaPageProps, RouteComponentProps<{}> {}

/** A page that shows all of a repository's branches. */
export class RepositoryBranchesAllPage extends React.PureComponent<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('RepositoryBranchesAll')
    }

    public render(): JSX.Element | null {
        return (
            <div>
                <PageTitle title="All branches" />
                <FilteredConnection<GitRefFields>
                    className=""
                    listClassName="list-group list-group-flush"
                    noun="branch"
                    pluralNoun="branches"
                    queryConnection={this.queryBranches}
                    nodeComponent={GitReferenceNode}
                    defaultFirst={20}
                    autoFocus={true}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private queryBranches = (args: FilteredConnectionQueryArguments): Observable<GitRefConnectionFields> =>
        queryGitReferences({ ...args, repo: this.props.repo.id, type: GitRefType.GIT_BRANCH })
}
