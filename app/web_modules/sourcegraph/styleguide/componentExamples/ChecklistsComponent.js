import * as React from "react";
import CSSModules from "react-css-modules";
import base from "sourcegraph/components/styles/_base.css";
import {Heading, Panel, Tabs, TabItem, TabPanels, TabPanel, Table, Code, ChecklistItem} from "sourcegraph/components";

class ChecklistsComponent extends React.Component {
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
				<Heading level="3" className={base.mb2}>Checklist Items</Heading>

				<Tabs color="purple">
					<TabItem
						active={this.state.activeExample === 0}>
						<a href="#" onClick={(e) => {
							this.setState({activeExample: 0});
							e.preventDefault();
						}}>
							Without CTA
						</a>
					</TabItem>
					<TabItem
						active={this.state.activeExample === 1}>
						<a href="#" onClick={(e) => {
							this.setState({activeExample: 1});
							e.preventDefault();
						}}>
							With CTA
						</a>
					</TabItem>
				</Tabs>

				<Panel hoverLevel="low">
					<TabPanels active={this.state.activeExample}>
						<TabPanel>
							<div className={base.pa4}>
								<ChecklistItem complete={true} className={base.mb5}>
									<Heading level="4">Connect with FooBar</Heading>
									<p className={base.mt2}>
										For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel.
									</p>
								</ChecklistItem>
								<ChecklistItem complete={false} className={base.mb5}>
									<Heading level="4">Connect with FooBar Editor</Heading>
									<p className={base.mt2}>
										For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel.
									</p>
								</ChecklistItem>
							</div>
							<hr />
							<code>
								<pre className={base.ph4} style={{whiteSpace: "pre-wrap"}}>
{
	`
<ChecklistItem complete={true}>
	<Heading level="4">Connect with FooBar</Heading>
	<p className={base.mt2}>
		For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel.
	</p>
</ChecklistItem>
<ChecklistItem complete={false}>
	<Heading level="4">Connect with FooBar Editor</Heading>
	<p className={base.mt2}>
		For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel.
	</p>
</ChecklistItem>
	`
}
								</pre>
							</code>
						</TabPanel>
						<TabPanel>
							<div className={base.pa4}>
								<ChecklistItem complete={true} actionText="Install" actionOnClick={function() { alert("Boo"); }} className={base.mb5}>
									<Heading level="4">Connect with FooBar</Heading>
									<p className={base.mt2}>
										For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel.
									</p>
								</ChecklistItem>
								<ChecklistItem complete={false} actionText="Install" actionOnClick={function() { alert("Boo"); }} className={base.mb5}>
									<Heading level="4">Connect with FooBar Editor</Heading>
									<p className={base.mt2}>
										For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel.
									</p>
								</ChecklistItem>
							</div>
							<hr />
							<code>
								<pre className={base.ph4} style={{whiteSpace: "pre-wrap"}}>
{
	`
<ChecklistItem complete={true} actionText="Install" actionOnClick={function() { alert("Boo"); }}>
	<Heading level="4">Connect with FooBar</Heading>
	<p className={base.mt2}>
		For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel.
	</p>
</ChecklistItem>
<ChecklistItem complete={false} actionText="Install" actionOnClick={function() { alert("Boo"); }}>
	<Heading level="4">Connect with FooBar</Heading>
	<p className={base.mt2}>
		For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel.
	</p>
</ChecklistItem>

	`
}
								</pre>
							</code>
						</TabPanel>
						<TabPanel>
							<div className={base.pa4}>
								<Heading level="7" className={base.mb3} color="cool-mid-gray">Horizontal (default)</Heading>
								<Tabs direction="horizontal">
									<TabItem active={true}>Tab 1</TabItem>
									<TabItem>Tab 2</TabItem>
									<TabItem>Tab 3</TabItem>
								</Tabs>
								<Heading level="7" className={base.mv3} color="cool-mid-gray">Vertical</Heading>
								<Tabs direction="vertical">
									<TabItem active={true}>Tab 1</TabItem>
									<TabItem>Tab 2</TabItem>
									<TabItem>Tab 3</TabItem>
								</Tabs>
							</div>
							<hr />
							<code>
								<pre className={base.ph4} style={{whiteSpace: "pre-wrap"}}>
{
	`
<Tabs direction="horizontal">
	<TabItem active={true}>Tab 1</TabItem>
	<TabItem>Tab 2</TabItem>
	<TabItem>Tab 3</TabItem>
</Tabs>
<Tabs direction="vertical">
	<TabItem active={true}>Tab 1</TabItem>
	<TabItem>Tab 2</TabItem>
	<TabItem>Tab 3</TabItem>
</Tabs>
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
								<td><Code>complete</Code></td>
								<td><Code>null</Code></td>
								<td>
									<Code>true</Code>, <Code>false</Code>
								</td>
							</tr>
							<tr>
								<td><Code>actionText</Code></td>
								<td><Code>null</Code></td>
								<td>
									Any <Code>string</Code>
								</td>
							</tr>
							<tr>
								<td><Code>actionOnClick</Code></td>
								<td><Code>null</Code></td>
								<td>
									<Code>function</Code> to fire when CTA is clicked
								</td>
							</tr>
						</tbody>
					</Table>
				</Panel>
			</div>
		);
	}
}

export default CSSModules(ChecklistsComponent, base, {allowMultiple: true});
