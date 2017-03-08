import * as React from "react";
import { Input, Panel, Select } from "sourcegraph/components";
import * as base from "sourcegraph/components/styles/_base.css";
import { whitespace } from "sourcegraph/components/utils";

interface State {
	activeExample: number;
}

export class FormsComponent extends React.Component<{}, State> {
	state: State = {
		activeExample: 0,
	};

	render(): JSX.Element | null {
		return (

			<div>
				<Panel hoverLevel="low">

					<div className={base.pa4}>

						<Input placeholder="Placeholder text" block={true} label="Input label" helperText="This is optional helper text." className={base.mb4} />
						<Select defaultValue="" label="Select label" containerStyle={{ marginBottom: whitespace[5] }}>
							<option value="" disabled={true}>Placeholder</option>
							<option>Option 1</option>
							<option>Option 2</option>
							<option>Option 3</option>
						</Select>

						<Input placeholder="Placeholder text" block={true} error={true} label="Input label" className={base.mb4} />

						<Input placeholder="Placeholder text" block={true} icon="User" iconPosition="right" label="Small input" compact={true} className={base.mb4} />

						<Input placeholder="Placeholder text" block={true} icon="Search" iconPosition="right" error={true} optionalText="This" label="Input label" errorText="This is an error." className={base.mb4} />

						<Select label="Select label" containerStyle={{ marginBottom: whitespace[5] }} placeholder="Select an option">
							<option>Option 1</option>
							<option>Option 2</option>
							<option>Option 3</option>
						</Select>

						<Select error={true} label="Select label" errorText="This is an error" placeholder="Select an option" containerStyle={{ marginBottom: whitespace[5] }}>
							<option>Option 1</option>
							<option>Option 2</option>
							<option>Option 3</option>
						</Select>

					</div>
					<hr />
					<code>
						<pre className={base.ph4} style={{ whiteSpace: "pre-wrap" }}>
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
			</div>
		);
	}
}
