import * as autobind from "autobind-decorator";
import * as React from "react";
import * as Relay from "react-relay";
import { Link } from "react-router";
import { getRoutePattern } from "sourcegraph/app/routePatterns";
import { RouterContext } from "sourcegraph/app/router";
import { Commit } from "sourcegraph/blob/CommitInfoBar/Commit";
import { Popover } from "sourcegraph/components";
import { colors, layout } from "sourcegraph/components/utils";
import { urlWithRev } from "sourcegraph/repo/routes";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

interface Props {
	repo: string;
	path: string;
	rev: string | null;
	relay: any;
	root: GQL.IRoot;
}

@autobind
class CommitInfoBarComponent extends React.Component<Props, {}> {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	context: RouterContext;

	private commitInfoForRev(rev: string, commits: Array<GQL.ICommitInfo>): GQL.ICommitInfo {
		for (const commit of commits) {
			if (commit.rev === rev) {
				return commit;
			}
		}

		return commits[0];
	}

	private revSwitcherURL(rev: string | null): string {
		return `${urlWithRev(getRoutePattern(this.context.router.routes), this.context.router.params, rev)}${this.context.router.location.hash}`;
	}

	render(): JSX.Element | null {
		const repository = this.props.root.repository;
		if (!(repository && repository.commit && repository.commit.commit && repository.commit.commit.file && repository.commit.commit.file.commits && repository.commit.commit.file.commits[0])) {
			return null;
		}
		const commits = repository.commit.commit.file.commits;
		const commitSelected = (commits[0].rev === this.props.rev || !this.props.rev) ? commits[0] : this.commitInfoForRev(this.props.rev, commits);
		const eventProps = { repo: this.props.repo, path: this.props.path, rev: commitSelected.rev };

		// Business Logic: Designs don't want the latest commit to show in the drop down
		// if commitSelected === commits[0] (the latest commit)
		let commitOffset: Array<GQL.ICommitInfo> = this.props.rev !== commits[0].rev ? commits :
			repository.commit.commit.file.commits.slice(repository.commit.commit.file.commits.length > 1 ? 1 : 0);

		const dropdownClickHandler = visible => {
			if (visible) {
				AnalyticsConstants.Events.CommitInfo_Initiated.logEvent(eventProps);
			} else {
				AnalyticsConstants.Events.CommitInfo_Dismissed.logEvent(eventProps);
			}
		};

		const dropdownSx = {
			zIndex: 1,
			display: "block",
			left: 0,
			background: colors.blueGrayD2(),
			boxShadow: `${colors.black(0.3)} 0 1px 6px 0px`,
			borderRadius: 0,
			width: "100%",
		};

		const dropdownHeight = commitOffset.length > 5 ? layout.editorCommitInfoBarHeight * 5 : layout.editorCommitInfoBarHeight * commitOffset.length;

		const commitList = commitOffset.map(commit => {
			function commitClickHandler(): void {
				AnalyticsConstants.Events.CommitInfoItem_Selected.logEvent(Object.assign(
					{ selectedRev: commit.rev },
					eventProps
				));
			}
			return <Link
				key={`${commit.rev}${commit.message}`}
				style={{ width: "100%" }}
				role="menu_item"
				to={this.revSwitcherURL(commit.rev)}
				onClick={commitClickHandler}>
				<Commit commit={commit} selected={commitSelected.rev === commit.rev} />
			</Link>;
		});

		return <Popover
			left={true}
			pointer={false}
			popoverStyle={dropdownSx}
			onClick={dropdownClickHandler}
			style={{ zIndex: 1 }}>
			<Commit
				commit={commitSelected}
				showChevron={true}
				hover={false}
				style={{ boxShadow: `${colors.black(0.3)} 0 1px 6px 0px` }} />
			<div style={{ height: dropdownHeight, overflow: "auto" }}>{commitList}</div>
		</Popover>;
	}
};

const CommitInfoBarContainer = Relay.createContainer(CommitInfoBarComponent, {
	initialVariables: {
		repo: "",
		rev: "",
		path: "",
	},
	fragments: {
		root: () => Relay.QL`
			fragment on Root {
				repository(uri: $repo) {
					commit(rev: $rev) {
						commit {
							file(path: $path) {
								commits {
									rev
									message
									committer {
										person {
											name
											email
											gravatarHash
										}
										date
									}
								}
							}
						}
					}
				}
			}
		`,
	},
});

export const CommitInfoBar = function (props: { repo: string, rev: string, path: string }): JSX.Element {
	return <Relay.RootContainer
		Component={CommitInfoBarContainer}
		route={{
			name: "Root",
			queries: {
				root: () => Relay.QL`
					query { root }
				`,
			},
			params: props,
		}}
	/>;
};
