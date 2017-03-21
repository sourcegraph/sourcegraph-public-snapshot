import * as React from "react";

import { context } from "sourcegraph/app/context";
import { FlexContainer, Heading, Panel } from "sourcegraph/components";

import { colors, whitespace } from "sourcegraph/components/utils";
import { SignupForm } from "sourcegraph/user/Signup";

export function ZapBetaInstructions(): JSX.Element {

	const codeSx = {
		display: "block",
		backgroundColor: `${colors.blueGrayD2()}`,
		color: `${colors.greenL1()}`,
		padding: "12px 16px",
		borderRadius: 4,
		marginBottom: whitespace[4],
	};
	const stepSx = {
		display: "block",
		backgroundColor: `${colors.blueGrayL3()}`,
		color: `${colors.blueGrayL1()}`,
		borderRadius: whitespace[3],
		width: 24,
		height: 24,
		textAlign: "center",
		position: "relative",
		top: "-12px",
		lineHeight: "23px",
		margin: "auto",
	};
	const keySx = {
		display: "inline",
		color: `${colors.blueGray()}`,
		backgroundColor: `${colors.blueGrayL3()}`,
		border: `1px solid ${colors.blueGrayL2()}`,
		boxShadow: "rgba(201, 212, 227, 1) 0px 4px 0px 0px",
		borderRadius: 4,
		padding: "12.5px 24px",
		margin: whitespace[0],
	};

	if (!context.user) {
		return <Panel hoverLevel="low" style={{ paddingTop: whitespace[9], textAlign: "center" }}>
			<p>You must sign up to view installation instructions.</p>
			<SignupForm />
		</Panel>;
	}

	return (
		<div>

			<FlexContainer direction="top-bottom" style={{
				maxWidth: 640,
				margin: "auto",
				paddingTop: whitespace[5],
				paddingRight: whitespace[3],
				paddingLeft: whitespace[3],
			}}>
				<Heading level={2} align="center" style={{
					marginTop: whitespace[5],
					marginBottom: whitespace[5],
					color: colors.blueGrayD1(),
				}}>
					Instructions
				</Heading>

				<div style={{
					borderTop: `1px solid ${colors.blueGrayL3()}`,
				}}>
					<div style={stepSx}>1</div>
					<p>To install and run the Zap server, open a terminal and paste:</p>
					<code style={codeSx}>{`sh <(curl -sSf https://sourcegraph.com/install/zap)`}</code>
				</div>

				<div style={{ borderTop: `1px solid ${colors.blueGrayL3()}`, marginTop: whitespace[4] }}>
					<div style={stepSx}>2</div>
					<p>Before initializing Zap on your repo in the next step, visit the repo you'd like to use on Sourcegraph:</p>
					<a href="https://sourcegraph.com/" target="_blank" style={{ wordWrap: "break-word" }}>https://sourcegraph.com/github.com/<strong>org</strong>/<strong>repo</strong></a>
					<p>This repo should also be cloned and avaliable for editing in your local environment.</p>
				</div>

				<div style={{ borderTop: `1px solid ${colors.blueGrayL3()}`, marginTop: whitespace[4], }}>
					<div style={stepSx}>3</div>
					<p>To start watching a repo, go to the repo directory you want to watch, and type:</p>
					<code style={codeSx}>zap auth</code>
					<code style={codeSx}>zap init</code>
				</div>

				<div style={{ borderTop: `1px solid ${colors.blueGrayL3()}`, marginTop: whitespace[4] }}>
					<div style={stepSx}>4</div>
					<p>Open Visual Studio Code in your repo's directory:</p>
					<code style={codeSx}>code.</code>
				</div>

				<div style={{ borderTop: `1px solid ${colors.blueGrayL3()}`, marginTop: whitespace[4] }}>
					<div style={stepSx}>5</div>
					<p>To install the Visual Studio Code extension, open the command palette (Command+P) and type:</p>
					<code style={codeSx}>ext install sqs.vscode-zap</code>
					<p>Click the “Reload” button to reload Visual Studio Code.</p>
				</div>

				<div style={{ borderTop: `1px solid ${colors.blueGrayL3()}`, marginTop: whitespace[4] }}>
					<div style={stepSx}>6</div>
					<p>If the Zap button in the status bar is off (see: <a href="https://cl.ly/363k0D423g2D" target="_blank">https://cl.ly/363k0D423g2D</a>), click it to turn Zap on (see: <a href="https://cl.ly/0H2a0y3k2d2d" target="_blank">https://cl.ly/0H2a0y3k2d2d)</a>.</p>
				</div>

				<div style={{ borderTop: `1px solid ${colors.blueGrayL3()}`, borderBottom: `1px solid ${colors.blueGrayL3()}`, marginTop: whitespace[4], paddingBottom: whitespace[3], textAlign: "center" }}>
					<div style={stepSx}>7</div>
					<p>To jump to Sourcegraph from Visual Studio Code, use the keyboard shortcut:</p>
					<div style={{
						display: "block",
						margin: `${whitespace[5]} 0`,
					}}>
						<Heading level={4} style={keySx}>Option</Heading>
						<span style={{
							fontSize: 24,
							lineHeight: whitespace[3],
							fontWeight: 800,
							margin: `0 ${whitespace[3]}`,
							color: colors.blueGray(),
						}}>+</span>
						<Heading level={4} style={keySx}>S</Heading>
					</div>
					<p>or right-click anywhere in the file and select “Open in Web Browser.”</p>
				</div>

				<div style={{
					margin: `${whitespace[3]} 0`,
					textAlign: "center",
				}}>
					<p>
						Running into problems?&nbsp;
						<a href="https://slack-files.com/T02FSM7DL-F4BRWRCDC-00bd5b24eb" target="_blank">Check our troubleshooting</a>.
					</p>
				</div>

			</FlexContainer>

		</div>
	);
};
