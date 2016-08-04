// tslint:disable

import * as React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/stepper.css";
import Icon from "./Icon";

// @TODO(chexee): Doesn't scale well with large step labels. Keep 'em short for now.
// @TODO(chexee): Checks are a little off center in Firefox
// @TODO(chexee): animation between states

class Stepper extends React.Component<any, any> {
	static propTypes = {
		className: React.PropTypes.string,
		children: React.PropTypes.any,
		steps: React.PropTypes.array, // Array of step labels. Pass in null for no labels.
		stepsComplete: React.PropTypes.number, // Number of steps complete
		color: React.PropTypes.string, // "purple", "blue", "green", "orange"
	};

	static defaultProps = {
		steps: [null, null, null],
		stepsComplete: 0,
		color: "green",
	};

	renderSteps() {
		const {steps, stepsComplete, color} = this.props;
		return steps.map((step, i) => {
			if (i < stepsComplete) {
				return (
					<span styleName={`step step_complete line_${color}`} key={i}>
						<span styleName={`step_node_complete node_${color}`}>
							<Icon icon="check" width="16px" styleName="check" />
						</span>
						<span styleName="step_text">{step}</span>
					</span>
				);
			}
			return (
				<span styleName="step step_incomplete" key={i}>
					<span styleName="step_node_incomplete" />
					<span styleName="step_text">{step}</span>
				</span>
			);
		});
	}

	render(): JSX.Element | null {
		return <div className={this.props.className} styleName="stepper">{this.renderSteps()}</div>;
	}
}

export default CSSModules(Stepper, styles, {allowMultiple: true});
