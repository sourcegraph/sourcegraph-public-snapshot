import { hover, style } from "glamor";
import * as React from "react";
import { Link } from "react-router";
import { colors, whitespace } from "sourcegraph/components/utils";

export function Footer(): JSX.Element {
	return <div style={{ backgroundColor: colors.coolGray4(0.2) }}>
		<div style={{
			padding: whitespace[3],
			textAlign: "center",
		}}>
			<Item url="/about">About</Item>
			<Item url="/plan">Master plan</Item>
			<Item url="https://text.sourcegraph.com" anchor>Blog</Item>
			<Item url="/docs">Docs</Item>
			<Item url="/jobs" anchor>Careers</Item>
			<Item url="/contact">Contact</Item>
			<Item url="/pricing">Pricing</Item>
			<Item url="/privacy">Privacy</Item>
			<Item url="/security">Security</Item>
			<Item url="/sitemap" anchor>Sitemap</Item>
			<Item url="/terms">Terms</Item>
		</div>
	</div>;
};

interface ItemProps {
	children?: React.ReactNode[];
	url: string;
	anchor?: boolean;
}

function Item({anchor, url, children}: ItemProps): JSX.Element {
	const sx = style({
		color: colors.coolGray3(),
		display: "inline-block",
		paddingLeft: whitespace[3],
		paddingRight: whitespace[3],
		paddingTop: whitespace[2],
		paddingBottom: whitespace[2],
	});

	const hoverSx = hover({ color: colors.coolGray2() });

	if (anchor) {
		return <a href={url} {...sx} {...hoverSx}>{children}</a>;
	}
	return <Link to={url} {...sx} {...hoverSx}>{children}</Link>;
};
