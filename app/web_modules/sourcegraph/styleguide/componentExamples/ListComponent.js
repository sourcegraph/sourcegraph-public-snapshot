import * as React from "react";
import CSSModules from "react-css-modules";
import base from "sourcegraph/components/styles/_base.css";
import {Heading, Panel, Table, Code, List} from "sourcegraph/components";

class ListComponent extends React.Component {

	render() {
		return (
			<div className={base.mv4}>
				<Heading level="3" className={base.mb2}>Lists</Heading>

				<Panel hoverLevel="low">
					<div className={base.pa4}>

						<Heading level="7" className={base.mb3} color="cool_mid_gray">Normal</Heading>

						<List>
							<li>Item 1</li>
							<li>Item 2</li>
							<li>Item 3</li>
						</List>

						<Heading level="7" className={base.mb3} color="cool_mid_gray">Node style</Heading>

						<List listStyle="node">
							<li>Item 1</li>
							<li>Item 2</li>
							<li>Item 3</li>
						</List>


					</div>
					<hr />
					<code>
						<pre className={base.ph4} style={{whiteSpace: "pre-wrap"}}>
{
`
<List>
	<li>Item 1</li>
	<li>Item 2</li>
	<li>Item 3</li>
</List>

<List listStyle="node">
	<li>Item 1</li>
	<li>Item 2</li>
	<li>Item 3</li>
</List>

`
}
						</pre>
					</code>
				</Panel>
				<Heading level="4" className={`${base.mt5} ${base.mb3}`}>Properties</Heading>
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
								<td><Code>listStyle</Code></td>
								<td><Code>normal</Code></td>
								<td>
									<Code>normal</Code>, <Code>node</Code>
								</td>
							</tr>
						</tbody>
					</Table>
				</Panel>
			</div>
		);
	}
}

export default CSSModules(ListComponent, base, {allowMultiple: true});
