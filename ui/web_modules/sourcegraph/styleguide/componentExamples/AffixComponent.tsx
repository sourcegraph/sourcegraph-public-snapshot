import * as React from "react";
import { Code, Panel } from "sourcegraph/components";
import * as base from "sourcegraph/components/styles/_base.css";

export class AffixComponent extends React.Component<{}, any> {

	render(): JSX.Element | null {
		return (
			<div>
				<Panel hoverLevel="low">
					<div className={base.pa4}>
						<p>The <Code>Affix</Code> component makes any content inside of it <Code>position:fixed</Code> once you've scrolled past it's original position. If you scroll back, the content returns to it's original position.</p>
						<p>See the menu to the right for a working example.</p>
					</div>
					<hr />
					<code>
						<pre className={base.ph4} style={{ whiteSpace: "pre-wrap" }}>
							{
								`
<Affix offset={20}>
	...
</Affix>

`
							}
						</pre>
					</code>
				</Panel>
			</div>
		);
	}
}
