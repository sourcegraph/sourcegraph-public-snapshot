import * as React from "react";
import { Avatar, FlexContainer } from "sourcegraph/components";
import { colors, typography } from "sourcegraph/components/utils";
import { whitespace } from "sourcegraph/components/utils/whitespace";

interface Props {
	email?: string;
	nickname: string;
	avatar?: string;
	style?: React.CSSProperties;
	simple?: boolean;
}

const sx = Object.assign(
	{ color: colors.coolGray3() },
	typography.size[7],
);

export function User(props: Props): JSX.Element {
	const {
		avatar,
		email,
		nickname,
		simple = false,
		style,
	} = props;

	return <div style={style}>
		<FlexContainer items="center">
			<div style={{ marginRight: simple ? whitespace[2] : whitespace[3], float: "left", lineHeight: 0 }}>
				<Avatar img={avatar} size={simple ? "tiny" : "medium"} style={{ marginRight: 2 }} />
			</div>
			<div>
				<div style={{ lineHeight: "19px" }}>{nickname}</div>
				{!simple && <div style={sx}>{email}</div>}
			</div>
		</FlexContainer>
	</div>;
};
