import * as React from 'react';
import { render } from 'react-dom';
import { BrowserRouter, Route, RouteComponentProps, Switch } from 'react-router-dom';
import { resolveRev } from 'sourcegraph/backend';
import * as xhr from 'sourcegraph/backend/xhr';
import { Home } from 'sourcegraph/home';
import { Navbar } from 'sourcegraph/nav';
import { Repository } from 'sourcegraph/repo/Repository';
import { SearchResults } from 'sourcegraph/search/SearchResults';
import * as activeRepos from 'sourcegraph/util/activeRepos';
import { sourcegraphContext } from 'sourcegraph/util/sourcegraphContext';

window.addEventListener('DOMContentLoaded', () => {
    xhr.useAccessToken(sourcegraphContext.accessToken);

    // Be a bit proactive and try to fetch/store active repos now. This helps
    // on the first search query, and when the data in local storage is stale.
    activeRepos.get().catch(err => console.error(err));

});

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

class Layout extends React.Component<RouteComponentProps<string[]>, {}> {
    public render(): JSX.Element | null {
        return <div>
            <Navbar {...this.props} />
            <div id='app-container'>
                <AppRouter {...this.props} />
            </div>
        </div>;
    }
}

interface WithResolvedRevProps {
    component: any;
    uri: string;
    rev?: string;
    [key: string]: any;
}

interface WithResolvedRevState {
    commitID?: string;
}

class WithResolvedRev extends React.Component<WithResolvedRevProps, WithResolvedRevState> {
    public state: WithResolvedRevState = {};

    constructor(props: WithResolvedRevProps) {
        super(props);
        resolveRev(this.props.uri, this.props.rev).then(resp => this.setState({ commitID: resp.commitID }));
    }

    public componentWillReceiveProps(nextProps: WithResolvedRevProps): void {
        if (this.props.uri !== nextProps.uri || this.props.rev !== nextProps.rev) {
            // clear state so the child won't render until the revision is resolved for new props
            this.state = {};
            resolveRev(nextProps.uri, nextProps.rev).then(resp => this.setState({ commitID: resp.commitID }));
        }
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
            return <div className='search'><SearchResults /></div>;
        }

        const uriPathSplit = this.props.match.params[0].split('/-/');
        const repoRevSplit = uriPathSplit[0].split('@');
        if (uriPathSplit.length === 1) {
            return <WithResolvedRev uri={repoRevSplit[0]} rev={repoRevSplit[1]} history={this.props.history} location={this.props.location} component={Repository} />;
        }
        const path = uriPathSplit[1]; // e.g. '[blob|tree]/path/to/file/or/directory'; ignore the first component
        return <WithResolvedRev
            uri={repoRevSplit[0]}
            rev={repoRevSplit[1]}
            path={path.split('/').slice(1).join('/')}
            history={this.props.history}
            location={this.props.location}
            component={Repository} />;
    }
}

window.addEventListener('DOMContentLoaded', () => {
    render(<App />, document.querySelector('#root'));
});
