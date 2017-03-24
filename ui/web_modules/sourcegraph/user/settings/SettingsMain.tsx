import * as React from "react";
import { Link } from "react-router";

import { abs, isAtRoute } from "sourcegraph/app/routePatterns";
import { PageTitle, Panel, TabItem, Tabs } from "sourcegraph/components";
import { colors, layout, whitespace } from "sourcegraph/components/utils";
import { ComponentWithRouter } from "sourcegraph/core/ComponentWithRouter";
import { OrgContainer } from "sourcegraph/org/OrgContainer";
import { BillingContainer } from "sourcegraph/user/Billing";

export class SettingsMain extends ComponentWithRouter<{}, {}> {
	render(): JSX.Element {
		const billingPage = isAtRoute(this.context.router, abs.settings);
		return <div style={{
			background: colors.blueGrayL3(),
			flex: "1 1 auto",
			paddingTop: whitespace[7],
			paddingBottom: whitespace[6],
		}}>
			<PageTitle title="Account and billing settings" />
			<div style={{
				...layout.container.sm,
				display: "flex",
				minHeight: "min-content",
			}}>
				<Tabs style={{ height: "100%", minWidth: 200 }} direction="vertical">
					<TabItem direction="vertical" active={billingPage}>
						<Link to={`/${abs.settings}`}>Billing information</Link>
					</TabItem>
					<TabItem direction="vertical" active={!billingPage}>
						<Link to={`/${abs.orgSettings}`}>Organizations</Link>
					</TabItem>
				</Tabs>
				<Panel style={{ width: "100%" }} hoverLevel="low" hover={false}>
					{billingPage ? <BillingContainer /> : <OrgContainer />}
				</Panel>
			</div>
		</div>;
	}
}
