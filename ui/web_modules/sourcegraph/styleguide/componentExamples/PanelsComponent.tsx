// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as base from "sourcegraph/components/styles/_base.css";
import {Heading, Panel, Tabs, TabItem, TabPanels, TabPanel, Table, Code} from "sourcegraph/components";
import * as classNames from "classnames";

interface State {
	activeExample: number;
}

export class PanelsComponent extends React.Component<{}, State> {
	state: State = {
		activeExample: 0,
	};

	render(): JSX.Element | null {
		return (
			<div className={base.mv4}>
				<Heading level={3} className={base.mb2}>Panels</Heading>

				<Tabs color="purple">
					<TabItem
						active={this.state.activeExample === 0}>
						<a href="#" onClick={(e) => {
							this.setState({activeExample: 0});
							e.preventDefault();
						}}>
							Colors and styles
						</a>
					</TabItem>
					<TabItem
						active={this.state.activeExample === 1}>
						<a href="#" onClick={(e) => {
							this.setState({activeExample: 1});
							e.preventDefault();
						}}>
							Hoverables
						</a>
					</TabItem>
				</Tabs>

				<Panel hoverLevel="low">
					<TabPanels active={this.state.activeExample}>
						<TabPanel>
							<div className={base.pa4}>
								<Heading level={7} className={base.mv3} color="cool_mid_gray">Solid panels</Heading>
								<Panel className={classNames(base.mb3, base.pa3)} color="blue">
									For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time. But conversely, the dolphins had always believed that they were far more intelligent than man—for precisely the same reasons.
								</Panel>
								<Panel className={classNames(base.mb3, base.pa3)} color="purple">
									For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time. But conversely, the dolphins had always believed that they were far more intelligent than man—for precisely the same reasons.
								</Panel>
								<Panel className={classNames(base.mb3, base.pa3)} color="green">
									For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time. But conversely, the dolphins had always believed that they were far more intelligent than man—for precisely the same reasons.
								</Panel>
								<Panel className={classNames(base.mb3, base.pa3)} color="orange">
									For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time. But conversely, the dolphins had always believed that they were far more intelligent than man—for precisely the same reasons.
								</Panel>
								<Heading level={7} className={base.mv3} color="cool_mid_gray">Shadowed Panels</Heading>
								<Panel className={classNames(base.mb3, base.pa3)} hoverLevel="low">
									For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time. But conversely, the dolphins had always believed that they were far more intelligent than man—for precisely the same reasons.
								</Panel>
								<Panel className={classNames(base.mb3, base.pa3)} hoverLevel="high">
									For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time. But conversely, the dolphins had always believed that they were far more intelligent than man—for precisely the same reasons.
								</Panel>
							</div>
							<hr />
							<code>
								<pre className={base.ph4} style={{whiteSpace: "pre-wrap"}}>
{
	`
<Panel color="blue">
	For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time. But conversely, the dolphins had always believed that they were far more intelligent than man—for precisely the same reasons.
</Panel>
<Panel color="purple">
	For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time. But conversely, the dolphins had always believed that they were far more intelligent than man—for precisely the same reasons.
</Panel>
<Panel color="green">
	For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time. But conversely, the dolphins had always believed that they were far more intelligent than man—for precisely the same reasons.
</Panel>
<Panel color="orange">
	For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time. But conversely, the dolphins had always believed that they were far more intelligent than man—for precisely the same reasons.
</Panel>
<Panel hoverLevel="low">
	For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time. But conversely, the dolphins had always believed that they were far more intelligent than man—for precisely the same reasons.
</Panel>
<Panel hoverLevel="high">
	For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time. But conversely, the dolphins had always believed that they were far more intelligent than man—for precisely the same reasons.
</Panel>
	`
}
								</pre>
							</code>
						</TabPanel>
						<TabPanel>
							<div className={base.pa4}>
								<Panel hoverLevel="high" className={classNames(base.mb3, base.pa3)} hover={true}>
									For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time. But conversely, the dolphins had always believed that they were far more intelligent than man—for precisely the same reasons.
								</Panel>
							</div>
							<hr />
							<code>
								<pre className={base.ph4} style={{whiteSpace: "pre-wrap"}}>
{
	`
<Panel hoverLevel="low" hover={true}>
	For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time. But conversely, the dolphins had always believed that they were far more intelligent than man—for precisely the same reasons.
</Panel>
	`
}
								</pre>
							</code>
						</TabPanel>
					</TabPanels>
				</Panel>
				<Heading level={6} className={classNames(base.mt5, base.mb3)}>Properties</Heading>
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
								<td><Code>color</Code></td>
								<td><Code>null</Code></td>
								<td>
									<Code>blue</Code>, <Code>white</Code>, <Code>purple</Code>, <Code>green</Code>, <Code>orange</Code>
								</td>
							</tr>
							<tr>
								<td><Code>inverse</Code></td>
								<td><Code>null</Code></td>
								<td>
									<Code>true</Code>, <Code>false</Code>
								</td>
							</tr>
							<tr>
								<td><Code>hover</Code></td>
								<td><Code>false</Code></td>
								<td>
									<Code>true</Code>, <Code>false</Code>
								</td>
							</tr>
							<tr>
								<td><Code>hoverLevel</Code></td>
								<td><Code>null</Code></td>
								<td>
									<Code>low</Code>, <Code>high</Code>
								</td>
							</tr>
						</tbody>
					</Table>
				</Panel>
			</div>
		);
	}
}
