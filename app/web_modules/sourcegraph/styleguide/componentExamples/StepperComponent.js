import * as React from "react";
import CSSModules from "react-css-modules";
import base from "sourcegraph/components/styles/_base.css";
import {Heading, Panel, Table, Code, Stepper} from "sourcegraph/components";

class StepperComponent extends React.Component {
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
				<Heading level="3" className={base.mb2}>Stepper</Heading>

				<Panel hoverLevel="low">
					<div className={base.pa4}>
						<Heading level="7" className={base.mb3} color="cool-mid-gray">Colors</Heading>
						<Stepper steps={[null, null, null, null]} stepsComplete={0} color="blue" />
						<Stepper steps={[null, null, null, null]} stepsComplete={1} color="purple" />
						<Stepper steps={[null, null, null, null]} stepsComplete={2} color="green" />
						<Stepper steps={[null, null, null, null]} stepsComplete={3} color="orange" />

						<Heading level="7" className={base.mv3} color="cool-mid-gray">With Labels</Heading>
						<Stepper steps={["Step 1", "Step 2", "Step 3", "Step 4"]} stepsComplete={4} color="green" />

					</div>
					<hr />
					<code>
						<pre className={base.ph4} style={{whiteSpace: "pre-wrap"}}>
{
`
<Stepper steps={[null, null, null, null]} stepsComplete={0} color="blue" />
<Stepper steps={[null, null, null, null]} stepsComplete={1} color="purple" />
<Stepper steps={[null, null, null, null]} stepsComplete={2} color="green" />
<Stepper steps={[null, null, null, null]} stepsComplete={3} color="orange" />
<Stepper steps={["Step 1", "Step 2", "Step 3", "Step 4"]} stepsComplete={4} color="green" />

`
}
						</pre>
					</code>
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
								<td><Code>color</Code></td>
								<td><Code>green</Code></td>
								<td>
									<Code>green</Code>, <Code>purple</Code>, <Code>blue</Code>, <Code>orange</Code>
								</td>
							</tr>
							<tr>
								<td><Code>steps</Code></td>
								<td><Code>[null, null, null]</Code></td>
								<td>
									An <Code>array</Code> of <Code>null</Code> or <Code>string</Code> values
								</td>
							</tr>
							<tr>
								<td><Code>stepsComplete</Code></td>
								<td><Code>0</Code></td>
								<td>
									Any positive integer that is less than or equal to length of array passed to the <Code>steps</Code> prop
								</td>
							</tr>
						</tbody>
					</Table>
				</Panel>
			</div>
		);
	}
}

export default CSSModules(StepperComponent, base, {allowMultiple: true});
