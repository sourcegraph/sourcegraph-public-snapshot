import * as React from "react";
import { Heading, Panel, User } from "sourcegraph/components";
import * as base from "sourcegraph/components/styles/_base.css";
import { whitespace } from "sourcegraph/components/utils";

export function UserComponent(): JSX.Element {
	return <div>
		<Panel hoverLevel="low">
			<div className={base.pa4}>

				<Heading level={7} color="gray">User (detailed)</Heading>
				<User nickname="Trillian" email="tmcmillan@gmail.com" avatar="https://placekitten.com/g/200/200" />

				<Heading level={7} style={{ marginTop: whitespace[3] }} color="gray">User (simple)</Heading>
				<User nickname="Trillian" email="tmcmillan@gmail.com" avatar="https://placekitten.com/g/200/200" simple={true} />

			</div>
			<hr />
			<code>
				<pre className={base.ph4} style={{ whiteSpace: "pre-wrap" }}>
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
	</div>;
}
