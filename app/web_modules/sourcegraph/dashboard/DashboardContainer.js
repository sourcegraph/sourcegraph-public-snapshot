import React from "react";
import Helmet from "react-helmet";
import {Link} from "react-router";

import Container from "sourcegraph/Container";
import Dispatcher from "sourcegraph/Dispatcher";
import "./DashboardBackend"; // for side effects
import DashboardStore from "sourcegraph/dashboard/DashboardStore";
import DashboardRepos from "sourcegraph/dashboard/DashboardRepos";
import EventLogger, {EventLocation} from "sourcegraph/util/EventLogger";
import * as DashboardActions from "sourcegraph/dashboard/DashboardActions";
import context from "sourcegraph/app/context";

import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";

import {Button} from "sourcegraph/components";
import {GitHubIcon} from "sourcegraph/components/Icons";
import {urlToGitHubOAuth} from "sourcegraph/util/urlTo";

import deepFreeze from "sourcegraph/util/deepFreeze";

class DashboardContainer extends Container {
	constructor(props) {
		super(props);
		this.state = {
			numRepos: 79272,
			numFunctions: 143203753,
		};
	}

	componentDidMount() {
		super.componentDidMount();
		if (this.state.githubRedirect) {
			EventLogger.logEvent("LinkGitHubCompleted");
		}

		// Dummy counters for repos and functions.
		this._counter = setInterval(() => {
			this.setState({
				numRepos: this.state.numRepos + Math.floor(Math.random() * 2),
				numFunctions: this.state.numFunctions + Math.floor(Math.random() * 10),
			});
		}, 2500);
	}

	componentWillUnmount() {
		super.componentWillUnmount();
		clearInterval(this._counter);
	}

	reconcileState(state, props) {
		Object.assign(state, props);
		state.exampleRepos = this._exampleRepos();
		state.repos = DashboardStore.repos || null;
		state.remoteRepos = DashboardStore.remoteRepos || null;
		state.githubRedirect = props.location && props.location.query ? (props.location.query["github-onboarding"] || false) : false;
	}

	onStateTransition(prevState, nextState) {
		if (nextState.repos === null && nextState.repos !== prevState.repos) {
			Dispatcher.Backends.dispatch(new DashboardActions.WantRepos());
		}
		if (nextState.remoteRepos === null && nextState.remoteRepos !== prevState.remoteRepos) {
			Dispatcher.Backends.dispatch(new DashboardActions.WantRemoteRepos());
		}
	}

	stores() { return [DashboardStore]; }

	_exampleRepos() {
		return deepFreeze([{
			URI: "github.com/golang/go",
			Owner: "golang",
			Name: "go",
			Language: "Go",
			Examples: [
				{Functions: {
					Path: "/github.com/golang/go@master/-/def/GoPackage/net/http/-/Get",
					FunctionCallCount: "2313",
					FmtStrings: {
						Name: {
							ScopeQualified: "http.Get"}, Type: {
								ScopeQualified: "(url string) (resp *Response, err error)",
							}, NameAndTypeSeparator: "", DefKeyword: "func"}},
				},
				{Functions: {
					Path: "/github.com/golang/go@master/-/def/GoPackage/fmt/-/Sprintf",
					FunctionCallCount: "1313",
					FmtStrings: {
						Name: {ScopeQualified: "fmt.Sprintf"},
						Type: {ScopeQualified: "(format string, a ...interface{}) string"}, NameAndTypeSeparator: "", DefKeyword: "func",
					},
				},
				},
			],
		},
		{
			URI: "github.com/gorilla/mux",
			Owner: "gorilla",
			Name: "mux",
			Language: "Go",
			Examples: [
				{Functions: {
					Path: "/github.com/gorilla/mux@master/-/def/GoPackage/github.com/gorilla/mux/-/NewRouter",
					FunctionCallCount: "40",
					FmtStrings: {
						Name: {ScopeQualified: "NewRouter"},
						Type: {ScopeQualified: "()*Router"},
						NameAndTypeSeparator: "",
						DefKeyword: "func",
					},
				},
			},
				{Functions: {
					Path: "/github.com/gorilla/mux@master/-/def/GoPackage/github.com/gorilla/mux/-/Route/PathPrefix",
					FunctionCallCount: "14",
					FmtStrings: {
						Name: {ScopeQualified: "(r *Route) PathPrefix"},
						Type: {ScopeQualified: "(tpl string) *Route"},
						NameAndTypeSeparator: "",
						DefKeyword: "func",
					},
				},
			},
			]},
			]);
	}

	render() {
		return (<div styleName="container">
			<Helmet title="Home" />

			{!context.currentUser &&
				<div styleName="action">
					<Link to="/login"> Sign in </Link>
				</div>
			}

			{!context.currentUser &&
				<div styleName="anon-section">
					<img styleName="logo" src={`${context.assetsRoot || ""}/img/sourcegraph-logo.svg`}/>
					<div styleName="anon-title">Code Intelligence for Teams</div>
					<div styleName="anon-header-sub">Search, browse, and cross-reference code</div>
					<div styleName="anon-header-counts">{this.state.numRepos.toLocaleString()} repositories &middot; {this.state.numFunctions.toLocaleString()} functions</div>
				</div>
			}

			{!context.hasLinkedGitHub && context.currentUser &&
				<div styleName="header">
					<span styleName="cta">
						<a href={urlToGitHubOAuth} onClick={() => EventLogger.logEventForPage("SubmitLinkGitHub", EventLocation.Dashboard)}>
						<Button outline={true} color="warning"><GitHubIcon styleName="github-icon" />Add My GitHub Repositories</Button>
						</a>
					</span>
				</div>
			}

			<div styleName="repos">
				<DashboardRepos repos={(this.state.repos || []).concat(this.state.remoteRepos || [])}
					exampleRepos={this.state.exampleRepos}/>
			</div>

			{!context.currentUser &&
				<div styleName="cta-box">
					<span styleName="cta">
						<a href="join" onClick={() => EventLogger.logEventForPage("JoinCTAClicked", EventLocation.Dashboard)}>
							<Button color="info" size="large">Add Sourcegraph to my code</Button>
						</a>
					</span>
				</div>
			}

		</div>);
	}
}

DashboardContainer.propTypes = {
};


export default CSSModules(DashboardContainer, styles);
