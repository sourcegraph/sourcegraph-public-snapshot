import * as React from "react";

import { context } from "sourcegraph/app/context";
import { LocationProps } from "sourcegraph/app/router";
import { FlexContainer, Heading, Hero } from "sourcegraph/components";

import { PageTitle } from "sourcegraph/components/PageTitle";
import { colors, whitespace } from "sourcegraph/components/utils";

import { ZapBetaInstructions } from "sourcegraph/page/ZapBetaInstructions";

export function ZapBetaPage({ location }: LocationProps): JSX.Element {
	const zigSx = {
		background: `linear-gradient(-160deg, #FCFCFD 16px, transparent 0) 0 16px, linear-gradient(160deg,  #FCFCFD 16px, ${colors.blueGrayL2()} 0) 0 16px`,
		backgroundColor: "white",
		backgroundPosition: "left top",
		backgroundRepeat: "repeat-x",
		backgroundSize: "64px 32px",
		height: 32,
		margin: whitespace[0],
	};

	return (
		<div>
			<PageTitle title="Zap beta" />

			<Hero style={{
				background: `linear-gradient( ${colors.blueD1()}, ${colors.blueL1()} )`,
			}}>
				<FlexContainer direction="top-bottom" style={{ margin: "auto", paddingRight: whitespace[3], paddingLeft: whitespace[3], textAlign: "center" }}>
					<Heading level={1} color="white">Real-time code collaboration + intelligence</Heading>
					<Heading level={4} color="white" style={{ marginBottom: whitespace[1] }}>Sourcegraph extends your editor to the web so you can share work-in-progress code instantly with teammates.</Heading>
					<div>
						<iframe style={{
							position: "relative",
							top: 20,
							bottom: -80,
							marginBottom: -80,
							maxWidth: 640,
						}} src={"https://player.vimeo.com/video/209506088" width="640" height="400"></iframe>
					</div>
				</FlexContainer>
			</Hero>

			<Hero color="dark" style={{ paddingTop: whitespace[8], paddingBottom: whitespace[0] }}>
				<FlexContainer style={{ margin: "auto", maxWidth: 640, paddingTop: whitespace[1], paddingRight: whitespace[3], paddingLeft: whitespace[3] }}>
					<Heading level={4} align="center" style={{
						marginTop: whitespace[4],
						marginBottom: whitespace[0],
						color: "white",
					}}>
						This beta feature is currently available for Visual Studio Code on Mac OS X. See below for instructions.
					</Heading>
				</FlexContainer>

				<img style={{
					position: "relative",
					bottom: -64,
					maxWidth: 640,
				}} src={`${context.assetsRoot}/img/zap-lp-illus-1.png`} />
			</Hero>

			<ZapBetaInstructions />

			<div style={zigSx}></div>

			<div style={{
				textAlign: "center",
				backgroundColor: colors.blueGrayL2(),
				paddingBottom: whitespace[3],
			}}>
				<FlexContainer direction="top-bottom" style={{ margin: "auto", maxWidth: 640, paddingRight: 16, paddingLeft: 16 }}>
					<Heading level={1} style={{
						marginTop: whitespace[3],
						marginBottom: whitespace[2],
					}}>
						“Whoa, this is amazing.”
						</Heading>
					<p>Zachary at Raise</p>
				</FlexContainer>
			</div>
		</div>
	);
};
