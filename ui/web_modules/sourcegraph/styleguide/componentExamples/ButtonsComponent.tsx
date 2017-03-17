import * as classNames from "classnames";
import * as React from "react";
import { Button, Heading, Panel, SplitButton, TabItem, TabPanel, TabPanels, Tabs } from "sourcegraph/components";
import * as base from "sourcegraph/components/styles/_base.css";

interface State {
	activeExample: number;
}

export class ButtonsComponent extends React.Component<{}, State> {
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
							Colors and styles
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
				</Tabs>

				<Panel hoverLevel="low">
					<TabPanels active={this.state.activeExample}>
						<TabPanel>
							<div className={base.pa4}>
								<Heading level={7} className={base.mb3} color="blueGray">Solid</Heading>
								<Button className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="blue" className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="purple" className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="green" className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="red" className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="orange" className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button disabled={true} className={classNames(base.mb3, base.mr1)}>Disabled</Button>
							</div>
							<div className={base.pa4}>
								<Heading level={7} className={base.mb3} color="blueGray">Outlined</Heading>
								<Button outline={true} className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="blue" outline={true} className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="purple" outline={true} className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="green" outline={true} className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="red" outline={true} className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="orange" outline={true} className={classNames(base.mb3, base.mr1)}>Submit</Button>
							</div>
							<div className={base.pa4}>
								<Heading level={7} className={base.mb3} color="blueGray">Split Buttons</Heading>
								<SplitButton className={classNames(base.mb3, base.mr1)} secondaryText="For great measure">Submit</SplitButton>
								<SplitButton color="blue" className={classNames(base.mb3, base.mr1)} secondaryText="3 recipients">Send</SplitButton>
								<SplitButton color="purple" className={classNames(base.mb3, base.mr1)} secondaryText="Always free">Sign up</SplitButton>
								<SplitButton color="green" className={classNames(base.mb3, base.mr1)} secondaryText="$100/month">Upgrade</SplitButton>
								<SplitButton color="red" className={classNames(base.mb3, base.mr1)} secondaryText="7 unfixed errors">Deploy</SplitButton>
								<SplitButton color="orange" className={classNames(base.mb3, base.mr1)} secondaryText="19 projects affected">Change</SplitButton>
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
<SplitButton secondaryText="For great measure">Submit</SplitButton>
<SplitButton color="blue" secondaryText="3 recipients">Send</SplitButton>
<SplitButton color="purple" secondaryText="Always free">Sign up</SplitButton>
<SplitButton color="green" secondaryText="$100/month">Upgrade</SplitButton>
<SplitButton color="red" secondaryText="7 unfixed errors">Deploy</SplitButton>
<SplitButton color="orange" secondaryText="19 projects affected">Change</SplitButton>

`
									}
								</pre>
							</code>
						</TabPanel>
						<TabPanel>
							<div className={base.pa4}>
								<Heading level={7} className={base.mb3} color="blueGray">Sizes</Heading>
								<Button color="blue" size="tiny" className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="blue" size="small" className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="blue" className={classNames(base.mb3, base.mr1)}>Submit</Button>
								<Button color="blue" size="large" className={classNames(base.mb3, base.mr1)}>Submit</Button>
							</div>
							<div className={base.pa4}>
								<Heading level={7} className={base.mb3} color="blueGray">Block</Heading>
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
			</div>
		);
	}
}
