import * as React from "react";
import { Heading, Panel, TabItem, TabPanel, TabPanels, Tabs } from "sourcegraph/components";
import * as base from "sourcegraph/components/styles/_base.css";

interface State {
	activeExample: number;
}

export class TabsComponent extends React.Component<{}, State> {
	state: State = {
		activeExample: 0,
	};

	render(): JSX.Element | null {
		return (
			<div>
				<Tabs>
					<TabItem
						active={this.state.activeExample === 0}>
						<a href="#" onClick={(e) => {
							this.setState({ activeExample: 0 });
							e.preventDefault();
						}}>
							Colors
						</a>
					</TabItem>
					<TabItem
						active={this.state.activeExample === 1}>
						<a href="#" onClick={(e) => {
							this.setState({ activeExample: 1 });
							e.preventDefault();
						}}>
							Sizes
						</a>
					</TabItem>
					<TabItem
						active={this.state.activeExample === 2}>
						<a href="#" onClick={(e) => {
							this.setState({ activeExample: 2 });
							e.preventDefault();
						}}>
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
									<TabItem hideMobile={true}>Tab 2</TabItem>
									<TabItem>Tab 3</TabItem>
								</Tabs>
								<Heading level={7} className={base.mv3} color="cool_mid_gray">Purple</Heading>
								<Tabs>
									<TabItem color="purple" active={true}>Tab 1</TabItem>
									<TabItem color="purple">Tab 2</TabItem>
									<TabItem color="purple">Tab 3</TabItem>
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
<Tabs>
	<TabItem color="purple" active={true}>Tab 1</TabItem>
	<TabItem color="purple">Tab 2</TabItem>
	<TabItem color="purple">Tab 3</TabItem>
</Tabs>
	`
									}
								</pre>
							</code>
						</TabPanel>
						<TabPanel>
							<div className={base.pa4}>
								<Heading level={7} className={base.mb3} color="cool_mid_gray">Small</Heading>
								<Tabs>
									<TabItem size="small" active={true}>Tab 1</TabItem>
									<TabItem size="small">Tab 2</TabItem>
									<TabItem size="small">Tab 3</TabItem>
								</Tabs>
								<Heading level={7} className={base.mv3} color="cool_mid_gray">Default</Heading>
								<Tabs>
									<TabItem active={true}>Tab 1</TabItem>
									<TabItem>Tab 2</TabItem>
									<TabItem>Tab 3</TabItem>
								</Tabs>
								<Heading level={7} className={base.mv3} color="cool_mid_gray">Large</Heading>
								<Tabs>
									<TabItem size="large" active={true}>Tab 1</TabItem>
									<TabItem size="large">Tab 2</TabItem>
									<TabItem size="large">Tab 3</TabItem>
								</Tabs>
							</div>
							<hr />
							<code>
								<pre className={base.ph4} style={{ whiteSpace: "pre-wrap" }}>
									{
										`
<Tabs>
	<TabItem active={true}>Tab 1</TabItem>
	<TabItem size="small">Tab 2</TabItem>
	<TabItem size="small">Tab 3</TabItem>
</Tabs>
<Tabs>
	<TabItem active={true}>Tab 1</TabItem>
	<TabItem>Tab 2</TabItem>
	<TabItem>Tab 3</TabItem>
</Tabs>
<Tabs>
	<TabItem size="large" active={true}>Tab 1</TabItem>
	<TabItem size="large">Tab 2</TabItem>
	<TabItem size="large">Tab 3</TabItem>
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
									<TabItem direction="vertical" active={true}>Tab 1</TabItem>
									<TabItem direction="vertical">Tab 2</TabItem>
									<TabItem direction="vertical">Tab 3</TabItem>
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
	<TabItem direction="vertical">Tab 2</TabItem>
	<TabItem direction="vertical">Tab 3</TabItem>
</Tabs>
	`
									}
								</pre>
							</code>
						</TabPanel>
					</TabPanels>
				</Panel>
			</div>
		);
	}
}
