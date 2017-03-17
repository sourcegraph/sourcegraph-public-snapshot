import * as React from "react";

import { context } from "sourcegraph/app/context";
import { LocationProps } from "sourcegraph/app/router";
import { FlexContainer, Heading, Hero } from "sourcegraph/components";

import { PageTitle } from "sourcegraph/components/PageTitle";
import { colors, whitespace } from "sourcegraph/components/utils";

export function ZapBetaPage({ location }: LocationProps): JSX.Element {
	const codeSx = {
		display: "block",
		backgroundColor: `${colors.blueGrayD2()}`,
		color: `${colors.greenL1()}`,
		padding: "12px 16px",
		borderRadius: 4,
		marginBottom: whitespace[5],
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
					<Heading level={4} color="white" style={{ marginBottom: whitespace[5] }}>Sourcegraph extends your editor to the web so you can share your code instantly with your team.</Heading>

					<video poster={`${context.assetsRoot}/img/zap-vid-placeholder.png`} style={{
						width: "96%",
						maxWidth: 640,
						backgroundColor: colors.blueGrayL3(),
						borderRadius: 4,
						margin: "auto",
						marginBottom: -80,
						objectFit: "cover",
					}} controls>
						<source src={`${context.assetsRoot}/zap-beta-demo.mp4`} type="video/mp4" />
					</video>
				</FlexContainer>
			</Hero>

			<Hero color="dark" style={{ paddingTop: whitespace[5], paddingBottom: whitespace[0] }}>
				<FlexContainer style={{ margin: "auto", maxWidth: 640, paddingTop: whitespace[5], paddingRight: whitespace[3], paddingLeft: whitespace[3], }}>
					<Heading level={2} align="center" style={{
						marginTop: whitespace[6],
						marginBottom: whitespace[0],
						color: "white",
					}}>
						Collaborate in real-time with teammates, contributors, or customers.
					</Heading>
				</FlexContainer>

				<img style={{
					position: "relative",
					bottom: -64,
					maxWidth: 640,
				}} src={`${context.assetsRoot}/img/zap-lp-illus-1.png`} />
			</Hero>

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
					<code style={codeSx}>{`sh <(curl -sSf httcodes://sourcegraph.com/install/zap)`}</code>
				</div>

				<div style={{
					borderTop: `1px solid ${colors.blueGrayL3()}`,
					marginTop: whitespace[3],
				}}>
					<div style={stepSx}>2</div>
					<p>To authenticate Zap against your Sourcegraph account, type zap auth into the terminal and follow the prompts.</p>
					<code style={codeSx}>zap auth</code>
				</div>

				<div style={{
					borderTop: `1px solid ${colors.blueGrayL3()}`,
					marginTop: whitespace[3],
				}}>
					<div style={stepSx}>3</div>
					<p>Before initializing Zap on your repo in the next step, visit the repo you'd like to use on Sourcegraph:</p>
					<a href="https://sourcegraph.com/" target="_blank" style={{ wordWrap: "break-word" }}>https://sourcegraph.com/github.com/your_org/your_repo</a>
					<p>This repo should also be cloned and avaliable for editing in your local environment.</p>
				</div>

				<div style={{
					borderTop: `1px solid ${colors.blueGrayL3()}`,
					marginTop: whitespace[3],
				}}>
					<div style={stepSx}>4</div>
					<p>To install and run the Zap server, open a terminal and paste:</p>
					<code style={codeSx}>zap init</code>
					<code style={codeSx}>zap remote set origin wss://sourcegraph.com/.api/zap github.com/your_org/your_repo</code>
					<code style={codeSx}>zap checkout -upstream origin -overwrite -create your_branch@your_unix_user_name</code>
				</div>

				<div style={{
					borderTop: `1px solid ${colors.blueGrayL3()}`,
					marginTop: whitespace[3],
				}}>
					<div style={stepSx}>5</div>
					<p>Open Visual Studio Code in your repo's directory.</p>
					<code style={codeSx}>code.</code>
				</div>

				<div style={{
					borderTop: `1px solid ${colors.blueGrayL3()}`,
					marginTop: whitespace[3],
				}}>
					<div style={stepSx}>6</div>
					<p>To install the Visual Studio Code extension, open the command palette (Command+P) and type:</p>
					<code style={codeSx}>ext install sqs.vscode-zap</code>
					<p>Click the “Reload” button to reload Visual Studio Code.</p>
				</div>

				<div style={{
					borderTop: `1px solid ${colors.blueGrayL3()}`,
					borderBottom: `1px solid ${colors.blueGrayL3()}`,
					marginTop: whitespace[3],
					paddingBottom: whitespace[3],
					textAlign: "center",
				}}>
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
					<p>Zachary at <a href="http://raise.com/" target="_blank">Raise.com</a></p>
				</FlexContainer>
			</div>
		</div>
	);
};
