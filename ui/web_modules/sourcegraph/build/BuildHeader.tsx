// tslint:disable

import * as React from "react";
import * as classNames from "classnames";

import Component from "sourcegraph/Component";
import TimeAgo from "sourcegraph/util/TimeAgo";
import {buildStatus, buildClass, elapsed} from "sourcegraph/build/Build";
import * as styles from "./styles/Build.css";

class BuildHeader extends Component<any, any> {
	reconcileState(state, props) {
		if (state.build !== props.build) {
			state.build = props.build;
		}
	}

	render(): JSX.Element | null {
		return (
			<header className={classNames(styles.header, buildClass(this.state.build))}>
				<div className={styles.number}>#{this.state.build.ID}</div>
				<div className={styles.status}>{buildStatus(this.state.build)}</div>
				<div className={styles.date}>
					<TimeAgo time={this.state.build.EndedAt || this.state.build.StartedAt || this.state.build.CreatedAt} />
				</div>
				<div className={styles.elapsed}>{elapsed(this.state.build)}</div>
			</header>
		);
	}
}

(BuildHeader as any).propTypes = {
	build: React.PropTypes.object.isRequired,
};

export default BuildHeader;
