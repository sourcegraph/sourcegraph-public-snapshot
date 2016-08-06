// tslint:disable

import * as React from "react";
import CSSModules from "react-css-modules";
import * as styles from "./styles/stepper.css";
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
					<span className={`${styles.step} ${styles.step_complete} ${lineColorClasses[color] || styles.line_green}`} key={i}>
						<span className={`${styles.step_node_complete} ${nodeColorClasses[color] || styles.node_green}`}>
							<Icon icon="check" width="16px" className={styles.check} />
						</span>
						<span className={styles.step_text}>{step}</span>
					</span>
				);
			}
			return (
				<span className={`${styles.step} ${styles.step_incomplete}`} key={i}>
					<span className={styles.step_node_incomplete} />
					<span className={styles.step_text}>{step}</span>
				</span>
			);
		});
	}

	render(): JSX.Element | null {
		return <div className={`${this.props.className} ${styles.stepper}`}>{this.renderSteps()}</div>;
	}
}

const lineColorClasses = {
	"green": styles.line_green,
	"blue": styles.line_blue,
	"purple": styles.line_purple,
	"orange": styles.line_orange,
};

const nodeColorClasses = {
	"green": styles.node_green,
	"blue": styles.node_blue,
	"purple": styles.node_purple,
	"orange": styles.node_orange,
};

export default CSSModules(Stepper, styles, {allowMultiple: true});
