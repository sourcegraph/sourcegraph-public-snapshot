// tslint:disable

import * as React from "react";
import * as base from "sourcegraph/components/styles/_base.css";
import {Heading, Panel, Table, Code, Input, Select} from "sourcegraph/components/index";
import * as classNames from "classnames";

class StepperComponent extends React.Component<{}, any> {
	constructor(props) {
		super(props);
		this.state = {
			activeExample: 0,
		};
	}

	state: {
		activeExample: number,
	};

	render(): JSX.Element | null {
		return (

			<div className={base.mv4}>
				<Heading level="3" className={base.mb2}>Forms</Heading>

				<Panel hoverLevel="low">

					<div className={base.pa4}>

						<Input placeholder="Placeholder text" block={true} label="Input label" helperText="This is optional helper text." className={base.mb4} />
						<Select defaultValue="" label="Select label" className={base.mb4}>
							<option value="" disabled={true}>Placeholder</option>
							<option>Option 1</option>
							<option>Option 2</option>
							<option>Option 3</option>
						</Select>

						<Input placeholder="Placeholder text" block={true} error={true} label="Input label" className={base.mb4} />

						<Input placeholder="Placeholder text" block={true} error={true} label="Input label" errorText="This is an error." className={base.mb4} />

						<Select defaultValue="" error={true} label="Select label" className={base.mb4}>
							<option value="" disabled={true}>Placeholder</option>
							<option>Option 1</option>
							<option>Option 2</option>
							<option>Option 3</option>
						</Select>

						<Select defaultValue="" error={true} label="Select label" errorText="This is an error" className={base.mb4}>
							<option value="" disabled={true}>Placeholder</option>
							<option>Option 1</option>
							<option>Option 2</option>
							<option>Option 3</option>
						</Select>

					</div>
					<hr />
					<code>
						<pre className={base.ph4} style={{whiteSpace: "pre-wrap"}}>
{
`
<Input placeholder="Placeholder text" block={true} label="Input label" helperText="This is optional helper text." className={base.mb4} />
<Select defaultValue="" label="Select label">
	<option value="" disabled={true}>Placeholder</option>
	<option>Option 1</option>
	<option>Option 2</option>
	<option>Option 3</option>
</Select>

<Input placeholder="Placeholder text" block={true} error={true} label="Input label" className={base.mb4} />

<Input placeholder="Placeholder text" block={true} error={true} label="Input label" errorText="This is an error." className={base.mb4} />

<Select defaultValue="" error={true} label="Select label" className={base.mb4}>
	<option value="" disabled={true}>Placeholder</option>
	<option>Option 1</option>
	<option>Option 2</option>
	<option>Option 3</option>
</Select>

<Select defaultValue="" error={true} label="Select label" errorText="This is an error" className={base.mb4}>
	<option value="" disabled={true}>Placeholder</option>
	<option>Option 1</option>
	<option>Option 2</option>
	<option>Option 3</option>
</Select>

`
}
						</pre>
					</code>
				</Panel>
				<Heading level="4" className={classNames(base.mt5, base.mb3)}>Properties</Heading>
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

export default StepperComponent;
