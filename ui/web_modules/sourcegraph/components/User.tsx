import * as React from "react";
import { Avatar, FlexContainer } from "sourcegraph/components";
import { colors, typography } from "sourcegraph/components/utils";
import { whitespace } from "sourcegraph/components/utils/whitespace";

interface Props {
	email?: string;
	nickname: string;
	avatar?: string;
	size?: "small" | "medium" | "large";
	simple?: boolean;
	style?: React.CSSProperties;
}

const sx = {
	color: colors.blueGray(),
	...typography.size[7],
};

export function User(props: Props): JSX.Element {
	const {
		avatar,
		email,
		nickname,
		simple = false,
		size = "large",
		style,
	} = props;

	return <div style={style}>
		<FlexContainer items="center">
			<div style={{ marginRight: simple ? whitespace[2] : whitespace[3], float: "left", lineHeight: 0 }}>
				<Avatar img={avatar} size={simple ? "tiny" : size} style={{ marginRight: whitespace[1] }} />
			</div>
			<div>
				<div>{nickname}</div>
				{!simple && <div style={sx}>{email}</div>}
			</div>
		</FlexContainer>
	</div>;
};
