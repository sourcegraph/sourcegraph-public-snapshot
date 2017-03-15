import * as React from "react";
import { Link } from "react-router";

import { abs, isAtRoute } from "sourcegraph/app/routePatterns";
import { Panel } from "sourcegraph/components";
import { TabItem } from "sourcegraph/components/TabItem";
import { Tabs } from "sourcegraph/components/Tabs";
import { layout, whitespace } from "sourcegraph/components/utils";
import { ComponentWithRouter } from "sourcegraph/core/ComponentWithRouter";
import { OrgContainer } from "sourcegraph/org/OrgContainer";
import { BillingContainer } from "sourcegraph/user/Billing";

export class SettingsMain extends ComponentWithRouter<{}, {}> {
	render(): JSX.Element {
		const billingPage = isAtRoute(this.context.router, abs.settings);
		return <div style={{
			...layout.container,
			width: "90%",
			display: "flex",
			marginBottom: whitespace[4],
			marginTop: whitespace[4],
			minHeight: "min-content",
		}}>
			<Tabs style={{ height: "100%", minWidth: 200 }} direction="vertical">
				<TabItem direction="vertical" active={billingPage}>
					<Link to={`/${abs.settings}`}>Billing Information</Link>
				</TabItem>
				<TabItem direction="vertical" active={!billingPage}>
					<Link to={`/${abs.orgSettings}`}>Organizations</Link>
				</TabItem>
			</Tabs>
			<Panel style={{ width: "100%" }} hoverLevel="low" hover={false}>
				{billingPage ? <BillingContainer /> : <OrgContainer />}
			</Panel>
		</div>;
	}
}
