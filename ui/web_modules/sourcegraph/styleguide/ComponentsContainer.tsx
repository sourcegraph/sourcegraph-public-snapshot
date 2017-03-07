import * as React from "react";
import { Heading } from "sourcegraph/components";
import * as base from "sourcegraph/components/styles/_base.css";
import { whitespace } from "sourcegraph/components/utils";
import {
	AffixComponent,
	ButtonsComponent,
	ColorsComponent,
	FlexContainerComponent,
	FormsComponent,
	HeadingsComponent,
	ListComponent,
	LogoComponent,
	OrganizationCardComponent,
	PanelsComponent,
	RepositoryComponent,
	Symbols,
	TableComponent,
	TabsComponent,
	UserComponent,
} from "sourcegraph/styleguide/componentExamples";

export class ComponentsContainer extends React.Component<{}, any> {
	render(): JSX.Element | null {
		return (
			<div>
				<Heading level={3}>Brand</Heading>
				<ComponentContainer name="Principles" id="principles">
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
				</ComponentContainer>

				<ComponentContainer name="Voice and Tone" id="brand-voice">
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
				</ComponentContainer>

				<ComponentContainer name="Logo" id="brand-logo">
					<LogoComponent />
				</ComponentContainer>

				<ComponentContainer name="Colors" id="brand-colors">
					<ColorsComponent />
				</ComponentContainer>

				<Heading level={3}>Layout Components</Heading>
				<ComponentContainer name="FlexContainer" id="layout-flexcontainer">
					<FlexContainerComponent />
				</ComponentContainer>
				<ComponentContainer name="Affix" id="layout-affix">
					<AffixComponent />
				</ComponentContainer>

				<Heading level={3}>UI Components</Heading>
				<ComponentContainer name="Button" id="components-buttons">
					<ButtonsComponent />
				</ComponentContainer>
				<ComponentContainer name="Forms" id="components-forms">
					<FormsComponent />
				</ComponentContainer>
				<ComponentContainer name="Heading" id="components-headings">
					<HeadingsComponent />
				</ComponentContainer>
				<ComponentContainer name="List" id="components-list">
					<ListComponent />
				</ComponentContainer>
				<ComponentContainer name="Panels" id="components-panels">
					<PanelsComponent />
				</ComponentContainer>
				<ComponentContainer name="Symbols" id="components-symbols">
					<Symbols />
				</ComponentContainer>
				<ComponentContainer name="Table" id="components-table">
					<TableComponent />
				</ComponentContainer>
				<ComponentContainer name="Tabs" id="components-tabs">
					<TabsComponent />
				</ComponentContainer>
				<ComponentContainer name="User" id="components-user">
					<UserComponent />
				</ComponentContainer>
				<ComponentContainer name="Repository Card" id="components-repository-card">
					<RepositoryComponent />
				</ComponentContainer>
				<ComponentContainer name="Organization Card" id="components-organization-card">
					<OrganizationCardComponent />
				</ComponentContainer>
			</div>
		);
	}
}

function ComponentContainer({ id, name, children }: {
	id: string,
	name: string,
	children?: React.ReactNode[],
}): JSX.Element {
	return <div style={{ marginTop: whitespace[4], marginBottom: whitespace[5] }}>
		<a id={id}></a>
		<Heading level={4} style={{ marginBottom: whitespace[3] }}>{name}</Heading>
		{children}
	</div>;
}
