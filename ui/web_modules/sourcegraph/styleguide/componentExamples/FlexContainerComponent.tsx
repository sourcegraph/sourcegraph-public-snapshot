import * as classNames from "classnames";
import * as React from "react";
import { FlexContainer, Heading, Panel } from "sourcegraph/components";
import * as base from "sourcegraph/components/styles/_base.css";

export class FlexContainerComponent extends React.Component<{}, any> {

	render(): JSX.Element | null {
		return (
			<div>
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
			</div>
		);
	}
}
