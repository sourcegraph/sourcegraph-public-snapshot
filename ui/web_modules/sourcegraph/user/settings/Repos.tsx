import * as debounce from "lodash/debounce";
import * as React from "react";
import { context } from "sourcegraph/app/context";
import { FlexContainer, Heading, Input, RepositoryCard } from "sourcegraph/components";
import { GitHubAuthButton, GoogleAuthButton } from "sourcegraph/components";
import { Spinner } from "sourcegraph/components/symbols";
import { layout, whitespace } from "sourcegraph/components/utils";
import { Location } from "sourcegraph/Location";
import { Features } from "sourcegraph/util/features";
import { privateGitHubOAuthScopes, privateGoogleOAuthScopes } from "sourcegraph/util/urlTo";

interface Props {
	repos: GQL.IRemoteRepository[] | null;
	location?: Location;
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

	_repoList(): JSX.Element {
		let filteredRepos: GQL.IRemoteRepository[] = [];
		if (this.props.repos) {
			filteredRepos = this.props.repos.filter(this._showRepo).sort(this._repoSort);
		}

		return <div style={{ margin: whitespace[4] }}>
			<FlexContainer justify="between" wrap={true} style={{
				marginTop: whitespace[4],
				marginBottom: whitespace[3],
			}}>
				<Heading level={4}>My repositories</Heading>
				<Input type="text"
					placeholder="Filter repositories ..."
					domRef={(e) => this._filterInput = e}
					spellCheck={false}
					onChange={this._handleFilter}
					inputSize="small"
					style={{ marginTop: whitespace[3], minWidth: 225 }} />
			</FlexContainer>
			<div>
				{filteredRepos.length > 0 && filteredRepos.map((repo, i) => {
					return <RepositoryCard contributors={repo.contributors} repo={repo} key={i} style={{ marginBottom: whitespace[3] }} />;
				})}
			</div>
			{context.gitHubToken && this._filterInput && this._filterInput.value && filteredRepos.length === 0 &&
				<div style={{ margin: whitespace[4], textAlign: "center" }}>No matching repositories</div>
			}
		</div>;
	}

	render(): JSX.Element {
		const sx = Object.assign({},
			layout.container,
			{
				maxWidth: 960,
				padding: whitespace[3],
			},
		);

		return <div style={sx}>
			{this._header()}
			{!this.props.repos ? <div style={{ margin: whitespace[4], textAlign: "center" }}><Spinner /> Loading...</div> : this._repoList()}
		</div>;
	}
}
