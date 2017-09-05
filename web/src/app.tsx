import * as React from 'react';
import { render } from 'react-dom';
import { BrowserRouter, Route, RouteComponentProps, Switch } from 'react-router-dom';
import 'rxjs/add/observable/fromPromise';
import 'rxjs/add/operator/catch';
import { Observable } from 'rxjs/Observable';
import { Subject } from 'rxjs/Subject';
import { Subscription } from 'rxjs/Subscription';
import { Home } from 'sourcegraph/home/Home';
import { Navbar } from 'sourcegraph/nav/Navbar';
import { makeRepoURI, RepoURI } from 'sourcegraph/repo';
import { resolveRev } from 'sourcegraph/repo/backend';
import { Repository } from 'sourcegraph/repo/Repository';
import { SearchResults } from 'sourcegraph/search/SearchResults';
import * as activeRepos from 'sourcegraph/util/activeRepos';
import { parseHash } from 'sourcegraph/util/url';

window.addEventListener('DOMContentLoaded', () => {
    // Be a bit proactive and try to fetch/store active repos now. This helps
    // on the first search query, and when the data in local storage is stale.
    activeRepos.get().catch(err => console.error(err));
});

interface WithResolvedRevProps {
    component: any;
    uri: RepoURI;
    repoPath: string;
    rev?: string;
    [key: string]: any;
}

interface WithResolvedRevState {
    commitID?: string;
}

class WithResolvedRev extends React.Component<WithResolvedRevProps, WithResolvedRevState> {
    public state: WithResolvedRevState = {};
    private componentUpdates = new Subject<WithResolvedRevProps>();
    private subscriptions = new Subscription();

    constructor(props: WithResolvedRevProps) {
        super(props);
        this.subscriptions.add(
            this.componentUpdates
                // tslint:disable-next-line
                .switchMap(props =>
                    Observable.fromPromise(resolveRev(props))
                        .catch(err => {
                            console.error(err);
                            return [];
                        })
                )
                .subscribe(resolved => {
                    const commitID = resolved.commitID;
                    if (commitID) {
                        this.setState(resolved);
                    }
                }, err => {
                    console.error(err);
                })
        );
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props);
    }

    public componentWillReceiveProps(nextProps: WithResolvedRevProps): void {
        if (this.props.repoPath !== nextProps.repoPath || this.props.rev !== nextProps.rev) {
            // clear state so the child won't render until the revision is resolved for new props
            this.state = {};
            this.componentUpdates.next(nextProps);
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe();
    }

    public render(): JSX.Element | null {
        if (!this.state.commitID) {
            return null;
        }
        return <this.props.component {...this.props} commitID={this.state.commitID} />;
    }
}

class AppRouter extends React.Component<RouteComponentProps<string[]>, {}> {
    public render(): JSX.Element | null {
        if (!this.props.match.params[0]) {
            return null;
        }
        if (this.props.match.params[0] === 'search') {
            return <SearchResults />;
        }

        const uriPathSplit = this.props.match.params[0].split('/-/');
        const repoRevSplit = uriPathSplit[0].split('@');
        const hash = parseHash(this.props.location.hash);
        const position = hash.line ? { line: hash.line, char: hash.char } : undefined;
        const repoParams = { repoPath: repoRevSplit[0], rev: repoRevSplit[1], position };
        if (uriPathSplit.length === 1) {
            return <WithResolvedRev uri={makeRepoURI(repoParams)} {...repoParams} history={this.props.history} location={this.props.location} component={Repository} />;
        }
        const filePath = uriPathSplit[1].split('/').slice(1).join('/'); // e.g. '[blob|tree]/path/to/file/or/directory'; ignore the first component
        return <WithResolvedRev {...repoParams}
            filePath={filePath}
            uri={makeRepoURI({...repoParams, filePath})}
            history={this.props.history}
            location={this.props.location}
            component={Repository} />;
    }
}

/**
 * Defines the layout of all pages that have a navbar
 */
class Layout extends React.Component<RouteComponentProps<string[]>, {}> {
    public render(): JSX.Element | null {
        return (
            <div className='layout'>
                <Navbar {...this.props} />
                <div className='layout__app-router-container'>
                    <AppRouter {...this.props} />
                </div>
            </div>
        );
    }
}

/**
 * The root component
 */
class App extends React.Component<{}, {}> {
    public render(): JSX.Element | null {
        return <BrowserRouter>
            <Switch>
                <Route exact path='/' component={Home} />
                <Route path='/*' component={Layout} />
            </Switch>
        </BrowserRouter>;
    }
}

window.addEventListener('DOMContentLoaded', () => {
    render(<App />, document.querySelector('#root'));
});
