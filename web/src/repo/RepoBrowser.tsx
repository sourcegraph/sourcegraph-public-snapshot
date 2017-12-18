import groupBy from 'lodash/groupBy'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { fetchRepositories } from './backend'

interface State {
    repositories?: GQL.IRepository[]
}

export class RepoBrowser extends React.PureComponent<{}, State> {
    public state: State = {}
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            fetchRepositories().subscribe(
                ({ nodes }) => {
                    this.setState({ repositories: nodes })
                },
                err => console.error(err)
            )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element {
        return (
            <div className="repo-browser">
                <PageTitle title="Browse repositories" />
                <h2>Repositories</h2>
                {!this.state.repositories && <span>Loading...</span>}
                {this.state.repositories && this.repoGroups()}
                {this.state.repositories && (
                    <h3>
                        <a
                            className="repo-browser__add-more"
                            href="https://about.sourcegraph.com/docs/server/install/#configure-sourcegraph-server-to-index-your-code-host"
                        >
                            Add more
                        </a>
                    </h3>
                )}
            </div>
        )
    }

    private repoGroups = () => {
        const groups = groupBy(this.state.repositories, repo => repo.uri.split('/')[0])
        const hosts = Object.keys(groups).sort()
        return hosts.map((host, i) => (
            <div className="repo-browser__group" key={i}>
                <h3>{host}</h3>
                {groups[host].map((repo, j) => (
                    <div key={j}>
                        <Link to={'/' + repo.uri}>
                            {repo.uri // remove first path component
                                .split('/')
                                .slice(1)
                                .join('/')}
                        </Link>
                    </div>
                ))}
            </div>
        ))
    }
}
