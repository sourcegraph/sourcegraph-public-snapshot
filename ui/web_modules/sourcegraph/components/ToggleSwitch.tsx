import { merge } from "glamor";
import * as React from "react";
import { colors } from "sourcegraph/components/utils";

interface Props {
	defaultChecked?: boolean;
	onChange: (checked: boolean) => void;
	labels?: boolean;
	size?: "small";
	style?: React.CSSProperties;
}

interface State {
	checked: boolean;
}

export class ToggleSwitch extends React.Component<Props, State> {
	state: State = {
		checked: false,
	};

	constructor(props: Props) {
		super(props);
		this.state = {
			checked: props.defaultChecked || false,
		};
	}

	_toggle(): void {
		this.setState({ checked: !this.state.checked }, () => this.props.onChange(this.state.checked));
	}

	render(): JSX.Element | null {
		const toggleWidth = this.props.size === "small" ? 57 : 70;
		const toggleHeight = this.props.size === "small" ? 22 : 36;
		const switchSize = this.props.size === "small" ? 16 : 30;

		const onOffLabelSx = {
			color: colors.coolGray2(),
			position: "absolute",
			lineHeight: `${toggleHeight + 2}px`,
			transition: "all 0.3s ease-in-out",
		};

		const sx = merge({
			display: "inline-block",
			verticalAlign: "middle",
			userSelect: "none",
		}, this.props.style ? this.props.style : {});

		return <div {...sx} onClick={this._toggle.bind(this)}>
			<input type="checkbox" name="toggle" style={{ display: "none" }} checked={this.state.checked} readOnly={true} />
			<label style={{
				background: "white",
				borderRadius: 20,
				cursor: "pointer",
				display: "block",
				overflow: "hidden",
				position: "relative",
				height: toggleHeight,
				width: toggleWidth,
			}}>
				<div style={{
					backgroundColor: this.state.checked ? colors.green() : colors.coolGray3(),
					left: this.state.checked ? toggleWidth - (switchSize + 6) : 0,
					borderRadius: "50%",
					boxSizing: "border-box",
					float: "left",
					height: switchSize,
					width: switchSize,
					margin: 3,
					padding: 0,
					position: "relative",
					transition: "all 0.3s ease-in-out",
					zIndex: 2,
				}}></div>
				{this.props.labels &&
					<strong style={Object.assign({
						left: 8,
						opacity: this.state.checked ? 1 : 0,
					}, onOffLabelSx)}>ON</strong>
				}
				{this.props.labels &&
					<strong style={Object.assign({
						right: 8,
						opacity: this.state.checked ? 0 : 1,
					}, onOffLabelSx)}>OFF</strong>
				}
			</label>
		</div>;
	}
}
