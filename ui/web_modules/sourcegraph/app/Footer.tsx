import { hover, media, style } from "glamor";
import * as React from "react";
import { Link } from "react-router";

import { context } from "sourcegraph/app/context";
import { FlexContainer } from "sourcegraph/components";
import { GitHub, LinkedIn, Twitter } from "sourcegraph/components/symbols/Material";
import { colors, layout, whitespace } from "sourcegraph/components/utils";

export function Footer(): JSX.Element {
	return <div style={{ backgroundColor: "#f0f5fb" }}>
		<FlexContainer wrap={true} style={{
			padding: whitespace[3],
			paddingBottom: whitespace[5],
			maxWidth: 960,
			margin: "auto",
		}}>
			<Item url="/plan">Master plan</Item>
			<Item url="/docs">Docs</Item>
			<Item url="https://text.sourcegraph.com" anchor>Blog</Item>
			<Item url="/pricing">Pricing</Item>
			<Item url="/terms">Terms</Item>
			<Item url="/security">Security</Item>
			<Item url="/privacy">Privacy</Item>
			<Item target="_blank" url="/beta" anchor>Beta program</Item>
			<Item target="_blank" url="/jobs" anchor>Careers</Item>
			<Item url="/about">About</Item>
			<Item url="/contact">Contact</Item>
		</FlexContainer>

		<FlexContainer justify="between" style={{
			padding: whitespace[3],
			maxWidth: 960,
			margin: "auto",
		}}>
			<FlexContainer>
				<div style={{
					marginRight: whitespace[3],
					color: colors.blueGray(),
				}}>
					&copy; 2017 Sourcegraph, Inc.
				</div>
				<a target="_blank" href="https://github.com/sourcegraph">
					<GitHub width={22} color={colors.blueGray()} style={{
						marginRight: whitespace[3],
					}} />
				</a>
				<a target="_blank" href="https://twitter.com/srcgraph">
					<Twitter width={22} color={colors.blueGray()} style={{
						marginRight: whitespace[3],
					}} />
				</a>
				<a target="_blank" href="https://www.linkedin.com/company/sourcegraph">
					<LinkedIn width={22} color={colors.blueGray()} style={{
						marginRight: whitespace[3],
					}} />
				</a>
			</FlexContainer>

			<a target="_blank" href="https://sourcegraph.com">
				<img src={`${context.assetsRoot}/img/about/foot-logo+tag.svg`} />
			</a>

		</FlexContainer>

	</div>;
};

interface ItemProps {
	children?: React.ReactNode[];
	url: string;
	anchor?: boolean;
	target?: string;
}

function Item({ anchor, url, children, target }: ItemProps): JSX.Element {
	const sx = style({
		flex: "0 0 15%",
		display: "block",
		color: colors.blueGray(),
		paddingTop: whitespace[2],
		paddingBottom: whitespace[2],
	});

	const mediaSx = media(layout.breakpoints.sm, { flex: "0 0 48% !important" });
	const hoverSx = hover({ color: colors.blueGrayD1() });

	return anchor
		? <a target={target} href={url} {...sx} {...hoverSx} {...mediaSx}>{children}</a>
		: <Link to={url} {...sx} {...hoverSx} {...mediaSx}>{children}</Link>;
};
