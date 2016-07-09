// @flow

import React from "react";
import CSSModules from "react-css-modules";
import base from "sourcegraph/components/styles/_base.css";
import {Heading, Panel, Tabs, TabItem, TabPanels, TabPanel, Table, Code} from "sourcegraph/components";

class TableComponent extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
			activeExample: 0,
		};
	}

	state: {
		activeExample: number,
	};

	render() {
		return (
			<div className={base.mv4}>
				<Heading level="3" className={base.mb2}>Table</Heading>

				<Tabs color="purple">
					<TabItem
						active={this.state.activeExample === 0}>
						<a href="#" onClick={(e) => {
							this.setState({activeExample: 0});
							e.preventDefault();
						}}>
							Default
						</a>
					</TabItem>
					<TabItem
						active={this.state.activeExample === 1}>
						<a href="#" onClick={(e) => {
							this.setState({activeExample: 1});
							e.preventDefault();
						}}>
							Bordered
						</a>
					</TabItem>
				</Tabs>

				<Panel hoverLevel="low">
					<TabPanels active={this.state.activeExample}>
						<TabPanel>
							<div className={base.pa4}>
								<Heading level="7" className={base.mb3} color="cool-mid-gray">Default table style</Heading>
								<Table style={{width: "100%"}}>
									<thead>
										<tr>
											<td>Name</td>
											<td>User ID</td>
											<td>Email</td>
										</tr>
									</thead>
									<tbody>
										<tr>
											<td>Chelsea Otakan</td>
											<td>0</td>
											<td>chelsea@example.com</td>
										</tr>
										<tr>
											<td>Quinn Slack</td>
											<td>1</td>
											<td>sqs@example.com</td>
										</tr>
										<tr>
											<td>John Rothfels</td>
											<td>2</td>
											<td>john@example.com</td>
										</tr>
										<tr>
											<td>Renfred Harper</td>
											<td>3</td>
											<td>renfred@example.com</td>
										</tr>
									</tbody>
								</Table>
							</div>
							<hr />
							<code>
								<pre className={base.ph4} style={{whiteSpace: "pre-wrap"}}>
{
	`
<Table style={{width: "100%"}}>
	<thead>
		<tr>
			<td>Name</td>
			<td>User ID</td>
			<td>Email</td>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>Chelsea Otakan</td>
			<td>0</td>
			<td>chelsea@example.com</td>
		</tr>
		<tr>
			<td>Quinn Slack</td>
			<td>1</td>
			<td>sqs@example.com</td>
		</tr>
		<tr>
			<td>John Rothfels</td>
			<td>2</td>
			<td>john@example.com</td>
		</tr>
		<tr>
			<td>Renfred Harper</td>
			<td>3</td>
			<td>renfred@example.com</td>
		</tr>
	</tbody>
</Table>
	`
}
								</pre>
							</code>
						</TabPanel>
						<TabPanel>
							<div className={base.pa4}>
								<Heading level="7" className={base.mb3} color="cool-mid-gray">Default table style</Heading>
								<Table style={{width: "100%"}} tableStyle="bordered">
									<thead>
										<tr>
											<td>Name</td>
											<td>User ID</td>
											<td>Email</td>
										</tr>
									</thead>
									<tbody>
										<tr>
											<td>Chelsea Otakan</td>
											<td>0</td>
											<td>chelsea@example.com</td>
										</tr>
										<tr>
											<td>Quinn Slack</td>
											<td>1</td>
											<td>sqs@example.com</td>
										</tr>
										<tr>
											<td>John Rothfels</td>
											<td>2</td>
											<td>john@example.com</td>
										</tr>
										<tr>
											<td>Renfred Harper</td>
											<td>3</td>
											<td>renfred@example.com</td>
										</tr>
									</tbody>
								</Table>
							</div>
							<hr />
							<code>
								<pre className={base.ph4} style={{whiteSpace: "pre-wrap"}}>
{
	`
<Table style={{width: "100%"}} tableStyle="bordered">
	<thead>
		<tr>
			<td>Name</td>
			<td>User ID</td>
			<td>Email</td>
		</tr>
	</thead>
	<tbody>
		<tr>
			<td>Chelsea Otakan</td>
			<td>0</td>
			<td>chelsea@example.com</td>
		</tr>
		<tr>
			<td>Quinn Slack</td>
			<td>1</td>
			<td>sqs@example.com</td>
		</tr>
		<tr>
			<td>John Rothfels</td>
			<td>2</td>
			<td>john@example.com</td>
		</tr>
		<tr>
			<td>Renfred Harper</td>
			<td>3</td>
			<td>renfred@example.com</td>
		</tr>
	</tbody>
</Table>
	`
}
								</pre>
							</code>
						</TabPanel>
					</TabPanels>
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
								<td><Code>tableStyle</Code></td>
								<td><Code>default</Code></td>
								<td>
									<Code>default</Code>, <Code>bordered</Code>
								</td>
							</tr>
						</tbody>
					</Table>
				</Panel>
			</div>
		);
	}
}

export default CSSModules(TableComponent, base, {allowMultiple: true});
