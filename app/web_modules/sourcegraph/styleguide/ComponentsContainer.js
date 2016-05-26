// @flow

import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/ComponentsContainer.css";
import base from "sourcegraph/components/styles/_base.css";
import {Heading, Panel, Stepper, ChecklistItem, TabItem, Button} from "sourcegraph/components";
import ComponentCard from "./ComponentCard";

class ComponentsContainer extends React.Component {
	render() {
		return (
			<div styleName="container">
				<div styleName="container-fixed">
					<ComponentCard>
						<Heading level="3" className={base.mb4}>Headings</Heading>
						<Panel>
							<div className={base.pa4}>

								<Heading level="5" className={base.mb3} color="cool-mid-gray">Heading sizes</Heading>
								<Heading level="1" className={base.mb3}>Heading 1</Heading>
								<Heading level="2" className={base.mb3}>Heading 2</Heading>
								<Heading level="3" className={base.mb3}>Heading 3</Heading>
								<Heading level="4" className={base.mb3}>Heading 4</Heading>
								<Heading level="5" className={base.mb3}>Heading 5</Heading>

								<Heading level="5" className={base.mv4} color="cool-mid-gray">Headings with color</Heading>
								<Heading level="4" className={base.mb3} color="blue">Blue fourth level heading</Heading>
								<Heading level="4" className={base.mb3} color="purple">Purple fourth level heading</Heading>
								<Heading level="4" className={base.mb3} color="orange">Orange fourth level heading</Heading>
								<Heading level="4" className={base.mb3} color="cool-mid-gray">Mid-gray fourth level heading</Heading>

								<Heading level="5" className={base.mv4} color="cool-mid-gray">Headings with underline</Heading>
								<Heading level="3" className={base.mb3} underline="blue">Third level heading with blue underline</Heading>
								<Heading level="4" className={base.mb3} underline="purple">Fourth level heading with purple underline</Heading>

								<Heading level="5" className={base.mv4} color="cool-mid-gray">Alignment</Heading>
								<Heading level="3" className={base.mb3} underline="blue" align="center">Third level heading with blue underline, aligned center</Heading>

							</div>
							<hr />
							<div className={base.pa4}>
								<Heading level="5" className={base.mb3} color="cool-mid-gray">Usage</Heading>
								<pre>
									{"<Heading level=\"1\" color=\"blue\" underline=\"blue\">Heading 1</Heading>"}
								</pre>
							</div>
						</Panel>
					</ComponentCard>
					<ComponentCard>
						<Panel className={base.pa5}><Button color="disabled">Disabled button</Button></Panel>
					</ComponentCard>
					<ComponentCard>
						<Heading level="3" className={base.mb4}>Tabs</Heading>
						<Panel className={base.pa5}>
							<div>
								<TabItem active={true}>Components</TabItem>
								<TabItem>Colors</TabItem>
								<TabItem>Typography</TabItem>
								<TabItem>Layout</TabItem>
							</div>
						</Panel>
					</ComponentCard>
					<ComponentCard>
						<Heading level="3" className={base.mb4}>Panels</Heading>
						<Panel hoverLevel="low" hover={false} className={base.pa5}>
							For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time. But conversely, the dolphins had always believed that they were far more intelligent than man—for precisely the same reasons.
						</Panel>
						<br />
						<Panel hoverLevel="high" hover={true} className={base.pa5}>
							For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time. But conversely, the dolphins had always believed that they were far more intelligent than man—for precisely the same reasons.
						</Panel>
					</ComponentCard>
					<ComponentCard>
						<Heading level="3" className={base.mb4}>Stepper</Heading>
						<Panel className={base.pa5}><Stepper steps={[null, null, null, null]} stepsComplete={2} color="green" /></Panel>
					</ComponentCard>
					<ComponentCard>
						<Heading level="3" className={base.mb4}>Checklist Items</Heading>
						<Panel className={base.pa5}>
							<ChecklistItem complete={true} className={base.mb5}>
								<Heading level="4">Connect with FooBar</Heading>
								<p className={base.mt2}>
									For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel.
								</p>
							</ChecklistItem>
							<ChecklistItem actionText="Install" actionOnClick={function() { alert("Boo"); }}>
								<Heading level="4">Connect with FooBar</Heading>
								<p className={base.mt2}>
									For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel.
								</p>
							</ChecklistItem>
						</Panel>
					</ComponentCard>
				</div>
			</div>
		);
	}
}

export default CSSModules(ComponentsContainer, styles, {allowMultiple: true});
