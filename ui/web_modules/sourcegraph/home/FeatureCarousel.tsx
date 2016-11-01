import {hover, style} from "glamor";
import * as React from "react";
import {FlexContainer, Heading} from "sourcegraph/components";
import {colors, layout, whitespace} from "sourcegraph/components/utils";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

interface Props { assetsURL: string; }
interface State { active: 0 | 1 | 2; }

export class FeatureCarousel extends React.Component<Props, State> {

	constructor(props: Props) {
		super(props);
		this.state = { active: 0 };
	}

	render(): JSX.Element {

		return <FlexContainer wrap={true} style={Object.assign({},
			layout.container,
			{
				marginTop: whitespace[5],
				paddingright: whitespace[4],
				maxWidth: 1080,
			},
		)}>
			<FlexContainer direction="top_bottom" wrap={true} justify="around" style={{
				flex: "1 1 220px",
				minWidth: 175,
				paddingBottom: whitespace[5],
				paddingRight: whitespace[4],
			}}>
				<SliderNavItem
					title="View all references"
					subtitle="Quickly understand forward and backward dependencies"
					active={this.state.active === 0}
					onClick={() => { AnalyticsConstants.Events.HomeCarousel_Clicked.logEvent(); this.setState({active: 0}); }} />
				<SliderNavItem
					title="Peek definition"
					subtitle="Grok what code is doing without losing context"
					active={this.state.active === 1}
					onClick={() => { AnalyticsConstants.Events.HomeCarousel_Clicked.logEvent(); this.setState({active: 1}); }} />
				<SliderNavItem
					title="Search by symbol"
					subtitle="Jump to and search for symbols within a repository"
					active={this.state.active === 2}
					onClick={() => { AnalyticsConstants.Events.HomeCarousel_Clicked.logEvent(); this.setState({active: 2}); }} />
			</FlexContainer>
			<div style={{flex: "1 1 60%", position: "relative", minHeight: 480}}>
				<img src={`${this.props.assetsURL}/img/Homepage/screen-placeholder.png`} width="100%" />
				<SliderPanel assetsURL={this.props.assetsURL} img="screen-refs.png" active={this.state.active === 0} />
				<SliderPanel assetsURL={this.props.assetsURL} img="screen-def.png" active={this.state.active === 1} />
				<SliderPanel assetsURL={this.props.assetsURL} img="screen-search.png" active={this.state.active === 2} />
			</div>
		</FlexContainer>;
	}
}

interface SliderNavItemProps {
	title: string;
	subtitle: string;
	active: boolean;
	onClick: () => void;
}

function SliderNavItem({title, subtitle, active, onClick}: SliderNavItemProps): JSX.Element {
	const sx = style({
		backgroundImage: active ? `linear-gradient(270deg, ${colors.white()}, ${colors.blue5(0.25)} 100%)` : "transparent",
		borderRadius: 3,
		color: colors.coolGray2(0.75),
		marginTop: whitespace[4],
		paddingBottom: whitespace[2],
		paddingLeft: whitespace[4],
		paddingTop: whitespace[3],
	});

	return <a href="#" {...sx} {...hover({ color: colors.coolGray2(1) })} onClick={onClick}>
		<Heading level={5} color="blue">{title}</Heading>
		<p style={{marginTop: 0}}>{subtitle}</p>
	</a>;
};

interface SliderPanelProps {
	img: string;
	assetsURL: string;
	active: boolean;
}

function SliderPanel({img, assetsURL, active}: SliderPanelProps): JSX.Element {
	return <div style={{
		maxHeight: 540,
		opacity: active ? 1 : 0,
		position: "absolute",
		left: 0,
		top: 0,
		right: "-10px",
		transition: "all 0.4s ease-in-out",
	}}>
		<img src={`${assetsURL}/img/Homepage/${img}`} width="100%" style={{maxWidth: "100%", minWidth: 700}} />
	</div>;
};
