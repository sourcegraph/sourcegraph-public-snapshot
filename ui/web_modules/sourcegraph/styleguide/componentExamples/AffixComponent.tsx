// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as base from "sourcegraph/components/styles/_base.css";
import {Heading, Panel, Table, Code} from "sourcegraph/components/index";
import * as classNames from "classnames";

export class AffixComponent extends React.Component<{}, any> {

	render(): JSX.Element | null {
		return (
			<div className={base.mv4}>
				<Heading level="3" className={base.mb3}>Affix</Heading>

				<Panel hoverLevel="low">
					<div className={base.pa4}>
						<p>The <Code>Affix</Code> component makes any content inside of it <Code>position:fixed</Code> once you've scrolled past it's original position. If you scroll back, the content returns to it's original position.</p>
						<p>See the menu to the right for a working example.</p>
					</div>
					<hr />
					<code>
						<pre className={base.ph4} style={{whiteSpace: "pre-wrap"}}>
{
`
<Affix offset={20}>
	...
</Affix>

`
}
						</pre>
					</code>
				</Panel>
				<Heading level="4" className={classNames(base.mt5, base.mb3)}>Properties</Heading>
				<Panel hoverLevel="low" className={base.pa4}>
					<Table style={{width: "100%"}}>
						<thead>
							<tr>
								<td>Prop</td>
								<td>Default value</td>
								<td>Values</td>
							</tr>
						</thead>
						<tbody>
							<tr>
								<td><Code>offset</Code></td>
								<td><Code>null</Code></td>
								<td>
									Any positive <Code>integer</Code>. This determines how far from the top of the window the content will be posititioned.
								</td>
							</tr>
						</tbody>
					</Table>
				</Panel>
			</div>
		);
	}
}
