import { $ as glamorSelector, css, lastChild } from "glamor";
import * as React from "react";
import { Link } from "react-router";
import { handleIntercomToggle } from "sourcegraph/app/GlobalNav/UserMenu";
import { BookClosed } from "sourcegraph/components/symbols/Primaries";
import { colors } from "sourcegraph/components/utils";
import { whitespace } from "sourcegraph/components/utils/whitespace";
import { Events, PAGE_LOCATION_DASHBOARD_SIDEBAR } from "sourcegraph/tracking/constants/AnalyticsConstants";

const listItemStyle = css(
	glamorSelector(":hover .inner", { color: colors.white() }),
	glamorSelector(":hover svg", { fill: colors.white() }),
	lastChild({ marginBottom: 0 }),
	{
		marginBottom: "1.5rem",
		display: "block"
	},
);

const linkStyle = css({
	color: colors.blueGrayL2(),
	display: "inline-block",
});

const svgStyle = {
	display: "inline",
	marginRight: whitespace[3],
	color: colors.blueGrayL2(),
};

export function HelpBar(): JSX.Element {
	// const createSvg = <Compose style={svgStyle} width={24} />;
	const bookClosed = <BookClosed style={svgStyle} width={24} />;
	// const videoCamera = <VideoCamera style={svgStyle} width={24} />;
	return <ul style={{ margin: "1.5rem", marginRight: 0, listStyle: "none", padding: 0 }}>
		<HelpBarListItem text="Documentation" href="/docs" onClick={documentationClicked} svg={bookClosed} />
		{/*<HelpBarListItem text="Videos" href="/videos" onClick={videosClicked} svg={videoCamera} />
		<HelpBarListItem text="Contact" onClick={contactClicked} svg={createSvg} />*/}
	</ul>;
}

interface ListItemProps {
	text: string;
	svg: JSX.Element;
	href?: string;
	onClick?: (event: Object) => void;
}

function HelpBarListItem(props: ListItemProps): JSX.Element {
	// the cast to string below is unavaoidable; passing in null or undefined does not
	// break the <Link> class.
	return <li {...listItemStyle}>
		<Link className="inner" {...linkStyle} to={props.href as string} onClick={props.onClick}>
			{props.svg}
			{props.text}
		</Link>
	</li>;
}

const documentationClicked = (event: React.MouseEvent<HTMLAnchorElement>) => {
	Events.ReposPageDocsButton_Clicked.logEvent();
};

const videosClicked = (event: React.MouseEvent<HTMLAnchorElement>) => {
	Events.ReposPageVideoButton_Clicked.logEvent();
};

const contactClicked = (event: React.MouseEvent<HTMLAnchorElement>) => {
	Events.ReposPageContactButton_Clicked.logEvent();
	handleIntercomToggle(PAGE_LOCATION_DASHBOARD_SIDEBAR);
};
