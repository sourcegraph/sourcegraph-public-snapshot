// @flow

import React from "react";
import CSSModules from "react-css-modules";
import styles from "./styles/ComponentsContainer.css";
import base from "sourcegraph/components/styles/_base.css";
import {Heading, Panel, Stepper, ChecklistItem} from "sourcegraph/components";
import {HeadingsComponent, ButtonsComponent, TabsComponent} from "./componentExamples";

class ComponentsContainer extends React.Component {
	render() {
		return (
			<div>
				<a id="principles"></a>
				<Heading level="2" underline="purple">Principles</Heading>
				<p>
					This styleguide and component library is a living reference to building and designing the Sourcegraph user interface. This reference allows us to build and design conistently, efficiently, and quickly. It's not a definitive framework – but it should follow these main principles:
				</p>
				<ol>
					<li className={base.mb3}>
						<strong>UI components are stateless</strong><br />
						All state and functionality should be handled outside of UI components. <a href="https://medium.com/@dan_abramov/smart-and-dumb-components-7ca2f9a7c7d0#.pk5bjyhmz">Read more about presentational and container components.</a>
					</li>
					<li>
						<strong>Maximise reusability</strong><br />
						Each component should be designed to be used in different contexts, at different widths, on different platforms.
					</li>
				</ol>
				<p>More work on this section TBD.</p>

				<a id="brand-voice"></a>
				<Heading level="2" underline="purple" className={base.mt5}>Voice</Heading>
				<p>
					All of our writing across the product, codebase, and marketing material should stem from these qualities. Tone is variable and contextual – quality of voice should be consistent.
				</p>
				<ul>
					<li>Intelligent, but not arrogant</li>
					<li>Accountable, but not hyperbolic</li>
					<li>Authentic, but not elitist</li>
					<li>Efficient and concise, but not aloof</li>
					<li>Omniscient, but not patronizing</li>
					<li>Opinionated, but not overzealous</li>
					<li>Casual, but not unprofessional</li>
				</ul>

				<Heading level="2" underline="purple" className={base.mt5}>Components</Heading>
				<div className={base.mv5}>
					<a id="components-headings"></a>
					<HeadingsComponent />
				</div>
				<div className={base.mv5}>
					<a id="components-buttons"></a>
					<ButtonsComponent />
				</div>
				<div className={base.mv5}>
					<a id="components-tabs"></a>
					<TabsComponent />
				</div>

				<div>
					<Heading level="3" className={base.mb4}>Panels</Heading>
					<Panel hoverLevel="low" className={base.pa5}>
						For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time. But conversely, the dolphins had always believed that they were far more intelligent than man—for precisely the same reasons.
					</Panel>
					<br />
					<Panel hoverLevel="high" hover={true} className={base.pa5}>
						For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time. But conversely, the dolphins had always believed that they were far more intelligent than man—for precisely the same reasons.
					</Panel>
					<br/>
					<Panel color="blue" className={base.pa5}>
						For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time. But conversely, the dolphins had always believed that they were far more intelligent than man—for precisely the same reasons.
					</Panel>
					<br/>
					<Panel color="blue" inverse={true} className={base.pa5}>
						For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time. But conversely, the dolphins had always believed that they were far more intelligent than man—for precisely the same reasons.
					</Panel>
				</div>
				<div>
					<Heading level="3" className={base.mb4}>Stepper</Heading>
					<Panel className={base.pa5}><Stepper steps={[null, null, null, null]} stepsComplete={2} color="green" /></Panel>
				</div>
				<div>
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
				</div>
			</div>
		);
	}
}

export default CSSModules(ComponentsContainer, styles, {allowMultiple: true});
