// @flow

import * as React from "react";
import CSSModules from "react-css-modules";
import base from "sourcegraph/components/styles/_base.css";
import {Heading, Panel, Table, Code, FlexContainer} from "sourcegraph/components";

class FlexContainerComponent extends React.Component {

	render() {
		return (
			<div className={base.mv4}>
				<Heading level="3" className={base.mb3}>FlexContainer</Heading>

				<Panel hoverLevel="low">
					<div className={base.pa4}>
						<Heading level="7" className={base.mb3} color="cool-mid-gray">Default</Heading>
						<FlexContainer>
							<div className={`${base.ba} ${base.pa2}`}>
								42.
							</div>
							<div className={`${base.ba} ${base.pa2}`}>
								For instance, on the planet Earth, man had always assumed that he was more intelligent than dolphins because he had achieved so much—the wheel, New York, wars and so on—whilst all the dolphins had ever done was muck about in the water having a good time.
							</div>
						</FlexContainer>

						<Heading level="7" className={base.mv3} color="cool-mid-gray">Wrap</Heading>
						<FlexContainer wrap={true}>
							<div className={`${base.ba} ${base.pa2}`}>
								Man had always assumed that he was more intelligent than dolphins.
							</div>
							<div className={`${base.ba} ${base.pa2}`}>
								Man had always assumed that he was more intelligent than dolphins.
							</div>
							<div className={`${base.ba} ${base.pa2}`}>
								Man had always assumed that he was more intelligent than dolphins.
							</div>
							<div className={`${base.ba} ${base.pa2}`}>
							Man had always assumed that he was more intelligent than dolphins.
							</div>
						</FlexContainer>

						<Heading level="7" className={base.mv3} color="cool-mid-gray">Space Between</Heading>
						<FlexContainer justify="between">
							<div className={`${base.ba} ${base.pa2}`}>
								42.
							</div>
							<div className={`${base.ba} ${base.pa2}`}>
								42.
							</div>
							<div className={`${base.ba} ${base.pa2}`}>
								42.
							</div>
							<div className={`${base.ba} ${base.pa2}`}>
								42.
							</div>
						</FlexContainer>

						<Heading level="7" className={base.mv3} color="cool-mid-gray">Space Around</Heading>
						<FlexContainer justify="around">
							<div className={`${base.ba} ${base.pa2}`}>
								42.
							</div>
							<div className={`${base.ba} ${base.pa2}`}>
								42.
							</div>
							<div className={`${base.ba} ${base.pa2}`}>
								42.
							</div>
							<div className={`${base.ba} ${base.pa2}`}>
								42.
							</div>
						</FlexContainer>

						<p className={base.mt5}>
							<a href="https://css-tricks.com/snippets/css/a-guide-to-flexbox/">Read more about how to use flexbox.</a>
						</p>

					</div>
					<hr />
					<code>
						<pre className={base.ph4} style={{whiteSpace: "pre-wrap"}}>
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
								<td><Code>direction</Code></td>
								<td><Code>left-right</Code></td>
								<td>
									<Code>left-right</Code>, <Code>right-left</Code>, <Code>top-bottom</Code>, <Code>bottom-top</Code>
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

export default CSSModules(FlexContainerComponent, base, {allowMultiple: true});
