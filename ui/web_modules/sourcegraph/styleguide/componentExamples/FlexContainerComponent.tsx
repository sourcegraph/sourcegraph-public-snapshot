import * as classNames from "classnames";
import * as React from "react";
import { Code, FlexContainer, Heading, Panel, Table } from "sourcegraph/components";
import * as base from "sourcegraph/components/styles/_base.css";
import { whitespace } from "sourcegraph/components/utils";

export class FlexContainerComponent extends React.Component<{}, any> {

	render(): JSX.Element | null {
		return (
			<div className={base.mv4}>
				<Heading level={3} className={base.mb3}>FlexContainer</Heading>

				<Panel hoverLevel="low">
					<div className={base.pa4}>
						<Heading level={7} className={base.mb3} color="cool_mid_gray">Default</Heading>
						<FlexContainer>
							<div className={classNames(base.ba, base.pa2)}>
								42.
							</div>
							<div className={classNames(base.ba, base.pa2)}>
								For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time.
							</div>
						</FlexContainer>

						<Heading level={7} className={base.mv3} color="cool_mid_gray">Wrap</Heading>
						<FlexContainer wrap={true}>
							<div className={classNames(base.ba, base.pa2)}>
								Man had always assumed that he was more intelligent than dolphins.
							</div>
							<div className={classNames(base.ba, base.pa2)}>
								Man had always assumed that he was more intelligent than dolphins.
							</div>
							<div className={classNames(base.ba, base.pa2)}>
								Man had always assumed that he was more intelligent than dolphins.
							</div>
							<div className={classNames(base.ba, base.pa2)}>
								Man had always assumed that he was more intelligent than dolphins.
							</div>
						</FlexContainer>

						<Heading level={7} className={base.mv3} color="cool_mid_gray">Space Between</Heading>
						<FlexContainer justify="between">
							<div className={classNames(base.ba, base.pa2)}>
								42.
							</div>
							<div className={classNames(base.ba, base.pa2)}>
								42.
							</div>
							<div className={classNames(base.ba, base.pa2)}>
								42.
							</div>
							<div className={classNames(base.ba, base.pa2)}>
								42.
							</div>
						</FlexContainer>

						<Heading level={7} className={base.mv3} color="cool_mid_gray">Space Around</Heading>
						<FlexContainer justify="around">
							<div className={classNames(base.ba, base.pa2)}>
								42.
							</div>
							<div className={classNames(base.ba, base.pa2)}>
								42.
							</div>
							<div className={classNames(base.ba, base.pa2)}>
								42.
							</div>
							<div className={classNames(base.ba, base.pa2)}>
								42.
							</div>
						</FlexContainer>

						<p className={base.mt5}>
							<a href="https://css-tricks.com/snippets/css/a-guide-to-flexbox/">Read more about how to use flexbox.</a>
						</p>

					</div>
					<hr />
					<code>
						<pre className={base.ph4} style={{ whiteSpace: "pre-wrap" }}>
							{
								`
<FlexContainer>
	<div>42.</div>
	<div>
		For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time.
	</div>
</FlexContainer>
<FlexContainer wrap={true}>
	<div>Man had always assumed that he was more intelligent than dolphins.</div>
	<div>Man had always assumed that he was more intelligent than dolphins.</div>
	<div>Man had always assumed that he was more intelligent than dolphins.</div>
	<div>Man had always assumed that he was more intelligent than dolphins.</div>
</FlexContainer>
<FlexContainer justify="between">
	<div>42.</div>
	<div>42.</div>
	<div>42.</div>
	<div>42.</div>
</FlexContainer>
<FlexContainer justify="around">
	<div>42.</div>
	<div>42.</div>
	<div>42.</div>
	<div>42.</div>
</FlexContainer>

`
							}
						</pre>
					</code>
				</Panel>
				<Heading level={6} style={{ marginTop: whitespace[4], marginBottom: whitespace[3] }}>Properties</Heading>
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
								<td><Code>direction</Code></td>
								<td><Code>left_right</Code></td>
								<td>
									<Code>left_right</Code>, <Code>right_left</Code>, <Code>top_bottom</Code>, <Code>bottom_top</Code>
								</td>
							</tr>
							<tr>
								<td><Code>wrap</Code></td>
								<td><Code>false</Code></td>
								<td>
									<Code>true</Code>, <Code>false</Code>
								</td>
							</tr>
							<tr>
								<td><Code>justify</Code></td>
								<td><Code>start</Code></td>
								<td>
									<Code>start</Code>, <Code>end</Code>, <Code>center</Code>, <Code>between</Code>, <Code>around</Code>
								</td>
							</tr>
							<tr>
								<td><Code>items</Code></td>
								<td><Code>stretch</Code></td>
								<td>
									<Code>start</Code>, <Code>end</Code>, <Code>center</Code>, <Code>baseline</Code>, <Code>stretch</Code>
								</td>
							</tr>
							<tr>
								<td><Code>content</Code></td>
								<td><Code>stretch</Code></td>
								<td>
									<Code>start</Code>, <Code>end</Code>, <Code>center</Code>, <Code>between</Code>, <Code>around</Code>, <Code>stretch</Code>
								</td>
							</tr>
						</tbody>
					</Table>
				</Panel>
			</div>
		);
	}
}
