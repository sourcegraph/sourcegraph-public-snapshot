import React from "react";
import {Link} from "react-router";
import RepoLink from "sourcegraph/components/RepoLink";
import CSSModules from "react-css-modules";
import styles from "./styles/Dashboard.css";
import base from "sourcegraph/components/styles/_base.css";
import {Input, Panel, Hero, Heading, Button, Icon} from "sourcegraph/components";
import debounce from "lodash/function/debounce";
import GitHubAuthButton from "sourcegraph/components/GitHubAuthButton";

class DashboardRepos extends React.Component {
	static contextTypes = {
		signedIn: React.PropTypes.bool.isRequired,
		githubToken: React.PropTypes.object,
		eventLogger: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);
		this._filterInput = null;
		this._handleFilter = this._handleFilter.bind(this);
		this._handleFilter = debounce(this._handleFilter, 25);
		this._showRepo = this._showRepo.bind(this);
	}

	// _repoSort is a comparison function that sorts more recently
	// pushed repos first.
	_repoSort(a, b) {
		if (a.PushedAt < b.PushedAt) return 1;
		else if (a.PushedAt > b.PushedAt) return -1;
		return 0;
	}

	_handleFilter() {
		this.forceUpdate();
	}

	_showRepo(repo) {
		if (this._filterInput && this._filterInput.value &&
			this._qualifiedName(repo).indexOf(this._filterInput.value.trim().toLowerCase()) === -1) {
			return false;
		}

		return true; // no filter; return all
	}

	_qualifiedName(repo) {
		return (`${repo.Owner}/${repo.Name}`).toLowerCase();
	}

	_hasGithubToken() {
		return this.context && this.context.githubToken;
	}

	render() {
		let repos = this.props.repos.filter(this._showRepo).sort(this._repoSort);

		return (
				<div styleName="bg">
					{this.context.signedIn &&
						<Hero pattern="objects" color="dark" className={base.pv6}>
							<Heading level="2" color="white">My Repositories</Heading>
							<p styleName="cool-pale-gray">Search, browse and cross-reference your own code.</p>
							<Input type="text"
								placeholder="Filter repositories..."
								domRef={(e) => this._filterInput = e}
								onChange={this._handleFilter} />
						</Hero>
					}
					<div styleName="container-fixed" className={base.pb4}>
						{this._hasGithubToken() && repos.length === 0 && (!this._filterInput || !this._filterInput.value) &&
							<Panel className={base.pa5} styleName="tc">Loading...</Panel>
						}

						{this._hasGithubToken() && this._filterInput && this._filterInput.value && repos.length === 0 &&
							<Panel className={base.pa4}>No matching repositories</Panel>
						}

						{!this._hasGithubToken() &&
							<div styleName="bg-white-50 tc br3" className={`${base.pa5} ${base.mt4} ${base.mh4}`}>
								<div styleName="max-width-500" className={base.center}>
									<Icon icon="github" width="120" className={base.mb4} />
									<Heading level="4" className={base.mb4}>
										Uh-oh! You'll need to connect your GitHub account to browse your private code with Sourcegraph.
									</Heading>
									<GitHubAuthButton styleName="inline-block">Connect with GitHub</GitHubAuthButton>
								</div>
							</div>
						}


						<div styleName="repositories">
							{repos.length > 0 && repos.map((repo, i) =>
								<Panel hover={true} key={i} className={`${base.mb4} ${base.pa4}`} styleName="item">
									<div styleName="content">
										<Heading level="3" color="cool-mid-gray">
											<RepoLink repo={repo.URI || `github.com/${repo.Owner}/${repo.Name}`} />
										</Heading>
										{repo.Description && <p styleName="mid-gray" className={base.mb0}>{repo.Description}</p>}
									</div>
									<Link to={repo.URI || `github.com/${repo.Owner}/${repo.Name}`} styleName="button">
										<Button color="blue">Explore code</Button>
									</Link>
								</Panel>
							)}
						</div>
					</div>
				</div>
		);
	}
}

DashboardRepos.propTypes = {
	repos: React.PropTypes.arrayOf(React.PropTypes.object).isRequired,
};

export default CSSModules(DashboardRepos, styles, {allowMultiple: true});
