import * as classNames from "classnames";
import * as React from "react";
import { Code, Heading, Panel, Table } from "sourcegraph/components";
import * as base from "sourcegraph/components/styles/_base.css";
import { Alert, Cone, Dismiss, GitHubLogo, Search, Spinner } from "sourcegraph/components/symbols";
import { Check, CheckCircle, ChevronDown, Close, Flag, Link, UserAdd } from "sourcegraph/components/symbols/Zondicons";
import { colors, whitespace } from "sourcegraph/components/utils/index";

export class Symbols extends React.Component<{}, any> {

	render(): JSX.Element | null {
		return (
			<div className={base.mv4}>
				<Heading level={3} className={base.mb2}>Symbols</Heading>

				<Panel hoverLevel="low">
					<div className={base.pa4}>

						<Alert color={colors.coolGray2()} style={{ padding: whitespace[2] }} />
						<Cone color={colors.coolGray2()} style={{ padding: whitespace[2] }} />
						<Dismiss color={colors.coolGray2()} style={{ padding: whitespace[2] }} />
						<GitHubLogo color={colors.coolGray2()} style={{ padding: whitespace[2] }} />
						<Search color={colors.coolGray2()} style={{ padding: whitespace[2] }} />
						<Spinner color={colors.coolGray2()} style={{ padding: whitespace[2] }} />
						<Flag color={colors.coolGray2()} style={{ padding: whitespace[2] }} />
						<ChevronDown color={colors.coolGray2()} style={{ padding: whitespace[2] }} />
						<Close color={colors.coolGray2()} style={{ padding: whitespace[2] }} />
						<Check color={colors.coolGray2()} style={{ padding: whitespace[2] }} />
						<CheckCircle color={colors.coolGray2()} style={{ padding: whitespace[2] }} />
						<Link color={colors.coolGray2()} style={{ padding: whitespace[2] }} />
						<UserAdd color={colors.coolGray2()} style={{ padding: whitespace[2] }} />
						<p>
							These symbols are each individual components that share the same props api. See <Code>/components/symbols</Code> for component usage.
						</p>

					</div>
					<hr />
					<code>
						<pre className={base.ph4} style={{ whiteSpace: "pre-wrap" }}>
							{
								`
<Alert color={colors.coolGray2()} style={{padding: whitespace[2]}} />
<Cone color={colors.coolGray2()} style={{padding: whitespace[2]}} />
<Dismiss color={colors.coolGray2()} style={{padding: whitespace[2]}} />
<GitHubLogo color={colors.coolGray2()} style={{padding: whitespace[2]}} />
<Search color={colors.coolGray2()} style={{padding: whitespace[2]}} />
<Spinner color={colors.coolGray2()} style={{padding: whitespace[2]}} />
<Flag color={colors.coolGray2()} style={{padding: whitespace[2]}} />
<ChevronDown color={colors.coolGray2()} style={{padding: whitespace[2]}} />
<Close color={colors.coolGray2()} style={{padding: whitespace[2]}} />
<Check color={colors.coolGray2()} style={{padding: whitespace[2]}} />
<CheckCircle color={colors.coolGray2()} style={{padding: whitespace[2]}} />
<Link color={colors.coolGray2()} style={{padding: whitespace[2]}} />
<UserAdd color={colors.coolGray2()} style={{padding: whitespace[2]}} />
`
							}
						</pre>
					</code>
				</Panel>
				<Heading level={6} className={classNames(base.mt5, base.mb3)}>Properties</Heading>
				<Panel hoverLevel="low" className={base.pa4}>
					<Table style={{ width: "100%" }}>
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
								<td><Code>inherit (black)</Code></td>
								<td>
									any color import from <Code>colors.tsx</Code>
								</td>
							</tr>
							<tr>
								<td><Code>width</Code></td>
								<td><Code>16</Code></td>
								<td>
									any number
								</td>
							</tr>
						</tbody>
					</Table>
				</Panel>
			</div>
		);
	}
}
