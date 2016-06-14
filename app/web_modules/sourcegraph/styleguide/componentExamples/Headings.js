// @flow

import React from "react";
import CSSModules from "react-css-modules";
import base from "sourcegraph/components/styles/_base.css";
import {Heading, Panel, Tabs, TabItem, TabPanels, TabPanel} from "sourcegraph/components";

class Headings extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
			activeExample: 0,
		};
	}

	render() {
		return (
			<div className={base.mv4}>
				<a name="components-headings"></a>
				<Heading level="2" className={base.mb4}>Headings</Heading>

				<Tabs>
					<TabItem
						active={this.state.activeExample === 0}>
						<a href="#" onClick={(e) => {
							this.setState({activeExample: 0});
							e.preventDefault();
						}}>
							Heading sizes
						</a>
					</TabItem>
					<TabItem
						active={this.state.activeExample === 1}>
						<a href="#" onClick={(e) => {
							this.setState({activeExample: 1});
							e.preventDefault();
						}}>
							Heading colors
						</a>
					</TabItem>
					<TabItem
						active={this.state.activeExample === 2}>
						<a href="#" onClick={(e) => {
							this.setState({activeExample: 2});
							e.preventDefault();
						}}>
							Headings with underlines
						</a>
					</TabItem>
				</Tabs>

				<Panel>
					<TabPanels active={this.state.activeExample}>
						<TabPanel>
							<div className={base.pa4}>
								<Heading level="5" className={base.mb3} color="cool-mid-gray">Heading sizes</Heading>
								<Heading level="1" className={base.mb3}>Heading 1</Heading>
								<Heading level="2" className={base.mb3}>Heading 2</Heading>
								<Heading level="3" className={base.mb3}>Heading 3</Heading>
								<Heading level="4" className={base.mb3}>Heading 4</Heading>
								<Heading level="5" className={base.mb3}>Heading 5</Heading>
							</div>
						</TabPanel>
						<TabPanel>
							<div className={base.pa4}>
								<Heading level="5" className={base.mb3} color="cool-mid-gray">Headings with color</Heading>
								<Heading level="4" className={base.mb3} color="blue">Blue fourth level heading</Heading>
								<Heading level="4" className={base.mb3} color="purple">Purple fourth level heading</Heading>
								<Heading level="4" className={base.mb3} color="orange">Orange fourth level heading</Heading>
								<Heading level="4" className={base.mb3} color="cool-mid-gray">Mid-gray fourth level heading</Heading>
							</div>
						</TabPanel>
						<TabPanel>
							<div className={base.pa4}>
								<Heading level="5" className={base.mb3} color="cool-mid-gray">Headings with underline</Heading>
								<Heading level="3" className={base.mb3} underline="blue">Third level heading with blue underline</Heading>
								<Heading level="4" className={base.mb3} underline="purple">Fourth level heading with purple underline</Heading>

								<Heading level="5" className={base.mv4} color="cool-mid-gray">Alignment</Heading>
								<Heading level="3" className={base.mb3} underline="blue" align="center">Third level heading with blue underline, aligned center</Heading>
							</div>
						</TabPanel>
						<TabPanel>This 4 this 4</TabPanel>
					</TabPanels>
				</Panel>
			</div>
		);
	}
}

export default CSSModules(Headings, {allowMultiple: true});
