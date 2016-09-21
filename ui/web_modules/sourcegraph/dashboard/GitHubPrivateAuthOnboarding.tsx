// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as base from "sourcegraph/components/styles/_base.css";
import * as classNames from "classnames";
import * as colors from "sourcegraph/components/styles/_colors.css";
import * as styles from "sourcegraph/dashboard/styles/Dashboard.css";
import * as typography from "sourcegraph/components/styles/_typography.css";
import Helmet from "react-helmet";
import {Button, Heading, Panel, RepoLink, ToggleSwitch, Loader} from "sourcegraph/components";
import {GitHubAuthButton} from "sourcegraph/components/GitHubAuthButton";
import {privateGitHubOAuthScopes} from "sourcegraph/util/urlTo";
import * as Dispatcher from "sourcegraph/Dispatcher";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";

interface Props {
	location?: any;
	privateCodeAuthed?: any;
	repos: any[];
	completeStep?: any;
}

type State = any;

export class GitHubPrivateAuthOnboarding extends React.Component<Props, State> {
	static contextTypes: React.ValidationMap<any> = {
		siteConfig: React.PropTypes.object.isRequired,
		router: React.PropTypes.object.isRequired,
	};

	constructor(props) {
		super(props);

		this.state = {
			showAll: false,
			isLoading: this.props.location.search.includes("CompletedGitHubOAuth2Flow"),
		};
	}

	_renderPrivateAuthCTA(): JSX.Element | null {
		return (
			<div>
				<Helmet title="Home" />
				<div className={styles.onboarding_container}>
					<Panel className={classNames(base.pb3, base.ph4, base.ba, base.br2, colors.b__cool_pale_gray)}>
						<Heading pt={4} align="center" level={3}>
							Browse your private code with Sourcegraph
						</Heading>
						<div className={styles.user_actions} style={{maxWidth: "380px"}}>
							<p className={classNames(typography.tc, base.mt3, base.mb2, typography.f6, colors.cool_gray_8)} >
								Enable Sourcegraph on any private GitHub repositories for a better coding experience
							</p>
							<div className={classNames(base.pv5)}>
								<img width={332} style={{marginBottom: "-95px"}} src={`${(this.context as any).siteConfig.assetsRoot}/img/Dashboard/OnboardingRepos.png`}></img>
								{this.state.isLoading && <div><Loader/></div>}
								{!this.state.isLoading && <GitHubAuthButton pageName={"GitHubPrivateCodeOnboarding"} scopes={privateGitHubOAuthScopes} returnTo={this.props.location} className={styles.github_button}>Add private repositories</GitHubAuthButton>}
							</div>
							<p>
								<a onClick={this._skipClicked.bind(this)}>Skip</a>
							</p>
						</div>
					</Panel>
				</div>
			</div>
		);
	}

	// _repoSort is a comparison function that sorts more recently
	// pushed repos first.
	_repoSort(a: any, b: any): number {
		if (a.PushedAt < b.PushedAt) {
			return 1;
		}
		if (a.PushedAt > b.PushedAt) {
			return -1;
		}

		return 0;
	}

	_qualifiedName(repo: any): string {
		return (`${repo.Owner}/${repo.Name}`).toLowerCase();
	}

	_toggleRepo(remoteRepo: any): void {
		EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_TOGGLE, "IndexRepoToggleClicked", {page_name: "GitHubPrivateCodeOnboarding"});
		Dispatcher.Backends.dispatch(new RepoActions.WantCreateRepo(remoteRepo.URI, remoteRepo, true));
	}

	_skipClicked() {
		EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_CLICK, "SkipGitHubPrivateAuth", {page_name: "GitHubPrivateCodeOnboarding"});
		this.props.completeStep();
	}

	_showAllReposClicked() {
		EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_CLICK, "ShowAllGitHubRepositories", {page_name: "GitHubPrivateCodeOnboarding"});
		this.setState({
			showAll: true,
		});
	}

	_renderIndexRepoCTA(): JSX.Element | null {
		let repos: any[] = [];
		if (this.props.repos) {
			if (this.state.showAll) {
				repos = this.props.repos.slice(0).sort(this._repoSort);
			} else {
				repos = this.props.repos.slice(0).sort(this._repoSort).slice(0, 5);
			}
		}

		return (
			<div>
				<Helmet title="Home" />
				<div className={styles.onboarding_container}>
					<Panel className={classNames(base.pb3, base.ph4, base.ba, base.br2, colors.b__cool_pale_gray)}>
						<Heading className={classNames(base.pt4)} align="center" level="">
							Browse your private code with Sourcegraph
						</Heading>
						<div className={styles.user_actions} style={{maxWidth: "380px"}}>
							<p className={classNames(typography.tc, base.mt3, base.mb2, typography.f6, colors.cool_gray_8)} >
								Enable Sourcegraph on any private GitHub repositories for a better coding experience
							</p>
						</div>
						<div className={classNames(styles.user_actions, base.pt2)} style={{maxWidth: "380px"}}>
							<span className={styles.list_label_right}>ENABLE</span>
							<div className={styles.repos_list}>
								{(repos && repos.length > 0) && repos.map((repo, i) =>
									<div className={styles.row} key={i}>
										<div className={styles.info}>
											{repo.ID ?
												<RepoLink repo={repo.URI || `github.com/${repo.Owner}/${repo.Name}`} /> :
												(repo.URI && repo.URI.replace("github.com/", "").replace("/", " / ", 1)) || `${repo.Owner} / ${repo.Name}`
											}
										{repo.Description && <p className={styles.repo_description}>
											{repo.Description.length > 40 ? `${repo.Description.substring(0, 40)}...` : repo.Description}
										</p>}
										</div>
										<div className={styles.toggle}>
											<ToggleSwitch defaultChecked={Boolean(repo.ID)} onChange={(checked) => {
												this._toggleRepo(repo);
											}}/>
										</div>
									</div>
								)}
								{(!this.state.showAll && repos.length > 5) && <div className={classNames(styles.info, base.pt2)}>
									<a onClick={this._showAllReposClicked.bind(this)}>Show more</a>
								</div>}
							</div>
							<p>
								<Button onClick={this.props.completeStep.bind(this)} className={styles.action_link} type="button" color="blue">Save and continue</Button>
							</p>
						</div>
					</Panel>
				</div>
			</div>
		);
	}

	render(): JSX.Element | null {
		let conditionalRender = this.props.privateCodeAuthed ? this._renderIndexRepoCTA() : this._renderPrivateAuthCTA();
		return (<div>
			{conditionalRender}
		</div>);
	}
}
