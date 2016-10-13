import * as React from "react";
import {Heading} from "sourcegraph/components";
import * as base from "sourcegraph/components/styles/_base.css";
import {
	AffixComponent,
	ButtonsComponent,
	ChecklistsComponent,
	FlexContainerComponent,
	FormsComponent,
	HeadingsComponent,
	ListComponent,
	LogoComponent,
	PanelsComponent,
	StepperComponent,
	Symbols,
	TableComponent,
	TabsComponent,
	UserComponent,
} from "sourcegraph/styleguide/componentExamples";

export class ComponentsContainer extends React.Component<{}, any> {
	render(): JSX.Element | null {
		return (
			<div>
				<a id="principles"></a>
				<Heading level={3} underline="purple">Principles</Heading>
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

				<Heading level={2} underline="purple" className={base.mt5}>Brand</Heading>

				<a id="brand-voice"></a>
				<Heading level={3} className={base.mt3}>Voice and Tone</Heading>
				<p>
					All of our writing across the product, codebase, and marketing material should stem from these qualities. Tone is variable and contextual – quality of voice should be consistent.
				</p>
				<ul>
					<li>Intelligent, but not arrogant or patronizing</li>
					<li>Accountable, not hyperbolic</li>
					<li>Authentic, not elitist</li>
					<li>Efficient and concise, but not aloof</li>
					<li>Opinionated, but not overzealous</li>
					<li>Casual, but not unprofessional</li>
				</ul>

				<div className={base.mv5}>
					<a id="brand-logo"></a>
					<LogoComponent />
				</div>

				<Heading level={3} underline="purple" className={base.mt5}>Layout Components</Heading>
				<div className={base.mv5}>
					<a id="layout-flexcontainer"></a>
					<FlexContainerComponent />
				</div>
				<div className={base.mv5}>
					<a id="layout-affix"></a>
					<AffixComponent />
				</div>

				<Heading level={3} underline="purple" className={base.mt5}>UI Components</Heading>
				<div className={base.mv5}>
					<a id="components-buttons"></a>
					<ButtonsComponent />
				</div>
				<div className={base.mv5}>
					<a id="components-checklists"></a>
					<ChecklistsComponent />
				</div>
				<div className={base.mv5}>
					<a id="components-forms"></a>
					<FormsComponent />
				</div>
				<div className={base.mv5}>
					<a id="components-headings"></a>
					<HeadingsComponent />
				</div>
				<div className={base.mv5}>
					<a id="components-list"></a>
					<ListComponent />
				</div>
				<div className={base.mv5}>
					<a id="components-panels"></a>
					<PanelsComponent />
				</div>
				<div className={base.mv5}>
					<a id="components-stepper"></a>
					<StepperComponent />
				</div>
				<div className={base.mv5}>
					<a id="components-symbols"></a>
					<Symbols />
				</div>
				<div className={base.mv5}>
					<a id="components-table"></a>
					<TableComponent />
				</div>
				<div className={base.mv5}>
					<a id="components-tabs"></a>
					<TabsComponent />
				</div>
				<div className={base.mv5}>
					<a id="components-user"></a>
					<UserComponent />
				</div>
			</div>
		);
	}
}
