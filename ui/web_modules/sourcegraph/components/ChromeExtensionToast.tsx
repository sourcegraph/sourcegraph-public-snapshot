import * as classNames from "classnames";
import * as React from "react";
import { RouterLocation } from "sourcegraph/app/router";
import * as base from "sourcegraph/components/styles/_base.css";
import { Close } from "sourcegraph/components/symbols/Primaries";
import { Toast } from "sourcegraph/components/Toast";
import { installChromeExtensionClicked } from "sourcegraph/util/ChromeExtensionInstallHandler";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import { EventLogger } from "sourcegraph/util/EventLogger";
import { shouldPromptToInstallBrowserExtension } from "sourcegraph/util/shouldPromptToInstallBrowserExtension";

const TOAST_TITLE = "Save time browsing code on GitHub with the Sourcegraph browser extension!";
const EXTENSION_TOAST_KEY = "chrome-extension-toast-dismissed";

interface State {
	isVisible: boolean;
}

interface Props {
	location: RouterLocation;
	layout: () => void;
}

export class ChromeExtensionToast extends React.Component<Props, State>  {
	constructor() {
		super();
		this.state = {
			isVisible: false,
		};
	}

	componentDidMount(): void {
		const isVisible = shouldPromptToInstallBrowserExtension() && !Boolean(window.localStorage[EXTENSION_TOAST_KEY]);
		if (isVisible) {
			EventLogger.logViewEvent("ViewChromeExtensionToast", this.props.location.pathname, { toastCopy: TOAST_TITLE });
		}
		this.setState({
			isVisible: isVisible,
		});
	}

	render(): JSX.Element | null {
		let { isVisible } = this.state;
		if (isVisible) {
			return (
				<Toast>
					<a onClick={this.closeClicked.bind(this)} className={classNames(base.fr, base.mt2)}><Close /></a>
					<p style={{ textAlign: "center" }}><a onClick={this.toastCTAClicked.bind(this)}>{TOAST_TITLE}</a></p>
				</Toast>
			);
		}

		return null;
	}

	private toastCTAClicked(): void {
		installChromeExtensionClicked("ChromeExtensionOnboarding");
		this.dismissToast();
	}

	private closeClicked(): void {
		AnalyticsConstants.Events.ToastChrome_Closed.logEvent({ pageName: "ChromeExtensionOnboarding" });
		this.dismissToast();
	}

	private dismissToast(): void {
		window.localStorage[EXTENSION_TOAST_KEY] = "true";
		this.setState({ isVisible: false });

		// Call layout to let workbench draw itself according to the new
		// dimensions.
		this.props.layout();
	}
}
