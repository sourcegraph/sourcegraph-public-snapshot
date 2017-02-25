import { $ as glamorSelector, css, focus, lastChild } from "glamor";
import * as React from "react";
import { EventListener, isNonMonacoTextArea } from "sourcegraph/Component";
import { Key } from "sourcegraph/components/Key";
import { ModalComp } from "sourcegraph/components/Modal";
import { Close } from "sourcegraph/components/symbols/Primaries";
import { black, blue, blueGrayD1, blueGrayL1, blueGrayL3, white } from "sourcegraph/components/utils/colors";
import { weight } from "sourcegraph/components/utils/typography";
import { Events } from "sourcegraph/tracking/constants/AnalyticsConstants";

export type Props = {
	onDismiss: () => void,
	showModal: boolean,
	activateShortcut: () => void,
};

const modalStyle = {
	backgroundColor: white(),
	margin: "0px auto",
	marginTop: "4.75rem",
	maxWidth: "19rem",
	borderRadius: 3,
	boxShadow: `0 5px 8px ${black(.5)}`,
	color: blueGrayD1(),
};

const shortcutStyle = css(
	lastChild({ marginBottom: 0 }),
	{
		marginBottom: "1rem",
		display: "block"
	},
);

const titleStyle = {
	borderStyle: "solid",
	borderColor: blueGrayL3(),
	borderWidth: "0px 0px 1px 0px",
	fontWeight: weight[2] as number,
	paddingLeft: "1.5rem",
	paddingBottom: "1rem",
};

const xLinkStyle = css(
	glamorSelector(":hover .inner", { color: blue() }),
	{ float: "right", height: 56, width: 56, color: blueGrayL1() });

export class ShortcutModalComponent extends React.Component<Props, {}> {

	searchModalShortcuts = (event: KeyboardEvent & Event): void => {
		if ((event.target as any).nodeName === "INPUT" || isNonMonacoTextArea((event.target as Node))) {
			return;
		}
		const SlashKeyCode = 191;
		if (event.shiftKey && (event.key === "/" || event.keyCode === SlashKeyCode)) {
			if (this.props.showModal) {
				this.dismissModal();
			} else {
				this.props.activateShortcut();
			}
			event.preventDefault();
		}
	}

	dismissModal = (): void => {
		Events.ShortcutMenu_Dismissed.logEvent();
		this.props.onDismiss();
	}

	componentDidUpdate(prevProps: Props, prevState: {}): void {
		const inputField = document.getElementById("inputFieldHackShortcutMenu");
		if (inputField) {
			inputField.focus();
		}
	}

	render(): JSX.Element {
		return <div>
			{this.props.showModal && <ModalComp onDismiss={() => this.dismissModal()}>
				<div style={modalStyle}>
					<div style={titleStyle}>
						<span style={{ display: "inline-block", marginTop: "1rem" }}>Keyboard shortcuts</span>
						<InputFieldForAnts />
						<a onClick={this.dismissModal} {...xLinkStyle}>
							<Close className="inner" width={24} style={{ display: "block", margin: "auto", marginTop: "-8px", top: "50%" }} />
						</a>
					</div>
					<div style={{ padding: "1.5rem" }}>
						{shortcutElement("?", "Show keyboard shortcuts")}
						{shortcutElement("/", "Open quick search")}
						{shortcutElement("a", "Toggle authors")}
						{shortcutElement("g", "View on GitHub")}
						{shortcutElement("y", "Show commit hash in URL bar")}
					</div>
				</div>
			</ModalComp>}
			<EventListener target={global.document.body} event="keydown" callback={this.searchModalShortcuts} />
		</div>;
	}
}

function shortcutElement(letter: string, text: string): JSX.Element {
	return <div {...shortcutStyle}><Key shortcut={letter} /><span style={{ marginLeft: 12 }}>{text}</span></div>;
}

const miniInputField = css(
	focus({ outline: "none" }),
	{
		height: 0,
		width: 0,
		border: 0,
		borderStyle: "none",
	}
);

function InputFieldForAnts(): JSX.Element {
	return <input {...miniInputField} id="inputFieldHackShortcutMenu" />;
}
