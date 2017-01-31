import "whatwg-fetch";

import { IDisposable } from "vs/base/common/lifecycle";
import { IContextViewService } from "vs/platform/contextview/browser/contextView";
import { IInstantiationService } from "vs/platform/instantiation/common/instantiation";
import { IStatusbarItem } from "vs/workbench/browser/parts/statusbar/statusbar";

import { FeedbackDropdown, IFeedback, IFeedbackService } from "sourcegraph/feedback/feedback";
import { checkStatus } from "sourcegraph/util/xhr";

const enableFeedback = true;

class SlackFeedbackService implements IFeedbackService {
	// This webhook URL posts messages to the #feedback channel. It is
	// NOT secret (and are visible to anyone reading our JavaScript
	// bundle). While knowing it would allow anyone to post to our
	// #feedback channel, it does not let them read the channel.
	private static WEBHOOK_URL: string = "https://hooks.slack.com/services/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";

	public submitFeedback(feedback: IFeedback): void {
		// Use global fetch, not defaultFetch from
		// sourcegraph/util/xhr, because we are POSTing cross-domain
		// and do not want to include our auth headers.
		const sentimentEmoji = feedback.sentiment === 1 ? ":heart_eyes:" : ":rage:";
		fetch(SlackFeedbackService.WEBHOOK_URL, {
			method: "POST",
			body: JSON.stringify({
				text: `${sentimentEmoji} ${feedback.feedback} — at <${window.location.href}|${document.title}>`,
				unfurl_links: false,
			}),
		})
			.then(checkStatus)
			.catch(err => {
				console.error("Error submitting feedback:", err);
				alert("Error submitting feedback. Please email support@sourcegraph.com.");
			});
	}
}

export class FeedbackStatusbarItem implements IStatusbarItem {
	constructor(
		@IInstantiationService private instantiationService: IInstantiationService,
		@IContextViewService private contextViewService: IContextViewService
	) {
	}

	public render(element: HTMLElement): IDisposable {
		if (enableFeedback) {
			return this.instantiationService.createInstance(FeedbackDropdown, element, {
				contextViewProvider: this.contextViewService,
				feedbackService: this.instantiationService.createInstance(SlackFeedbackService)
			});
		}
		return { dispose(): void { /* noop */ } };
	}
}
