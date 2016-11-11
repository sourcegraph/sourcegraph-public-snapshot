import * as classNames from "classnames";
import * as React from "react";
import {context} from "sourcegraph/app/context";
import {CloseIcon} from "sourcegraph/components/Icons";
import {LocationStateToggleLink} from "sourcegraph/components/LocationStateToggleLink";
import * as base from "sourcegraph/components/styles/_base.css";
import {Toast} from "sourcegraph/components/Toast";
import {Location} from "sourcegraph/Location";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";

const ChromeExtensionToastKey = "chrome-extension-toast-dismissed";
const ToastTitle = "Save time browsing code on GitHub with the Sourcegraph browser extension!";

interface State {
	isVisible: boolean;
}

interface Props {
	location: Location;
}

export class ChromeExtensionToast extends React.Component<Props, State>  {
	constructor(props: Props) {
		super(props);
		this.state = {isVisible: !window.localStorage[ChromeExtensionToastKey]};
	}

	componentDidMount(): void {
		if (this.state.isVisible) {
			EventLogger.logViewEvent("ViewChromeExtensionToast", this.props.location.pathname, {toastCopy: ToastTitle});
		}
	}

	render(): JSX.Element | null {
		if (this.state.isVisible && ! context.hasChromeExtensionInstalled()) {
			return (
				<Toast>
					<a onClick={this._closeClicked.bind(this)} className={classNames(base.fr, base.mt2)}><CloseIcon/></a>
					<LocationStateToggleLink href="/join" modalName="chrome" location={this.props.location}>
						<p onClick={this._toastCTAClicked.bind(this)} style={{textAlign: "center"}}>{ToastTitle}</p>
					</LocationStateToggleLink>
				</Toast>
			);
		}

		return null;
	}

	_toastCTAClicked(): void {
		AnalyticsConstants.Events.ToastChromeCTA_Clicked.logEvent({toastCopy: ToastTitle});
		this._dismissToast();
	}

	_closeClicked(): void {
		AnalyticsConstants.Events.ToastChrome_Closed.logEvent({toastCopy: ToastTitle});
		this._dismissToast();
	}

	_dismissToast(): void {
		window.localStorage[ChromeExtensionToastKey] = "true";
		this.setState({isVisible: false});
	}
}
