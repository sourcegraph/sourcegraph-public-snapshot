import { css } from "glamor";
import * as React from "react";
import { Avatar, FlexContainer } from "sourcegraph/components";
import { Check, ChevronDown } from "sourcegraph/components/symbols/Zondicons";
import { colors, layout, typography, whitespace } from "sourcegraph/components/utils";
import { TimeFromNowUntil } from "sourcegraph/util/dateFormatterUtil";

export function Commit({ commit, hover = true, showChevron = false, selected = false, style }: {
	commit: GQL.ICommitInfo;
	showChevron?: boolean;
	selected?: boolean;
	hover?: boolean;
	style?: React.CSSProperties;
}): JSX.Element {
	const committer = commit.committer;
	if (!committer || !committer.person) {
		return <span />;
	}
	const date = TimeFromNowUntil(committer.date, 14);
	return <div style={style} {...css(
		{
			backgroundColor: colors.black(0.2),
			zIndex: 2,
			height: layout.editorCommitInfoBarHeight,
			position: "relative",
		},
		hover ? {
			":hover": {
				color: "white",
				backgroundColor: colors.blue(),
			},
			":active": { backgroundColor: colors.blueD1() },
			":hover span": { color: "white !important" }
		} : {}) }>
		<FlexContainer justify="between" style={Object.assign({
			color: "white",
			paddingLeft: "0.6rem",
			paddingTop: "0.6rem",
		}, typography.size[7])}>
			<div style={{
				whiteSpace: "nowrap",
				paddingRight: "5%",
				textOverflow: "ellipsis",
				overflow: "hidden",
			}}>
				<Avatar
					img={`https://secure.gravatar.com/avatar/${committer.person.gravatarHash}?s=128&d=retro`}
					size="tiny"
					style={{ marginRight: whitespace[2], flex: "0 0 auto", verticalAlign: "top" }} />
				<span style={{ color: colors.blueGrayL1() }}>{committer.person.name} &ndash;</span>
				<div style={{ display: "inline", paddingLeft: 4 }}>
					{commit.message.substr(0, commit.message.length > 150 ? 150 : commit.message.length)}
				</div>
			</div>
			<FlexContainer items="start" justify="between" style={{
				flex: "0 0 180px",
				paddingRight: whitespace[1],
				textAlign: "left",
			}}>
				<span style={{ color: colors.blueGrayL1() }}>{date} @ {commit.rev.substr(0, 6)}</span>
				{showChevron &&
					<span style={{ color: colors.blueGrayL1() }}>
						<ChevronDown width={12} color="currentColor" style={{ padding: whitespace[2] }} />
					</span>
				}
				{selected &&
					<span style={{ color: colors.blueGrayL1() }}>
						<Check width={12} color="currentColor" style={{ padding: whitespace[2] }} />
					</span>
				}
			</FlexContainer>
		</FlexContainer>
	</div>;
}
