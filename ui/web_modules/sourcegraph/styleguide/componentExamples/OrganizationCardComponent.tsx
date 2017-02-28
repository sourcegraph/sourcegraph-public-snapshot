import * as React from "react";
import { Code, Heading, OrganizationCard, Panel, Table } from "sourcegraph/components";
import { whitespace } from "sourcegraph/components/utils";

export function OrganizationCardComponent(): JSX.Element {
	return <div style={{ marginBottom: whitespace[4], marginTop: whitespace[4] }}>
		<Heading level={3}>Organization Card</Heading>

		<Panel hoverLevel="low">
			<div style={{ padding: whitespace[4] }}>
				<Heading level={7} color="gray" style={{ marginBottom: whitespace[3] }}>Default</Heading>
				<OrganizationCard icon="https://avatars2.githubusercontent.com/u/3979584?v=3&s=200" name="sourcegraph" userCount={12} style={{ marginBottom: whitespace[4] }} />

				<Heading level={7} color="gray" style={{ marginBottom: whitespace[3] }}>With hover state</Heading>
				<OrganizationCard icon="https://avatars2.githubusercontent.com/u/3979584?v=3&s=200" name="sourcegraph" userCount={12} hover={true} />
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
		<Heading level={6} style={{ marginTop: whitespace[3], marginBottom: whitespace[2] }}>
			OrganizationCard Properties
		</Heading>
		<Panel hoverLevel="low" style={{ padding: whitespace[4] }}>
			<Table style={{ width: "100%" }}>
				<thead>
					<tr>
						<td>Prop</td>
						<td>Default value</td>
						<td>Values</td>
					</tr>
				</thead>
				<tbody>
					<tr>
						<td><Code>icon</Code></td>
						<td><Code>undefined</Code></td>
						<td>
							The URL to organization's icon or logo. Optional.
						</td>
					</tr>
					<tr>
						<td><Code>name</Code></td>
						<td><Code>undefined</Code></td>
						<td>
							The name of the organization. Required.
						</td>
					</tr>
					<tr>
						<td><Code>userCount</Code></td>
						<td><Code>undefined</Code></td>
						<td>
							The number of users in the organization. Required.
						</td>
					</tr>
					<tr>
						<td><Code>hover</Code></td>
						<td><Code>undefined</Code></td>
						<td>
							Enabled a hover state. Only use if clicking this component will execute an action. Optional.
						</td>
					</tr>
				</tbody>
			</Table>
		</Panel>
	</div>;
}
