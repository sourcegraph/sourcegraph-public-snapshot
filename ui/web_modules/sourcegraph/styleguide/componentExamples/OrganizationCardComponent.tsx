import * as React from "react";
import { OrganizationCard, Panel } from "sourcegraph/components";
import { whitespace } from "sourcegraph/components/utils";

export function OrganizationCardComponent(): JSX.Element {
	return <div>
		<Panel hoverLevel="low">
			<div style={{ padding: whitespace[4] }}>
				<OrganizationCard icon="https://avatars2.githubusercontent.com/u/3979584?v=3&s=200" name="sourcegraph" desc="Fast, global, semantic code search & cross-reference engine for developers." style={{ marginBottom: whitespace[4] }} />
			</div>
			<hr />
			<code>
				<pre style={{
					whiteSpace: "pre-wrap",
					paddingLeft: whitespace[4],
					paddingRight: whitespace[4],
				}}>
					{`
<OrganizationCard icon="https://avatars2.githubusercontent.com/u/3979584?v=3&s=200" name="sourcegraph" userCount={12} />
<OrganizationCard icon="https://avatars2.githubusercontent.com/u/3979584?v=3&s=200" name="sourcegraph" userCount={12} hover={true} />

`
					}
				</pre>
			</code>
		</Panel>
	</div>;
}
