import { media } from "glamor";
import * as React from "react";
import { Link } from "react-router";
import { Avatar, FlexContainer, Heading, LanguageLabel, Panel } from "sourcegraph/components";
import { colors, layout, whitespace } from "sourcegraph/components/utils";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

interface Props {
	contributors?: GQL.IContributor[];
	style?: React.CSSProperties;
	repo: GQL.IRemoteRepository;
}

export function RepositoryCard({style, repo, contributors}: Props): JSX.Element {

	const MAX_CONTRIBUTORS = 5;
	const hasMoreContribs = contributors && contributors.length > MAX_CONTRIBUTORS;

	let contributorAvatars;
	if (contributors && contributors.length > 0) {
		contributorAvatars = contributors.slice(0, MAX_CONTRIBUTORS).map((user, i) => {
			return <Avatar size="tiny" img={user.avatarURL} key={i} title={user.login} style={{
				marginRight: whitespace[2],
				verticalAlign: "bottom",
			}} />;
		});
	}

	function trackRepoClick(): void {
		AnalyticsConstants.Events.Repository_Clicked.logEvent({ repo });
	}

	const sx = Object.assign(
		{ padding: whitespace[4] },
		style,
	);

	return <Panel hoverLevel="low" style={sx}>
		<FlexContainer justify="between" wrap={true}>
			<div style={{ flex: "1 1", marginRight: whitespace[2] }}>
				<Heading level={6} style={{ marginTop: 0 }}>
					{repo.uri
						? <Link to={repo.uri} onClick={trackRepoClick}>{repo.owner} / {repo.name}</Link>
						: <span>{repo.owner} / {repo.name}</span>
					}
				</Heading>
				{repo.description && <div style={{ color: colors.coolGray3() }}>{repo.description}</div>}
				{contributorAvatars && <div style={{ marginTop: whitespace[3] }}>
					{contributorAvatars}
					{hasMoreContribs && contributors &&
						<span style={{ color: colors.coolGray3() }}>
							+ {contributors.length - MAX_CONTRIBUTORS}
							{contributors.length - MAX_CONTRIBUTORS === 1 ? " contributor" : " contributors"}
						</span>
					}
				</div>}
			</div>
			{repo.language &&
				<span style={{ alignSelf: "flex-end" }}
					{ ...media(layout.breakpoints.sm, {
						flex: "1 0 100%",
						marginTop: whitespace[2],
					}) }>
					<LanguageLabel language={repo.language} />
				</span>
			}
		</FlexContainer>
	</Panel>;
};
