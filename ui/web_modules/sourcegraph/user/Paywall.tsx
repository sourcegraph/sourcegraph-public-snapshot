import * as React from "react";

import * as colors from "sourcegraph/components/utils/colors";

const sales = <a href="mailto:sales@sourcegraph.com">sales@sourcegraph.com</a>;

const headerSx = {
	height: 100,
	background: colors.blueD2(),
	color: colors.white(),
};

export function Paywall(err: Error): JSX.Element {
	if (!err.message.match(/account blocked/)) {
		return <div></div>;
	}
	return <div>
		<div style={headerSx}>
			Your trial has ended. Please contact {sales} to continue using
			Sourcegraph on private code.
		</div>
		<div>
			A screenshot
		</div>
	</div>;
}

const trialEndingSx = {
	background: colors.yellow(),
	height: 50,
};

export function TrialEndingWarning({ layout }: { layout: () => void }): JSX.Element {
	return <div style={trialEndingSx}>
		Your free trial is ending. Please contact {sales} to continue using
		Sourcegraph on private code.
	</div>;
}
