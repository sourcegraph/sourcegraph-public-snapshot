// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as base from "sourcegraph/components/styles/_base.css";
import { Button, Heading, Panel, Tabs, TabItem, TabPanels, TabPanel, Table, Code } from "sourcegraph/components";
import * as classNames from "classnames";

interface State {
	activeExample: number;
}

export class ButtonsComponent extends React.Component<{}, State> {
	state: State = {
		activeExample: 0,
	};

	render(): JSX.Element | null {
		return (
			<div className={base.mv4}>
				<Heading level={3} className={base.mb2}>Buttons</Heading>

				<Tabs color="purple">
					<TabItem
						active={this.state.activeExample === 0}>
						<a href="#" onClick={(e) => {
							this.setState({ activeExample: 0 });
							e.preventDefault();
						} }>
							Colors and styles
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
				</Tabs>

				<Panel hoverLevel="low">
					<TabPanels active={this.state.activeExample}>
						<TabPanel>
							<div className={base.pa4}>
								<Heading level={7} className={base.mb3} color="cool_mid_gray">Solid</Heading>
								<Button className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="blue" className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="purple" className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="green" className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="red" className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="orange" className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button disabled={true} className={classNames(base.mb3, base.mr1)}>Disabled</Button>
							</div>
							<div className={base.pa4}>
								<Heading level={7} className={base.mb3} color="cool_mid_gray">Outlined</Heading>
								<Button outline={true} className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="blue" outline={true} className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="purple" outline={true} className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="green" outline={true} className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="red" outline={true} className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="orange" outline={true} className={classNames(base.mb3, base.mr1)}>Submit</Button>
							</div>
							<hr />
							<code>
								<pre className={base.ph4} style={{ whiteSpace: "pre-wrap" }}>
									{
										`
<Button>Submit</Button>
<Button color="blue">Submit</Button>
<Button color="purple">Submit</Button>
<Button color="green">Submit</Button>
<Button color="red">Submit</Button>
<Button color="orange">Submit</Button>
<Button outline={true}>Submit</Button>
<Button disable={true}>Disable</Button>
<Button color="blue" outline={true}>Submit</Button>
<Button color="purple" outline={true}>Submit</Button>
<Button color="green" outline={true}>Submit</Button>
<Button color="red" outline={true}>Submit</Button>
<Button color="orange" outline={true}>Submit</Button>
	`
									}
								</pre>
							</code>
						</TabPanel>
						<TabPanel>
							<div className={base.pa4}>
								<Heading level={7} className={base.mb3} color="cool_mid_gray">Sizes</Heading>
								<Button color="blue" size="small" className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="blue" className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="blue" size="large" className={classNames(base.mb3, base.mr1)}>Submit</Button>
							</div>
							<div className={base.pa4}>
								<Heading level={7} className={base.mb3} color="cool_mid_gray">Block</Heading>
								<Button color="blue" block={true} className={base.mb3}>Submit</Button>
							</div>
							<hr />
							<code>
								<pre className={base.ph4} style={{ whiteSpace: "pre-wrap" }}>
									{
										`
<Button color="blue" size="small">Submit</Button>
<Button color="blue">Submit</Button>
<Button color="blue" size="large">Submit</Button>
	`
									}
								</pre>
							</code>
						</TabPanel>
					</TabPanels>
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
								<td><Code>color</Code></td>
								<td><Code>default</Code></td>
								<td>
									<Code>default</Code>, <Code>blue</Code>, <Code>purple</Code>, <Code>green</Code>, <Code>red</Code>, <Code>orange</Code>
								</td>
							</tr>
							<tr>
								<td><Code>outline</Code></td>
								<td><Code>null</Code></td>
								<td>
									<Code>true</Code>, <Code>false</Code>
								</td>
							</tr>
							<tr>
								<td><Code>size</Code></td>
								<td><Code>null</Code></td>
								<td>
									<Code>small</Code>, <Code>large</Code>, <Code>null</Code>
								</td>
							</tr>
							<tr>
								<td><Code>block</Code></td>
								<td><Code>null</Code></td>
								<td>
									<Code>true</Code>, <Code>false</Code>
								</td>
							</tr>
							<tr>
								<td><Code>loading</Code></td>
								<td><Code>null</Code></td>
								<td>
									<Code>true</Code>, <Code>false</Code>
								</td>
							</tr>
						</tbody>
					</Table>
				</Panel>
			</div>
		);
	}
}
