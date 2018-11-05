import * as React from 'react'
import { RouteComponentProps } from 'react-router-dom'
import * as GQL from '../../backend/graphqlschema'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { GitRefNode, queryGitRefs } from '../GitRef'
import { RepositoryBranchesAreaPageProps } from './RepositoryBranchesArea'

interface Props extends RepositoryBranchesAreaPageProps, RouteComponentProps<{}> {}

/** A page that shows all of a repository's branches. */
export class RepositoryBranchesAllPage extends React.PureComponent<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('RepositoryBranchesAll')
    }

    public render(): JSX.Element | null {
        return (
            <div className="repository-branches-page">
                <PageTitle title="All branches" />
                <FilteredConnection<GQL.IGitRef>
                    className=""
                    listClassName="list-group list-group-flush"
                    noun="branch"
                    pluralNoun="branches"
                    queryConnection={this.queryBranches}
                    nodeComponent={GitRefNode}
                    defaultFirst={20}
                    autoFocus={true}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private queryBranches = (args: FilteredConnectionQueryArgs) =>
        queryGitRefs({ ...args, repo: this.props.repo.id, type: GQL.GitRefType.GIT_BRANCH })
}
