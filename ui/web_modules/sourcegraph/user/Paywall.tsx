import * as React from "react";

import { No } from "sourcegraph/components/symbols/Primaries";
import * as colors from "sourcegraph/components/utils/colors";

const sales = <a href="mailto:sales@sourcegraph.com">sales@sourcegraph.com</a>;

const headerSx = {
	height: 100,
	background: colors.blueD2(),
	color: colors.white(),
};

export function Paywall(err: Error): JSX.Element {
	if (!err.message.match(/(account blocked)|(trial expired)/)) {
		return <div />;
	}
	return <div>
		<div style={headerSx}>
			<No />Your trial has ended. Please contact {sales} to continue using
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

export function TrialEndingWarning({ layout, repo }: {
	layout: () => void, repo: GQL.IRepository
}): JSX.Element {
	if (!repo.expirationDate) {
		return <div />;
	}
	const time = new Date(repo.expirationDate * 1000);
	const msUntilExpiration = time.getTime() - Date.now();
	const fiveDays = 1000 * 60 * 60 * 24 * 5;
	if (msUntilExpiration > fiveDays || msUntilExpiration < 0) {
		return <div />;
	}
	const daysLeft = new Date(msUntilExpiration).getUTCDate() - 1;
	let timeLeft: string;
	switch (daysLeft) {
		case 0: timeLeft = "today"; break;
		case 1: timeLeft = "tomorrow"; break;
		case 2: timeLeft = "in two days"; break;
		case 3: timeLeft = "in three days"; break;
		case 4: timeLeft = "in four days"; break;
		case 5: timeLeft = "in five days"; break;
		default: timeLeft = "soon";
	}
	return <div style={trialEndingSx}>
		Your free trial is ending {timeLeft}. Please contact {sales} to continue using
		Sourcegraph on private code.
	</div>;
}
