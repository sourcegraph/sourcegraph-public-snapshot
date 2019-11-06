import * as React from 'react'
import { RouteComponentProps } from 'react-router-dom'
import * as GQL from '../../../../shared/src/graphql/schema'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { GitRefNode, queryGitRefs } from '../GitRef'
import { RepositoryReleasesAreaPageProps } from './RepositoryReleasesArea'
import { Observable } from 'rxjs'

interface Props extends RepositoryReleasesAreaPageProps, RouteComponentProps<{}> {}

/** A page that shows all of a repository's tags. */
export class RepositoryReleasesTagsPage extends React.PureComponent<Props> {
    public componentDidMount(): void {
        eventLogger.logViewEvent('RepositoryReleasesTags')
    }

    public render(): JSX.Element | null {
        return (
            <div className="repository-releases-page">
                <PageTitle title="Tags" />
                <FilteredConnection<GQL.IGitRef>
                    className=""
                    listClassName="list-group list-group-flush"
                    noun="tag"
                    pluralNoun="tags"
                    queryConnection={this.queryTags}
                    nodeComponent={GitRefNode}
                    defaultFirst={20}
                    autoFocus={true}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private queryTags = (args: FilteredConnectionQueryArgs): Observable<GQL.IGitRefConnection> =>
        queryGitRefs({ ...args, repo: this.props.repo.id, type: GQL.GitRefType.GIT_TAG })
}
