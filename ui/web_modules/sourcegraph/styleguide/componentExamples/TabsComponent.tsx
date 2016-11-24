// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as base from "sourcegraph/components/styles/_base.css";
import { Heading, Panel, Tabs, TabItem, TabPanels, TabPanel, Table, Code } from "sourcegraph/components";
import * as classNames from "classnames";

interface State {
	activeExample: number;
}

export class TabsComponent extends React.Component<{}, State> {
	state: State = {
		activeExample: 0,
	};

	render(): JSX.Element | null {
		return (
			<div className={base.mv4}>
				<Heading level={3} className={base.mb2}>Tabs</Heading>

				<Tabs color="purple">
					<TabItem
						active={this.state.activeExample === 0}>
						<a href="#" onClick={(e) => {
							this.setState({ activeExample: 0 });
							e.preventDefault();
						} }>
							Colors
						</a>
					</TabItem>
					<TabItem
						active={this.state.activeExample === 1}>
						<a href="#" onClick={(e) => {
							this.setState({ activeExample: 1 });
							e.preventDefault();
						} }>
							Sizes
						</a>
					</TabItem>
					<TabItem
						active={this.state.activeExample === 2}>
						<a href="#" onClick={(e) => {
							this.setState({ activeExample: 2 });
							e.preventDefault();
						} }>
							Orientation
						</a>
					</TabItem>
				</Tabs>

				<Panel hoverLevel="low">
					<TabPanels active={this.state.activeExample}>
						<TabPanel>
							<div className={base.pa4}>
								<Heading level={7} className={base.mb3} color="cool_mid_gray">Default (Blue)</Heading>
								<Tabs>
									<TabItem active={true}>Tab 1</TabItem>
									<TabItem>Tab 2</TabItem>
									<TabItem>Tab 3</TabItem>
								</Tabs>
								<Heading level={7} className={base.mv3} color="cool_mid_gray">Purple</Heading>
								<Tabs color="purple">
									<TabItem active={true}>Tab 1</TabItem>
									<TabItem>Tab 2</TabItem>
									<TabItem>Tab 3</TabItem>
								</Tabs>
							</div>
							<hr />
							<code>
								<pre className={base.ph4} style={{ whiteSpace: "pre-wrap" }}>
									{
										`
<Tabs>
	<TabItem active={true}>Tab 1</TabItem>
	<TabItem>Tab 2</TabItem>
	<TabItem>Tab 3</TabItem>
</Tabs>
<Tabs color="purple">
	<TabItem active={true}>Tab 1</TabItem>
	<TabItem>Tab 2</TabItem>
	<TabItem>Tab 3</TabItem>
</Tabs>
	`
									}
								</pre>
							</code>
						</TabPanel>
						<TabPanel>
							<div className={base.pa4}>
								<Heading level={7} className={base.mb3} color="cool_mid_gray">Small</Heading>
								<Tabs size="small">
									<TabItem active={true}>Tab 1</TabItem>
									<TabItem>Tab 2</TabItem>
									<TabItem>Tab 3</TabItem>
								</Tabs>
								<Heading level={7} className={base.mv3} color="cool_mid_gray">Default</Heading>
								<Tabs>
									<TabItem active={true}>Tab 1</TabItem>
									<TabItem>Tab 2</TabItem>
									<TabItem>Tab 3</TabItem>
								</Tabs>
								<Heading level={7} className={base.mv3} color="cool_mid_gray">Large</Heading>
								<Tabs size="large">
									<TabItem active={true}>Tab 1</TabItem>
									<TabItem>Tab 2</TabItem>
									<TabItem>Tab 3</TabItem>
								</Tabs>
							</div>
							<hr />
							<code>
								<pre className={base.ph4} style={{ whiteSpace: "pre-wrap" }}>
									{
										`
<Tabs size="small">
	<TabItem active={true}>Tab 1</TabItem>
	<TabItem>Tab 2</TabItem>
	<TabItem>Tab 3</TabItem>
</Tabs>
<Tabs>
	<TabItem active={true}>Tab 1</TabItem>
	<TabItem>Tab 2</TabItem>
	<TabItem>Tab 3</TabItem>
</Tabs>
<Tabs size="large">
	<TabItem active={true}>Tab 1</TabItem>
	<TabItem>Tab 2</TabItem>
	<TabItem>Tab 3</TabItem>
</Tabs>

	`
									}
								</pre>
							</code>
						</TabPanel>
						<TabPanel>
							<div className={base.pa4}>
								<Heading level={7} className={base.mb3} color="cool_mid_gray">Horizontal (default)</Heading>
								<Tabs direction="horizontal">
									<TabItem active={true}>Tab 1</TabItem>
									<TabItem>Tab 2</TabItem>
									<TabItem>Tab 3</TabItem>
								</Tabs>
								<Heading level={7} className={base.mv3} color="cool_mid_gray">Vertical</Heading>
								<Tabs direction="vertical">
									<TabItem active={true}>Tab 1</TabItem>
									<TabItem>Tab 2</TabItem>
									<TabItem>Tab 3</TabItem>
								</Tabs>
							</div>
							<hr />
							<code>
								<pre className={base.ph4} style={{ whiteSpace: "pre-wrap" }}>
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
				<Heading level={6} className={classNames(base.mt5, base.mb3)}>Tabs Properties</Heading>
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
								<td><Code>color</Code></td>
								<td><Code>blue</Code></td>
								<td>
									<Code>blue</Code>, <Code>purple</Code>
								</td>
							</tr>
							<tr>
								<td><Code>size</Code></td>
								<td><Code>null</Code></td>
								<td>
									<Code>small</Code>, <Code>large</Code>
								</td>
							</tr>
							<tr>
								<td><Code>direction</Code></td>
								<td><Code>horizontal</Code></td>
								<td>
									<Code>horizontal</Code>, <Code>vertical</Code>
								</td>
							</tr>
						</tbody>
					</Table>
				</Panel>
				<Heading level={6} className={classNames(base.mt5, base.mb3)}>TabItem Properties</Heading>
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
								<td><Code>active</Code></td>
								<td><Code>false</Code></td>
								<td>
									<Code>true</Code>, <Code>false</Code>
								</td>
							</tr>
							<tr>
								<td><Code>hideMobile</Code></td>
								<td><Code>null</Code></td>
								<td>
									<Code>true</Code>, <Code>false</Code>
								</td>
							</tr>
							<tr>
								<td><Code>icon</Code></td>
								<td><Code>null</Code></td>
								<td>
									<Code>Any icon rendered from Icon</Code>
								</td>
							</tr>
							<tr>
								<td><Code>size</Code></td>
								<td><Code>null</Code></td>
								<td>
									<Code>small</Code>, <Code>large</Code>, <Code>Inherited from Tabs</Code>
								</td>
							</tr>
							<tr>
								<td><Code>direction</Code></td>
								<td><Code>horizontal</Code></td>
								<td>
									<Code>horizontal</Code>, <Code>vertical</Code>, <Code>Inherited from Tabs</Code>
								</td>
							</tr>
						</tbody>
					</Table>
				</Panel>
			</div>
		);
	}
}
