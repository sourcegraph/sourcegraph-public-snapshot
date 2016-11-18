import * as classNames from "classnames";
import * as React from "react";
import {Code, Heading, Panel, Table, User} from "sourcegraph/components";
import * as base from "sourcegraph/components/styles/_base.css";
import {whitespace} from "sourcegraph/components/utils/index";

export function UserComponent(): JSX.Element {
	return <div className={base.mv4}>
		<Heading level={3}>User</Heading>

		<Panel hoverLevel="low">
			<div className={base.pa4}>

				<Heading level={7} color="gray">User (detailed)</Heading>
				<User nickname="Trillian" email="tmcmillan@gmail.com" avatar="https://placekitten.com/200/300" />

				<Heading level={7} style={{marginTop: whitespace[3]}} color="gray">User (simple)</Heading>
				<User nickname="Trillian" email="tmcmillan@gmail.com" avatar="https://placekitten.com/200/300" simple={true} />

			</div>
			<hr />
			<code>
				<pre className={base.ph4} style={{whiteSpace: "pre-wrap"}}>
{
`
<Heading level={7} color="gray">User (detailed)</Heading>
<User nickname="Trillian" email="tmcmillan@gmail.com" avatar="https://placekitten.com/200/300" />

<Heading level={7} style={{marginTop: whitespace[3]}} color="gray">User (simple)</Heading>
<User nickname="Trillian" email="tmcmillan@gmail.com" avatar="https://placekitten.com/200/300" simple={true} />
`
}
				</pre>
			</code>
		</Panel>
		<Heading level={6} className={classNames(base.mt5, base.mb3)}>User Properties</Heading>
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
						<td><Code>nickname</Code></td>
						<td><Code>null</Code></td>
						<td>
							A <Code>string</Code> of the user's nickname.
						</td>
					</tr>
					<tr>
						<td><Code>email</Code></td>
						<td><Code>null</Code></td>
						<td>
							A <Code>string</Code> of the user's email.
						</td>
					</tr>
					<tr>
						<td><Code>simple</Code></td>
						<td><Code>false</Code></td>
						<td>
							<Code>true</Code>, <Code>false</Code>
						</td>
					</tr>
				</tbody>
			</Table>
		</Panel>
	</div>;
}
