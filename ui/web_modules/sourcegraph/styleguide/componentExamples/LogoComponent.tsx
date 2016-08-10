// tslint:disable: typedef ordered-imports

import * as React from "react";
import * as base from "sourcegraph/components/styles/_base.css";
import {Heading, Panel, Table, Code, Logo} from "sourcegraph/components/index";
import * as classNames from "classnames";

export class LogoComponent extends React.Component<{}, any> {
	state: {
		activeExample: number,
	};

	constructor(props) {
		super(props);
		this.state = {
			activeExample: 0,
		};
	}

	render(): JSX.Element | null {
		return (
			<div className={base.mv4}>
				<Heading level="3" className={base.mb4}>Logo and Logotype</Heading>

				<Panel hoverLevel="low">
					<div className={base.pa4}>
						<Heading level="7" className={base.mb3} color="cool_mid_gray">Logomark</Heading>
						<Logo width="64px" />
						<Logo width="32px" />
						<Logo width="16px" />
					</div>
					<div className={base.pa4}>
						<Heading level="7" className={base.mb3} color="cool_mid_gray">Logotype</Heading>
						<p><Logo width="128px" type="logotype"/></p>
						<p><Logo width="256px" type="logotype"/></p>
						<p><Logo width="512px" type="logotype"/></p>
					</div>
					<div className={base.pa4}>
						<Heading level="7" className={base.mb3} color="cool_mid_gray">Logotype with tag</Heading>
						<p><Logo width="128px" type="logotype-with-tag"/></p>
						<p><Logo width="256px" type="logotype-with-tag"/></p>
						<p><Logo width="512px" type="logotype-with-tag"/></p>
					</div>
					<hr />
					<code>
						<pre className={base.ph4} style={{whiteSpace: "pre-wrap"}}>
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
								<td><Code>type</Code></td>
								<td><Code>null</Code></td>
								<td>
									<Code>null</Code>, <Code>logotype</Code>, <Code>logotype-with-tag</Code>
								</td>
							</tr>
							<tr>
								<td><Code>width</Code></td>
								<td><Code>null</Code></td>
								<td>
									A <Code>string</Code> of a positive integer appended with <Code>px</Code>
								</td>
							</tr>
						</tbody>
					</Table>
				</Panel>

			</div>
		);
	}
}
