// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as base from "sourcegraph/components/styles/_base.css";
import { Heading, Panel, Table, Code, List } from "sourcegraph/components";
import * as classNames from "classnames";

export class ListComponent extends React.Component<{}, any> {

	render(): JSX.Element | null {
		return (
			<div className={base.mv4}>
				<Heading level={3} className={base.mb2}>Lists</Heading>

				<Panel hoverLevel="low">
					<div className={base.pa4}>

						<List>
							<li>Item 1</li>
							<li>Item 2</li>
							<li>Item 3</li>
						</List>

					</div>
					<hr />
					<code>
						<pre className={base.ph4} style={{ whiteSpace: "pre-wrap" }}>
							{
								`
<List>
	<li>Item 1</li>
	<li>Item 2</li>
	<li>Item 3</li>
</List>

`
							}
						</pre>
					</code>
				</Panel>
				<Heading level={6} className={classNames(base.mt5, base.mb3)}>Properties</Heading>
				<Panel hoverLevel="low" className={base.pa4}>
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
								<td><Code>itemStyle</Code></td>
								<td><Code>undefined</Code></td>
								<td>
									<Code>undefined</Code>, <Code>style object</Code>
								</td>
							</tr>
						</tbody>
					</Table>
				</Panel>
			</div>
		);
	}
}
