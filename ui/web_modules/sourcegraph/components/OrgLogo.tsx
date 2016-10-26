import * as React from "react";
import {colors} from "sourcegraph/components/utils";

interface Props {
	style?: React.CSSProperties;
	img: string;
	size: "tiny" | "small" | "medium" | "large";
}
export function OrgLogo({size, img, style}: Props): JSX.Element {
	let imgSize: string;
	switch (size) {
		case "tiny":
			imgSize = "1.58rem";
			break;
		case "small":
			imgSize = "2rem";
			break;
		case "medium":
			imgSize = "3rem";
			break;
		case "large":
			imgSize = "4rem";
			break;
		default:
			throw new Error("invalid size");
	}

	return <div style={Object.assign({}, style, sx)}>
		<img style={imgSx(imgSize)} src={img} />
	</div>;
};

function imgSx(size: string): Object {
	return {
		borderRadius: "3px",
		display: "inline-block",
		width: size,
		height: size,
	};
}

const sx = {
	display: "inline-block",
	backgroundColor: "white",
	borderColor: colors.coolGray4(0.8),
	borderRadius: "3px",
	borderStyle: "solid",
	borderWidth: 1,
	padding: 3,
	lineHeight: "0",
};
