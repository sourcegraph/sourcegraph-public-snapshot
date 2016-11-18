import * as React from "react";
import {Link} from "react-router";
import {User} from "sourcegraph/api/index";
import {Avatar, FlexContainer, Heading, LanguageLabel, Panel} from "sourcegraph/components";
import {colors, whitespace} from "sourcegraph/components/utils";

interface Props {
	contributors?: User[];
	style?: React.CSSProperties;
	repo: GQL.IRemoteRepository;
}

export function RepositoryCard({style, repo, contributors}: Props): JSX.Element {

	const MAX_CONTRIBUTORS = 5;
	const hasMoreContribs = contributors && contributors.length > MAX_CONTRIBUTORS;

	let contributorAvatars;
	if (contributors && contributors.length > 0) {
		contributorAvatars = contributors.slice(0, MAX_CONTRIBUTORS).map( (user, i) => {
			return <Avatar size="tiny" img={user.AvatarURL} key={i} title={user.Login} style={{
				marginRight: whitespace[2],
				verticalAlign: "bottom",
			}} />;
		});
	}

	const sx = Object.assign(
		{ padding: whitespace[4] },
		style,
	);

	return <Panel hoverLevel="low" style={sx}>
		<FlexContainer justify="between">
			<div>
				<Heading level={6} style={{marginTop: 0}}>
					{repo.uri
						? <Link to={repo.uri}>{repo.owner} / {repo.name}</Link>
						: <span>{repo.owner} / {repo.name}</span>
					}
				</Heading>
				{repo.description && <div style={{color: colors.coolGray3()}}>{repo.description}</div>}
				{contributorAvatars && <div style={{marginTop: whitespace[3]}}>
					{contributorAvatars}
					{hasMoreContribs && contributors &&
						<span style={{color: colors.coolGray3()}}>
							+ {contributors.length - MAX_CONTRIBUTORS}
							{contributors.length - MAX_CONTRIBUTORS === 1 ? " contributor" : " contributors"}
						</span>
					}
				</div>}
			</div>
			{repo.language &&
				<LanguageLabel language={repo.language} style={{
					alignSelf: "flex-end",
					textAlign: "right",
				}} />
			}
		</FlexContainer>
	</Panel>;
};
