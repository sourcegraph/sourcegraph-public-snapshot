// tslint:disable

import * as React from "react";

import Component from "sourcegraph/Component";

export function formatDuration(ms) {
	if (ms === 0) return "0s";

	let s = Math.floor(ms / 1000) % 60;
	let m = Math.floor(ms / 1000 / 60) % 60;
	let h = Math.floor(ms / 1000 / 60 / 60) % 24;
	let d = Math.floor(ms / 1000 / 60 / 60 / 24) % 30;
	let mth = Math.floor(ms / 1000 / 60 / 60 / 24 / 30) % 12;
	let yr = Math.floor(ms / 1000 / 60 / 60 / 24 / 30 / 12);

	let parts: string[] = [];

	if (yr > 0) parts.push(`${yr}yr`);
	// Only show smaller time intervals if they are significant.
	if (yr < 2 && mth > 0) parts.push(`${mth}mth`);
	if (yr === 0 && mth < 3 && d > 0) parts.push(`${d}d`);
	if (mth === 0 && d < 3 && h > 0) parts.push(`${h}h`);
	if (d === 0 && h < 12 && m > 0) parts.push(`${m}m`);
	if (h === 0 && m < 5 && s > 0) parts.push(`${s}s`);

	return parts.join(" ");
}


class TimeAgo extends Component<any, any> {
	render(): JSX.Element | null {
		return <time title={this.props.time} {...this.props}>{formatDuration(Date.now() - new Date(this.props.time).getTime())} ago</time>;
	}
}
(TimeAgo as any).propTypes = {
	time: React.PropTypes.string.isRequired,
};

export default TimeAgo;
