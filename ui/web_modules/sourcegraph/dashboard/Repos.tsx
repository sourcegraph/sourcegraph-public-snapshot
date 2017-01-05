import * as debounce from "lodash/debounce";
import * as React from "react";
import { context } from "sourcegraph/app/context";
import { FlexContainer, GitHubAuthButton, GoogleAuthButton, Heading, Input, Panel, RepositoryCard } from "sourcegraph/components";
import { Spinner } from "sourcegraph/components/symbols";
import { whitespace } from "sourcegraph/components/utils";
import { RepositoryTabs } from "sourcegraph/dashboard";
import { Location } from "sourcegraph/Location";
import { Features } from "sourcegraph/util/features";
import { privateGitHubOAuthScopes, privateGoogleOAuthScopes } from "sourcegraph/util/urlTo";

interface Props {
	repos: GQL.IRemoteRepository[] | null;
	location?: Location;
	style?: React.CSSProperties;
	type: RepositoryTabs;
}

export class Repos extends React.Component<Props, {}> {
	_filterInput: any;

	constructor(props: Props) {
		super(props);
		this._filterInput = null;
		this._handleFilter = this._handleFilter.bind(this);
		this._handleFilter = debounce(this._handleFilter, 25);
		this._showRepo = this._showRepo.bind(this);
	}

	// _repoSort is a comparison function that sorts more recently
	// pushed repos first.
	_repoSort(a: GQL.IRemoteRepository, b: GQL.IRemoteRepository): number {
		if (a.pushedAt < b.pushedAt) {
			return 1;
		}
		if (a.pushedAt > b.pushedAt) {
			return -1;
		}
		return 0;
	}

	_handleFilter(): void {
		this.forceUpdate();
	}

	_showRepo(repo: GQL.IRemoteRepository): boolean {
		if (this._filterInput && this._filterInput.value &&
			this._qualifiedName(repo).indexOf(this._filterInput.value.trim().toLowerCase()) === -1) {
			return false;
		}

		return true; // no filter; return all
	}

	_qualifiedName(repo: GQL.IRemoteRepository): string {
		return (`${repo.owner}/${repo.name} ${repo.language}`).toLowerCase();
	}

	_header(): JSX.Element {
		const btnSx = {
			marginRight: whitespace[2],
			marginBottom: whitespace[3],
		};
		return <header>
			{!context.hasPrivateGitHubToken() && <p>Private code indexed on Sourcegraph is only available to you and those with permissions to the underlying GitHub repository.</p>}
			<FlexContainer items="center" justify="center" wrap={true}>
				{!context.hasPrivateGitHubToken() &&
					<GitHubAuthButton scopes={privateGitHubOAuthScopes} style={btnSx} returnTo={this.props.location}>
						Add private repositories
					</GitHubAuthButton>
				}
				{Features.googleCloudPlatform.isEnabled() && !context.hasPrivateGoogleToken() &&
					<GoogleAuthButton scopes={privateGoogleOAuthScopes} returnTo={this.props.location} style={btnSx} >
						Add GCP repositories
					</GoogleAuthButton>
				}
			</FlexContainer>
		</header>;
	}

	_getTitle(title: string): string {
		if (title === "mine") { return "My repositories"; };
		if (title === "starred") { return "Starred repositories"; };
		return "Repositories";
	}

	_repoList(): JSX.Element {
		if (Object.keys(this.props.repos).length === 0) {
			return <div style={{ margin: whitespace[4] }}>
				<Heading level={4} style={{ marginTop: 0 }}>{this._getTitle(this.props.type)}</Heading>
				<Panel hoverLevel="low" style={{
					marginTop: whitespace[3],
					padding: whitespace[4],
					textAlign: "center",
				}}>
					It doesn't look like you have any repositories.
					{this.props.type === "mine" &&
						<span> Try <a href="https://help.github.com/articles/create-a-repo/">creating some repositories on GitHub</a>.</span>
					}
					{this.props.type === "starred" &&
						<span> Try <a href="https://help.github.com/articles/about-stars/">starring some repositories on GitHub</a>.</span>
					}
				</Panel>
			</div>;
		}

		let filteredRepos: GQL.IRemoteRepository[] = [];
		if (this.props.repos) {
			filteredRepos = this.props.repos.filter(this._showRepo).sort(this._repoSort);
		}

		return <div style={{ margin: whitespace[4] }}>
			<FlexContainer justify="between" wrap={true} style={{
				marginTop: whitespace[3],
				marginBottom: whitespace[3],
			}}>
				<Heading level={4} style={{ marginTop: 0 }}>{this._getTitle(this.props.type)}</Heading>
				<Input type="text"
					placeholder="Filter repositories ..."
					domRef={(e) => this._filterInput = e}
					spellCheck={false}
					onChange={this._handleFilter}
					inputSize="small"
					style={{ marginTop: whitespace[1], minWidth: 225 }} />
			</FlexContainer>
			<div>
				{filteredRepos.length > 0 && filteredRepos.map((repo, i) => {
					return <RepositoryCard repo={repo} key={i} style={{ marginBottom: whitespace[3] }} />;
				})}
			</div>
			{context.gitHubToken && this._filterInput && this._filterInput.value && filteredRepos.length === 0 &&
				<div style={{ margin: whitespace[4], textAlign: "center" }}>No matching repositories. Try a different search.</div>
			}
		</div>;
	}

	render(): JSX.Element {
		return <div style={this.props.style}>
			{this._header()}
			{!this.props.repos ? <div style={{ margin: whitespace[4], marginTop: 0, textAlign: "center" }}><Spinner /> Loading...</div> : this._repoList()}
		</div>;
	}
}
