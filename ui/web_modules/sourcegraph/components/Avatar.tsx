import * as React from "react";

const PLACEHOLDER_IMAGE = "https://secure.gravatar.com/avatar?d=mm&f=y&s=128";

interface Props {
	img?: string;
	size: "tiny" | "small" | "medium" | "large";
	style?: React.CSSProperties;
	title?: string;
}

export function Avatar({img, size, style, title}: Props): JSX.Element {

	let avatarSize;

	switch (size) {
		case "tiny":
			avatarSize = "19px";
			break;
		case "small":
			avatarSize = "2rem";
			break;
		case "medium":
			avatarSize = "3rem";
			break;
		case "large":
			avatarSize = "4rem";
			break;
	}

	const sx = Object.assign(
		{
			borderRadius: "50%",
			display: "inline-block",
			width: avatarSize,
			height: avatarSize,
		},
		style,
	);

	return <img style={sx} src={img || PLACEHOLDER_IMAGE} title={title} />;

};
