import { css } from "glamor";
import * as React from "react";
import { FlexContainer } from "sourcegraph/components";
import { Close } from "sourcegraph/components/symbols/Primaries";
import { colors, whitespace } from "sourcegraph/components/utils";

interface Props {
	color: "gray" | "green" | "red" | "yellow" | "white";
	style?: Object;
	className?: string;
	isDismissable: boolean;
	onDismiss?: () => void;
}

interface State {
	isVisible: boolean;
}

export class Toast extends React.Component<Props, State> {

	constructor(props: Props) {
		super(props);
		this.state = { isVisible: true };
	}

	dismiss(): void {
		this.setState({ isVisible: false });
		if (this.props.onDismiss) {
			this.props.onDismiss();
		}
	}

	render(): JSX.Element {
		const { color, children, className, isDismissable, style, } = this.props;
		const linkSx = css({
			"& a": { color: linkColors[color](), textDecoration: "underline" },
			"& a:hover": { color: linkColors[color]() },
		});

		const closeLinkSx = css({
			"&": { color: linkColors[color](0.5) },
			"&:hover": { color: linkColors[color](0.75) },
			"&:active": { color: linkColors[color]() },
		});

		return <FlexContainer className={className} style={Object.assign({
			backgroundColor: bgColor[color],
			boxShadow: `0 2px 4px 0 ${colors.black(0.5)}`,
			boxSizing: "border-box",
			color: textColor[color],
			overflow: "hidden",

			maxHeight: this.state.isVisible ? "200px" : "0",
			transition: "max-height 500ms ease-in-out",
		}, style)}>
			<div style={{ padding: whitespace[3], flex: "2 2 auto" }}>
				<span {...linkSx}>{children}</span>
			</div>
			{isDismissable && <a onClick={() => this.dismiss()} style={{ flex: "0 0 56px" }} {...closeLinkSx}>
				<Close width={24} style={{ margin: whitespace[3] }} />
			</a>}
		</FlexContainer >;
	}
}

const linkColors = {
	gray: colors.blueL2,
	green: colors.white,
	red: colors.white,
	yellow: colors.black,
	white: colors.blue,
};

const bgColor = {
	gray: colors.blueGrayD1(),
	green: colors.green(),
	red: colors.red(),
	yellow: colors.yellowL1(),
	white: colors.white(),
};

const textColor = {
	gray: "white",
	green: "white",
	red: "white",
	yellow: colors.yellowD2(),
	white: null,
};
