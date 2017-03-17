import * as React from "react";
import { Heading, Logo, Panel } from "sourcegraph/components";
import * as base from "sourcegraph/components/styles/_base.css";

interface State {
	activeExample: number;
}

export class LogoComponent extends React.Component<{}, State> {
	state: State = {
		activeExample: 0,
	};

	render(): JSX.Element | null {
		return (
			<div>
				<Panel hoverLevel="low">
					<div className={base.pa4}>
						<Heading level={7} className={base.mb3} color="cool_mid_gray">Logomark</Heading>
						<Logo width="64px" />
						<Logo width="32px" />
						<Logo width="16px" />
					</div>
					<div className={base.pa4}>
						<Heading level={7} className={base.mb3} color="cool_mid_gray">Logotype</Heading>
						<p><Logo width="128px" type="logotype" /></p>
						<p><Logo width="256px" type="logotype" /></p>
						<p><Logo width="512px" type="logotype" /></p>
					</div>
					<div className={base.pa4}>
						<Heading level={7} className={base.mb3} color="cool_mid_gray">Logotype with tag</Heading>
						<p><Logo width="128px" type="logotype-with-tag" /></p>
						<p><Logo width="256px" type="logotype-with-tag" /></p>
						<p><Logo width="512px" type="logotype-with-tag" /></p>
					</div>
					<hr />
					<code>
						<pre className={base.ph4} style={{ whiteSpace: "pre-wrap" }}>
							{
								`
<Logo width="64px" />
<Logo width="32px" />
<Logo width="16px" />
<Logo width="128px" type="logotype"/>
<Logo width="256px" type="logotype"/>
<Logo width="512px" type="logotype"/>
<Logo width="128px" type="logotype-with-tag"/>
<Logo width="256px" type="logotype-with-tag"/>
<Logo width="512px" type="logotype-with-tag"/>
	`
							}
						</pre>
					</code>
				</Panel>
			</div>
		);
	}
}
