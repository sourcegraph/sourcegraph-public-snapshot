import * as classNames from "classnames";
import * as React from "react";

import {Icon} from "sourcegraph/components/Icon";
import * as styles from "sourcegraph/components/styles/stepper.css";

// @TODO(chexee): Doesn't scale well with large step labels. Keep 'em short for now.
// @TODO(chexee): Checks are a little off center in Firefox
// @TODO(chexee): animation between states

interface Props {
	className?: string;
	children?: any;
	steps: any[]; // Array of step labels. Pass in null for no labels.
	stepsComplete?: number; // Number of steps complete
	color?: string; // "purple", "blue", "green", "orange"
}

export class Stepper extends React.Component<Props, {}> {
	static defaultProps: Props = {
		steps: [null, null, null],
		stepsComplete: 0,
		color: "green",
	};

	getStepsFragment(): JSX.Element[] {
		const {steps, stepsComplete, color} = this.props;
		return steps.map((step, i) => {
			if (i < stepsComplete) {
				return (
					<span className={classNames(styles.step, styles.step_complete, lineColorClasses[color || "green"] || styles.line_green)} key={i}>
						<span className={classNames(styles.step_node_complete, nodeColorClasses[color || "green"] || styles.node_green)}>
							<Icon icon="check" width="16px" className={styles.check} />
						</span>
						<span className={styles.step_text}>{step}</span>
					</span>
				);
			}
			return (
				<span className={classNames(styles.step, styles.step_incomplete)} key={i}>
					<span className={styles.step_node_incomplete} />
					<span className={styles.step_text}>{step}</span>
				</span>
			);
		});
	}

	render(): JSX.Element | null {
		return <div className={classNames(this.props.className, styles.stepper)}>{this.getStepsFragment()}</div>;
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
