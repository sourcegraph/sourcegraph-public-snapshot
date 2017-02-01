import { css, focus, lastChild } from "glamor";
import * as React from "react";
import { EventListener, isNonMonacoTextArea } from "sourcegraph/Component";
import { Key } from "sourcegraph/components/Key";
import { ModalComp } from "sourcegraph/components/Modal";
import { Close } from "sourcegraph/components/symbols/Primaries";
import { blueGrayD1, white } from "sourcegraph/components/utils/colors";
import { weight } from "sourcegraph/components/utils/typography";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";

export type Props = {
	onDismiss: () => void,
	showModal: boolean,
	activateShortcut: () => void,
};

const modalStyle = {
	backgroundColor: white(),
	margin: "0px auto",
	marginTop: "4.75rem",
	maxWidth: "18rem",
	borderRadius: 3,
	boxShadow: "0  5px 8px rgba(0,0,0,.5)",
	color: blueGrayD1(),
};

const shortcutStyle = css(
	lastChild({ marginBottom: 0 }),
	{
		marginBottom: 20,
		display: "block"
	},
);

const titleStyle = {
	borderStyle: "solid",
	borderColor: "#f2f4f8",
	borderWidth: "0px 0px 1px 0px",
	fontWeight: weight[2] as number,
	paddingLeft: "1.5rem",
	paddingBottom: "1rem",
};

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
				setTimeout(() => {
					const inputField = document.getElementById("inputFieldHackShortcutMenu");
					if (inputField) {
						inputField.focus();
					}
				}, 1);
			}
			event.preventDefault();
		}
	}

	dismissModal = (): void => {
		AnalyticsConstants.Events.ShortcutMenu_Dismissed.logEvent();
		this.props.onDismiss();
	}

	render(): JSX.Element {
		return <div>
			{this.props.showModal && <ModalComp onDismiss={() => this.dismissModal()}>
				<div style={modalStyle}>
					<div style={titleStyle}>
						<span style={{ display: "inline-block", marginTop: "1.5rem" }}>Keyboard shortcuts</span>
						<InputFieldForAnts />
						<a onClick={this.dismissModal} style={{ float: "right", height: 56, verticalAlign: "middle", width: 56 }}>
							<Close width={24} color="#93a9c8" style={{ display: "block", margin: "auto", marginTop: "-6px", top: "50%" }} />
						</a>
					</div>
					<div style={{ padding: "1.5rem" }}>
						{shortcutElement("/", "Open quick search")}
						{shortcutElement("g", "View on GitHub")}
						{shortcutElement("a", "Toggle authors")}
						{shortcutElement("u", "Show full path in URL bar")}
						{shortcutElement("?", "Show keyboard shortcuts")}
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
