import { css } from "glamor";
import * as React from "react";
import { FlexContainer, Heading, Label, Panel } from "sourcegraph/components";
import { ChevronRight } from "sourcegraph/components/symbols/Primaries";
import { colors, typography, whitespace } from "sourcegraph/components/utils";
import { MONTHLY_SEAT_COST } from "sourcegraph/user/Paywall";

interface Props {
	select: (plan: PlanType) => () => void;
}

interface TileProps {
	name: string;
	desc: string;
	price: number;
	salePrice?: number;
	label?: string;
	unit: string;
	onClick?: () => void;
}

export type PlanType = "personal" | "organization" | "enterprise";

export function PlanSelector({ select }: Props): JSX.Element {
	return <div style={{ padding: whitespace[4] }}>
		<p style={{ marginBottom: whitespace[4], textAlign: "center", color: colors.blueGrayD1() }}>Choose your plan:</p>
		<PersonalPlan onClick={select("personal")} />
		<OrgPlan onClick={select("organization")} />
		<p style={{ marginTop: whitespace[5], textAlign: "center" }}>
			<a onClick={select("enterprise")}>
				<strong>Host Sourcegraph on premises? Contact us <ChevronRight /></strong>
			</a>
		</p>
	</div>;
};

interface PlanProps { onClick?: () => void; }

export function PersonalPlan({ onClick }: PlanProps): JSX.Element {
	return <PlanTile
		onClick={onClick}
		name="Personal"
		desc="Open source and personal code"
		price={0}
		unit="per month" />;
}

export function OrgPlan({ onClick }: PlanProps): JSX.Element {
	return <PlanTile
		onClick={onClick}
		name="Organization"
		desc="One organizaton's private code"
		label="Free 14-day trial"
		price={MONTHLY_SEAT_COST}
		unit="per user/mo" />;
}

export function PublicPlan({ onClick }: PlanProps): JSX.Element {
	return <PlanTile
		onClick={onClick}
		name="Public code"
		desc="Any publicy available code"
		price={0}
		unit="forever" />;
}

function PlanTile({ name, desc, price, salePrice, unit, label, onClick }: TileProps): JSX.Element {

	const unitSx = { ...{ color: colors.blueGrayL1() }, ...typography.small };
	const labelSx = { ...{ marginLeft: whitespace[2] }, ...typography.small };
	const descSx = { color: colors.blueGray() };
	const panelHoverSx = css({
		":hover": {
			border: `1px ${colors.blueGrayL2()} solid !important`,
			color: `${colors.blue()} !important`,
		}
	}).toString();
	const panelSx = {
		border: "1px solid transparent",
		marginBottom: whitespace[3],
		padding: whitespace[3],
		...onClick && { cursor: "pointer" },
	};

	return <Panel onClick={onClick} hoverLevel="low" style={panelSx} className={panelHoverSx}>
		<FlexContainer style={{ marginLeft: whitespace[2], marginRight: whitespace[2] }}>
			<div style={{ flex: "1 1 auto" }}>
				<Heading level={6} compact={true} style={{ marginBottom: whitespace[2] }}>
					{name}
					{label &&
						<Label color="green" text={label} style={labelSx} compact={true} />
					}
				</Heading>
				<span style={descSx}>{desc}</span>
			</div>
			<div style={{ flex: "0 0 80px", textAlign: "center" }}>
				{salePrice && <SalePrice price={9} />}
				<Price price={price} />
				<div style={unitSx}>{unit}</div>
			</div>
		</FlexContainer>
	</Panel>;
}

function Price({ price, style }: { price: number, style?: React.CSSProperties }): JSX.Element {
	return <Heading level={3} compact={true} style={{
		...{
			display: "inline-block",
			lineHeight: "1",
			marginLeft: -8,
		}, ...style
	}}>
		<sup style={{ ...{ color: colors.blueGray() }, ...typography.size[5] }}><em>$</em></sup>
		{price}
	</Heading>;
}

function SalePrice({ price }: { price: number }): JSX.Element {
	const slashSx = css({
		display: "inline-block",
		position: "relative",
		marginRight: "1.5rem",
		":before": {
			backgroundColor: colors.blueGray(),
			borderRadius: 4,
			content: "\"\"",
			transform: "rotate(-20deg)",

			display: "block",
			height: 4,
			width: 40,

			position: "absolute",
			top: "45%",
			left: -8,
			zIndex: 1,
		}
	});
	return <div {...slashSx}>
		<Price price={price} style={{ opacity: 0.25 }} />
	</div>;
}
