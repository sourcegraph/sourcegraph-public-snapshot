import * as React from "react";
import * as ReactDOM from "react-dom";
import {InjectedRouter} from "react-router";
import {context} from "sourcegraph/app/context";
import {Annotation, Button, Heading} from "sourcegraph/components";
import {GitHubAuthButton} from "sourcegraph/components/GitHubAuthButton";
import {Close, Flag} from "sourcegraph/components/symbols/Zondicons";
import {colors, typography, whitespace} from "sourcegraph/components/utils";
import {fontStack} from "sourcegraph/components/utils/typography";
import {Location} from "sourcegraph/Location";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";
import {privateGitHubOAuthScopes} from "sourcegraph/util/urlTo";

interface Props { location: Location; }

interface State {
	visibleMarks: Array<Number>;
	visibleAnnotation: number | null;
}

const _searchCoachmarkIndex: number = 0;
const _refCoachmarkIndex: number = 1;
const _defCoachmarkIndex: number = 2;

interface Coachmark {
	markIndex: number;
	markParentElementId: string;
	markId: string;
	markLineNumber: number;
	headingTitle: string;
	headingSubtitle: JSX.Element | null;
	actionTitle: string;
	actionCTA: JSX.Element | null;

}

const parentElementCssString = `display: inline-block; white-space: normal; cursor: auto; font-family: ${fontStack.sansSerif};`;

const closeSx = {
	cursor: "pointer",
	float: "right",
	paddingRight: whitespace[3],
	paddingTop: whitespace[3],
};

const actionSx = Object.assign({},
	typography.size[6],
	{
		backgroundColor: colors.blue(),
		color: colors.white(),
		display: "inline-block",
		paddingLeft: whitespace[3],
		paddingBottom: whitespace[3],
		width: 240,
	},
);

const headerSx = {
	backgroundColor: colors.white(),
	borderTopLeftRadius: 3,
	borderTopRightRadius: 3,
	paddingTop: whitespace[4],
	paddingLeft: whitespace[4],
	paddingRight: whitespace[4],
	paddingBottom: whitespace[2],
};

const p = Object.assign({},
	typography.size[6],
	{
		width: 270,
		color: colors.text(),
	},
);

const flagSx = {
	marginTop: whitespace[1],
	verticalAlign: "top",
};

const searchSx = {
	backgroundColor: colors.coolGray4(0.8),
	borderRadius: 3,
	padding: whitespace[1],
	height: 21,
	width: 16,
};

export class TourOverlay extends React.Component<Props, State>  {
	static contextTypes: React.ValidationMap<any> = {
		router: React.PropTypes.object.isRequired,
	};

	_coachmarks: Array<Coachmark>;

	context: { router: InjectedRouter };

	constructor() {
		super();
		// During the tour it is possible to redirect the user to GitHub and back to Auth private code and it is possible to revisit the tour by clicking the back button.
		// Therefore we store the current session in the window sessionStore to ensure we do not display the same
		// tooltip twice. To set up the current visibleMarks we assume all are visible. Then we check if the current session
		// has any information about the tour we defer to using the sessionStorage. If the user is a private code user then we will not show the refCoachmarkIndex'd tooltip
		// because the CTA prompts for private code usage. Always null out state.visibleAnnotation in the constructor.
		let visibleMarks = [0, 1, 2];
		let tourStore: string | null = window.sessionStorage.getItem("tour");
		if (tourStore !== null) {
			visibleMarks = tourStore.length > 0 ? tourStore.split(",").map(Number) : [];
		}
		if (context.hasPrivateGitHubToken() && visibleMarks.indexOf(_refCoachmarkIndex) !== -1) {
			visibleMarks.splice(visibleMarks.indexOf(_refCoachmarkIndex), 1);
		}

		this.state = {
			visibleAnnotation: null,
			visibleMarks: visibleMarks,
		};
	}

	componentDidMount(): void {
		// Only try to render the onboarding sequence if the tour contains the appropriate query params.
		if (this.props.location.query["tour"]) {
			this._tryForRenderedTokenIdentifier();
		}
	}

	componentDidUpdate(prevProps: Props, prevState: State): void {
		if (this.state.visibleAnnotation !== prevState.visibleAnnotation && (window as any).ed) {
			this._coachmarksShouldUpdate();
		}
	}

	componentWillReceiveProps(nextProps: Props): void {
		// Safety buffer for shortcircuting the url changes.
		if (nextProps.location !== this.props.location) {
			setTimeout(function(): void {
				// If the location has changed render the coachmarks in a new random location in the view.
				this._tryForRenderedTokenIdentifier();
			}.bind(this), 10);
		}
	}

	// Render the coachmarks at a random location upon the component mounting 
	// OR a location change in the same file triggered by jump to def.
	_tryForRenderedTokenIdentifier(): void {
		// "token identifier go"" is subject to change based on the language. For now, since we are hardcoding the endpoint we can assume this will always be true.
		// however since we will move to make this onboarding more dynamic we will need this to be more robust by either exploring a VSCode widget or more generic DOM injection.
		let tokenElements = document.getElementsByClassName("token identifier go");
		if (!tokenElements || tokenElements.length <= 0) {
			// Correctly time the rendering of the tokens with the response from the async file response.
			// This results in no delay and not prematurely trying to render on a token (which wouldn't exist)
			window.requestAnimationFrame(this._tryForRenderedTokenIdentifier.bind(this));
		} else if ((window as any).ed) {
			let x = document.getElementsByClassName("token identifier go");
			if (x.length > 2) {
				// Grab a random element that has been indexed and provides "code intelligence".
				// Divide the total number of visible intelligent elements in half and pick a random node from the first half.
				// Render the first tooltip in the top half. Then render the second tooltip based on the second half of visible nodes.
				let random = Math.random() * x.length / 2;
				let refrandom = Math.random() * ((x.length - x.length / 2) - 1 ) + x.length / 2;
				let defRandom = x[Math.floor((random) + 1)];
				let refRandom = x[Math.floor(refrandom + 1)];

				// Build custom fields for coachmark.
				let defSubtitle = <p style={p}>Click on any reference to an identifier and jump to its definition – even if it's in another repository.</p>;
				let defActionCTA = <Button onClick={this._installChromeExtensionClicked.bind(this)} style={{marginLeft: whitespace[4], fontSize: 14}} color="darkBlue" size="small">Install the Chrome extension</Button>;

				let refSubtitle = <p style={p}>Right click this or any other identifier to access <strong>references, peek definitions</strong>, and other useful actions.</p>;
				let refActionCTA = <div style={{paddingLeft: whitespace[4]}}><GitHubAuthButton pageName="BlobViewOnboarding" img={false} color="darkBlue" scopes={privateGitHubOAuthScopes} returnTo={this.props.location}>Authorize with GitHub</GitHubAuthButton></div>;

				this._coachmarks = [
					this._initCoachmarkAnnotation(defRandom, "def-coachmark", "def-mark", _defCoachmarkIndex, "Jump to definition", defSubtitle, "Jump to definition and hover documentation on GitHub", defActionCTA),
					this._initCoachmarkAnnotation(refRandom, "ref-coachmark", "ref-mark", _refCoachmarkIndex, "View references and definitions", refSubtitle, "Enable these feature for your private code", refActionCTA),
				];

				this._coachmarksShouldUpdate();

				// Setup listener for the editor modifying the DOM. When lines are scrolled past they are removed from the view and therefore we have to re-add the tooltip
				// when the user scrolls the line number back into the view. (window as any).ed is a reference to the editor.
				(window as any).ed.onDidScrollChange(e => {
					this._coachmarksShouldUpdate();
				});
			}
		}
	};

	// Inits the a coachmark and annotation by first finding a parent element given the current file structure.
	// Then finds the parent's parent and sets a reference to the line where the coachmark is rendered.
	// Lastly after a valid line number and element is found, creates the jump to def tooltip and annotation.
	_initCoachmarkAnnotation(element: Element, markId: string, markParentElementId: string, markIndex: number, headingTitle: string, headingSubtitle: JSX.Element | null, actionTitle: string, actionCTA: JSX.Element | null): Coachmark {
		let grandparentElement = this._getGrandparentForElement(element);
		return {
			markId: markId,
			markParentElementId: markParentElementId,
			markIndex: markIndex,
			markLineNumber: Number(grandparentElement.getAttribute("linenumber")),
			headingTitle: headingTitle,
			headingSubtitle: headingSubtitle,
			actionTitle: actionTitle,
			actionCTA: actionCTA,
		};
	}

	_getGrandparentForElement(element: Element): Element {
		let firstParent = element.parentNode || element;
		let topParent = firstParent && firstParent.parentNode ? firstParent.parentNode : firstParent;
		let topElement = topParent as Element;
		return topElement;
	}

	_coachmarksShouldUpdate(): void {
		let {visibleMarks} = this.state;
		this._coachmarks.map((coachmark, index) => {
			// Remove the element if the coachmark should not be displayed.
			if (visibleMarks.indexOf(coachmark.markIndex) === -1) {
				// Timeout to prevent errors that can happen when performing DOM manipulations during a redirect.
				setTimeout(() => {
					let elementToRemove = document.getElementById(coachmark.markId);
					if (elementToRemove !== null) {
						elementToRemove.remove();
					}
				}, 10);
				return;
			}

			// Get currently visible lines.
			let lineView = (window as any).ed.getCompletelyVisibleLinesRangeInViewport();
			// Check that the desired element is within the currently visible range.
			if (coachmark.markLineNumber >= lineView["startLineNumber"] && coachmark.markLineNumber <= lineView["endLineNumber"]) {
				// Lines are removed from the dom and added back when the user scrolls therefore we we have to find the same element.
				// First grab all elements based on the same class. Then loop over each "token identifier" element and find it's parent's parent.
				// Compare the line number with the original line number and element. If they are the same check if the coachmark is currently rendered.
				// If the element has not been rendered yet create it and add it to the DOM. If it does exist overwrite the refParentElementId container so it is not lost during scroll.
				let tokenIdentifier = document.getElementsByClassName("token identifier go");
				for (let i = 0; i < tokenIdentifier.length; i++) {
					let element = tokenIdentifier[i];
					let grandparentElement = this._getGrandparentForElement(element);
					let grandparentLineNumber = Number(grandparentElement.getAttribute("linenumber"));
					if (grandparentLineNumber === coachmark.markLineNumber) {
						if (!document.getElementById(coachmark.markId)) {
							let overwrite = document.createElement("div");
							overwrite.style.cssText = parentElementCssString;
							overwrite.id = coachmark.markParentElementId;
							element.appendChild(overwrite);
							this._renderCoachmarkAnnotationForContainer(coachmark, overwrite);
							return;
						} else {
							let node = document.getElementById(coachmark.markParentElementId);
							this._renderCoachmarkAnnotationForContainer(coachmark, node);
						}
					}
				}
			}
		});
	}

	_renderCoachmarkAnnotationForContainer(coachmark: Coachmark, containerNode: any): void {
		let {visibleAnnotation} = this.state;
		let refs =
			<div id={coachmark.markId} onClick={this._handleCoachmarkClicked.bind(this, coachmark.markIndex)} style={{whitespace: "normal", lineHeight: "1.4"}}>
				<Annotation color="purple" pulseColor="white" open={visibleAnnotation === coachmark.markIndex} active={true} tooltipStyle={{whitespace: "normal !important", zIndex: 102, backgroundColor: colors.blue()}}>
					<span style={closeSx}><Close width={12} color={colors.coolGray2(0.9)}/></span>
					<div style={headerSx}>
						<Heading color="blue" level={6} style={{marginTop: 0}}>{coachmark.headingTitle}</Heading>
						{coachmark.headingSubtitle}
					</div>
					<div style={{padding: whitespace[4]}}>
						<Flag width={15} style={flagSx} color={colors.blue2(0.9)}/>
						<strong style={actionSx}>{coachmark.actionTitle}</strong>
						{coachmark.actionCTA}
					</div>
				</Annotation>
			</div>;

		ReactDOM.render(refs, containerNode);
	}

	_handleCoachmarkClicked(markIndex: number): void {
		switch (markIndex) {
			case _refCoachmarkIndex: {
				if (markIndex === this.state.visibleAnnotation) {
					EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_CLOSE, "ReferencesAnnotationCloseClicked",  {page_name: "BlobViewOnboarding"});
				} else {
					EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_CLICK, "ReferencesCoachmarkCTAClicked",  {page_name: "BlobViewOnboarding"});
				}
			}
			break;
			case _defCoachmarkIndex: {
				if (markIndex === this.state.visibleAnnotation) {
					EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_CLOSE, "JumpToDefAnnotationCloseClicked",  {page_name: "BlobViewOnboarding"});
				} else {
					EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_CLICK, "JumpToDefCoachmarkCTAClicked",  {page_name: "BlobViewOnboarding"});
				}
			}
			break;
			case _searchCoachmarkIndex: {
				if (markIndex === this.state.visibleAnnotation) {
					EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_CLOSE, "SearchAnnotationCloseClicked", {page_name: "BlobViewOnboarding"});
				} else {
					EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_CLICK, "SearchCoachmarkCTAClicked", {page_name: "BlobViewOnboarding"});
				}
			}
			break;
			default:
				return;
		}

		this._updateVisibleCoachmarks(markIndex);
	}

	// The search coachmark annotation is different because it does not live inside of the editor therefore we can render it like a standard react component.
	_renderSearchCoachmarkAnnotation(visibleAnnotation: number | null, markIndex: number): JSX.Element | null {
		return (
			<div onClick={this._handleCoachmarkClicked.bind(this, markIndex)} style={{ position: "fixed", right: 160, top: 40}}>
				<Annotation color="purple" pulseColor="white" open={visibleAnnotation === markIndex} active={true} annotationPosition="left">
					<span style={closeSx}><Close width={12}/></span>
					<div style={Object.assign({},
						headerSx,
						{ borderRadius: 3 },
					)}>
						<Heading color="blue" level={6}>Jump to symbols, files, and repositories</Heading>
						<p>Click Search or hit the <span style={searchSx}>/</span> key to open up search from anywhere.</p>
					</div>
				</Annotation>
			</div>
		);
	}

	_successHandler(): void {
		EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_SUCCESS, "ChromeExtensionInstalled", {page_name: "BlobViewOnboarding"});
		EventLogger.setUserProperty("installed_chrome_extension", "true");
		// Syncs the our site analytics tracking with the chrome extension tracker.
		EventLogger.updateTrackerWithIdentificationProps();
	}

	_failHandler(msg: String): void {
		EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_ERROR, "ChromeExtensionInstallFailed", {page_name: "BlobViewOnboarding"});
		EventLogger.setUserProperty("installed_chrome_extension", "false");
	}

	_installChromeExtensionClicked(): void {
		EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_CLICK, "ChromeExtensionCTAClicked", {page_name: "BlobViewOnboarding"});

		if (!!global.chrome) {
			EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_CLICK, "ChromeExtensionInstallStarted", {page_name: "BlobViewOnboarding"});
			global.chrome.webstore.install("https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack", this._successHandler.bind(this), this._failHandler.bind(this));
		} else {
			EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_CLICK, "ChromeExtensionStoreRedirect", {page_name: "BlobViewOnboarding"});
			window.open("https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack", "_newtab");
		}
	}

	_updateVisibleCoachmarks(index: number): void {
		let visibleMarks;
		// Update the visible coachmarks upon dismissal of the annotation.
		if (this.state.visibleAnnotation !== null) {
			visibleMarks = this.state.visibleMarks.filter((mark: number) => {
				return mark !== this.state.visibleAnnotation;
			});

			this.setState({
				visibleAnnotation: index !== this.state.visibleAnnotation ? index : null,
				visibleMarks: visibleMarks,
			});
		// else update the visible annotation.
		} else {
			visibleMarks = this.state.visibleMarks;

			this.setState({
				visibleAnnotation: index,
				visibleMarks: visibleMarks,
			});
		}

		window.sessionStorage.setItem("tour", String(visibleMarks));
		if (visibleMarks.length === 0) {
			EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_SUCCESS, "OnboardingTourCompleted", {page_name: "BlobViewOnboarding"});
			this._endTour();
		}
	}

	_renderDismissButton(): JSX.Element | null {
		return (
			<div style={{ position: "fixed", right: 42, bottom: 36}}>
				<Button onClick={this._dismissTour.bind(this)} size="small" color="white" outline={false}>Dismiss pointers</Button>
			</div>
		);
	}

	_dismissTour(): void {
		EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_ONBOARDING, AnalyticsConstants.ACTION_CLOSE, "DismissTourCTAClicked", {page_name: "BlobViewOnboarding"});
		this._endTour();
	}

	_endTour(): void {
		window.sessionStorage.setItem("tour", "");
		delete this.props.location.query["tour"];
		const newLoc = Object.assign({}, this.props.location, {query: this.props.location.query});
		(this.context as any).router.replace(newLoc);
	}

	render(): JSX.Element | null {
		let {visibleMarks, visibleAnnotation} = this.state;
		return (<div style={{zIndex: 101}}>
					{visibleMarks.indexOf(_searchCoachmarkIndex) !== -1 && this._renderSearchCoachmarkAnnotation(visibleAnnotation, _searchCoachmarkIndex)}
					{visibleMarks.length > 0 && this._renderDismissButton()}
				</div>
		);
	}
}
