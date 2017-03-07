import * as React from "react";
import { List, Panel } from "sourcegraph/components";
import * as base from "sourcegraph/components/styles/_base.css";

export class ListComponent extends React.Component<{}, any> {

	render(): JSX.Element | null {
		return (
			<div>
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
			</div>
		);
	}
}
