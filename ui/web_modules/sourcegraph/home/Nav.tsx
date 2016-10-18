import {hover, media, merge} from "glamor";
import * as React from "react";
import {Link} from "react-router";
import {Button, FlexContainer, Logo} from "sourcegraph/components";
import {LocationStateToggleLink} from "sourcegraph/components/LocationStateToggleLink";
import {colors, layout, whitespace} from "sourcegraph/components/utils";
import {Location} from "sourcegraph/Location";

interface Props {
	context: any;
	location: Location;
	style?: React.CSSProperties;
}

const navItemSx = {
	marginTop: whitespace[2],
	marginRight: whitespace[4],
	color: colors.coolGray3(),
};

const navHover = hover({ color: colors.coolGray2()});

export function Nav({context, style, location}: Props): JSX.Element {

	return <FlexContainer justify="between" wrap={true} style={style}>
		<Logo type="logotype" width="195" />

		<div>
			<Link to="/about" {...merge(navItemSx, navHover, {marginLeft: 0})}>About</Link>
			<Link to="/pricing" {...merge(navItemSx, navHover)}>Pricing</Link>
			<a href="/jobs" {...merge(navItemSx, navHover)}>Jobs</a>

			{!(context as any).signedIn &&
				<LocationStateToggleLink
					href="/login"
					modalName="login"
					location={location}
					{...merge(navItemSx, navHover)}
				>Log in</LocationStateToggleLink>
			}

			{!(context as any).signedIn &&
				<LocationStateToggleLink
					href="/join"
					modalName="join"
					location={location}
					{...media(layout.breakpoints["sm"], { display: "none"})}
					style={{
						paddingTop: whitespace[2],
						paddingBottom: whitespace[2],
					}}
				><Button color="orange" size="small" style={{marginTop: whitespace[1]}}>Sign up</Button>
				</LocationStateToggleLink>
			}
		</div>
	</FlexContainer>;
};
