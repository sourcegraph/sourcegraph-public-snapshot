import * as React from "react";

import { context } from "sourcegraph/app/context";
import { FlexContainer, Toast } from "sourcegraph/components/";
import { No, Warning } from "sourcegraph/components/symbols/Primaries";
import { colors, whitespace } from "sourcegraph/components/utils";

const sales = <a href="mailto:hi@sourcegraph.com">hi@sourcegraph.com</a>;

export function Paywall(err: Error): JSX.Element {
	if (!err.message.match(/(account blocked)|(trial expired)/)) {
		return <div />;
	}
	return <FlexContainer direction="top-bottom" style={{ flex: "1 1 auto" }}>
		<Toast isDismissable={false} color="gray" style={{ flex: "0 0 auto" }}>
			<No color={colors.orangeL1()} width={24} style={{ marginRight: whitespace[3] }} />
			You don't have permission to view private repositories. Please contact {sales} to upgrade your account.
		</Toast>
		<div style={{
			backgroundColor: colors.blueGrayD2(),
			backgroundImage: `url(${context.assetsRoot}/img/blur-screenshot.png)`,
			backgroundSize: "cover",
			backgroundRepeat: "no-repeat",
			flex: "1 1 auto",
		}}></div>
	</FlexContainer>;
}

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
	return <Toast color="gray" isDismissable={true} style={{ zIndex: 6 }}>
		<Warning color={colors.yellow()} width={24} style={{ marginRight: whitespace[3] }} />
		Your free trial is ending {timeLeft}. Please contact {sales} to continue using
		Sourcegraph on private code.
	</Toast>;
}
