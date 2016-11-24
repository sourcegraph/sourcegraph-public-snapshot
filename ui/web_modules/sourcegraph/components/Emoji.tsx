import * as React from "react";
import { context } from "sourcegraph/app/context";

interface Props {
	className?: string;
	width?: string; // appended by "px"
	name: string; // See symbols directory
}

type State = any;

export class Emoji extends React.Component<Props, State> {

	static defaultProps: {
		width: "16px",
	};

	render(): JSX.Element | null {
		return <img src={`${context.assetsRoot}/img/emoji/${this.props.name}.svg`} width={this.props.width} className={this.props.className} />;
	}
}
