// @flow

import * as React from "react";
import CSSModules from "react-css-modules";
import base from "sourcegraph/components/styles/_base.css";
import {Heading, Panel, Tabs, TabItem, TabPanels, TabPanel, Table, Code} from "sourcegraph/components";

class HeadingsComponent extends React.Component {
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
				<Heading level="3" className={base.mb2}>Headings</Heading>

				<Tabs color="purple">
					<TabItem
						active={this.state.activeExample === 0}>
						<a href="#" onClick={(e) => {
							this.setState({activeExample: 0});
							e.preventDefault();
						}}>
							Sizes
						</a>
					</TabItem>
					<TabItem
						active={this.state.activeExample === 1}>
						<a href="#" onClick={(e) => {
							this.setState({activeExample: 1});
							e.preventDefault();
						}}>
							Colors
						</a>
					</TabItem>
					<TabItem
						active={this.state.activeExample === 2}>
						<a href="#" onClick={(e) => {
							this.setState({activeExample: 2});
							e.preventDefault();
						}}>
							Underlines
						</a>
					</TabItem>
					<TabItem
						active={this.state.activeExample === 3}>
						<a href="#" onClick={(e) => {
							this.setState({activeExample: 3});
							e.preventDefault();
						}}>
							Alignment
						</a>
					</TabItem>
				</Tabs>

				<Panel hoverLevel="low">
					<TabPanels active={this.state.activeExample}>
						<TabPanel>
							<div className={base.pa4}>
								<Heading level="1" className={base.mb3}>Heading 1</Heading>
								<Heading level="2" className={base.mb3}>Heading 2</Heading>
								<Heading level="3" className={base.mb3}>Heading 3</Heading>
								<Heading level="4" className={base.mb3}>Heading 4</Heading>
								<Heading level="5" className={base.mb3}>Heading 5</Heading>
								<Heading level="6" className={base.mb3}>Heading 6</Heading>
								<Heading level="7" className={base.mb3}>Heading 7</Heading>
							</div>
							<hr />
							<code>
								<pre className={base.ph4} style={{whiteSpace: "pre-wrap"}}>
{
	`
<Heading level="1">Heading 1</Heading>
<Heading level="2">Heading 2</Heading>
<Heading level="3">Heading 3</Heading>
<Heading level="4">Heading 4</Heading>
<Heading level="5">Heading 5</Heading>
<Heading level="6">Heading 6</Heading>
<Heading level="7">Heading 7</Heading>
	`
}
								</pre>
							</code>
						</TabPanel>
						<TabPanel>
							<div className={base.pa4}>
								<Heading level="7" className={base.mb3} color="cool-mid-gray">Headings with color</Heading>
								<Heading level="4" className={base.mb3} color="blue">Blue fourth level heading</Heading>
								<Heading level="4" className={base.mb3} color="purple">Purple fourth level heading</Heading>
								<Heading level="4" className={base.mb3} color="green">Green fourth level heading</Heading>
								<Heading level="4" className={base.mb3} color="orange">Orange fourth level heading</Heading>
								<Heading level="4" className={base.mb3} color="cool-mid-gray">Mid-gray fourth level heading</Heading>
							</div>
							<hr />
							<code>
								<pre className={base.ph4} style={{whiteSpace: "pre-wrap"}}>
{
	`
<Heading level="4" color="blue">Blue fourth level heading</Heading>
<Heading level="4" color="purple">Purple fourth level heading</Heading>
<Heading level="4" color="green">Green fourth level heading</Heading>
<Heading level="4" color="orange">Orange fourth level heading</Heading>
<Heading level="4" color="cool-mid-gray">Mid-gray fourth level heading</Heading>
	`
}
								</pre>
							</code>
						</TabPanel>
						<TabPanel>
							<div className={base.pa4}>
								<Heading level="7" className={base.mb3} color="cool-mid-gray">Headings with underline</Heading>
								<Heading level="4" className={base.mb3} underline="blue">Fourth level heading with blue underline</Heading>
								<Heading level="4" className={base.mb3} underline="purple">Fourth level heading with purple underline</Heading>
								<Heading level="4" className={base.mb3} underline="orange">Fourth level heading with orange underline</Heading>
								<Heading level="4" className={base.mb3} underline="green">Fourth level heading with green underline</Heading>
								<Heading level="4" className={base.mb3} underline="white">Fourth level heading with white underline</Heading>
							</div>
							<hr />
							<code>
								<pre className={base.ph4} style={{whiteSpace: "pre-wrap"}}>
{
	`
<Heading level="4" underline="blue">Fourth level heading with blue underline</Heading>
<Heading level="4" underline="purple">Fourth level heading with purple underline</Heading>
<Heading level="4" underline="orange">Fourth level heading with orange underline</Heading>
<Heading level="4" underline="green">Fourth level heading with green underline</Heading>
<Heading level="4" underline="white">Fourth level heading with white underline</Heading>
	`
}
								</pre>
							</code>
						</TabPanel>
						<TabPanel>
							<div className={base.pa4}>
								<Heading level="7" className={base.mb3} color="cool-mid-gray">Heading alignment</Heading>
								<Heading level="4" className={base.mb3} align="left">Left aligned</Heading>
								<Heading level="4" className={base.mb3} align="center">Center aligned</Heading>
								<Heading level="4" className={base.mb3} align="right">Right aligned</Heading>
							</div>
							<hr />
							<code>
								<pre className={base.ph4} style={{whiteSpace: "pre-wrap"}}>
{
	`
<Heading level="4" align="left">Left aligned</Heading>
<Heading level="4" align="center">Center aligned</Heading>
<Heading level="4" align="right">Right aligned</Heading>
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
								<td><Code>level</Code></td>
								<td><Code>3</Code></td>
								<td>
									<Code>1</Code>, <Code>2</Code>, <Code>3</Code>, <Code>4</Code>, <Code>5</Code>
								</td>
							</tr>
							<tr>
								<td><Code>color</Code></td>
								<td><Code>null</Code></td>
								<td>
									<Code>purple</Code>, <Code>blue</Code>, <Code>green</Code>, <Code>orange</Code>, <Code>cool-mid-gray</Code>
								</td>
							</tr>
							<tr>
								<td><Code>underline</Code></td>
								<td><Code>null</Code></td>
								<td>
									<Code>blue</Code>, <Code>purple</Code>, <Code>white</Code>, <Code>orange</Code>, <Code>green</Code>
								</td>
							</tr>
							<tr>
								<td><Code>align</Code></td>
								<td><Code>null (inherits from parent)</Code></td>
								<td>
									<Code>left</Code>, <Code>right</Code>, <Code>center</Code>
								</td>
							</tr>
						</tbody>
					</Table>
				</Panel>
			</div>
		);
	}
}

export default CSSModules(HeadingsComponent, base, {allowMultiple: true});
