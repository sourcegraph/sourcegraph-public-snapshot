import "whatwg-fetch";

import { IDisposable } from "vs/base/common/lifecycle";
import { IContextViewService } from "vs/platform/contextview/browser/contextView";
import { IInstantiationService } from "vs/platform/instantiation/common/instantiation";
import { IStatusbarItem } from "vs/workbench/browser/parts/statusbar/statusbar";

import { context } from "sourcegraph/app/context";
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
		const sentimentEmoji = feedback.sentiment === 1 ? ":heart_eyes:" : ":rage:";
		const user = feedback.email ? feedback.email : (context.user ? context.user.Login : "Anonymous user");

		doSubmitFeedback(SlackFeedbackService.WEBHOOK_URL, {
			text: `${sentimentEmoji} ${feedback.feedback} — by *${user}* at <${window.location.href}|${document.title}>\n\n<https://github.com/sourcegraph/sourcegraph/issues/new?title=${encodeURIComponent("[Feedback] " + feedback.feedback.slice(0, 30) + "...\n")}&body=${encodeURIComponent(feedback.feedback)}${encodeURIComponent("\n\nPosted by: **" + user + "**\n")}${encodeURIComponent("\nLocation: " + window.location.href)}|Create issue>`,
			unfurl_links: false,
		});
	}
}

class ZapierFeedbackService implements IFeedbackService {
	// This webhook URL posts messages to Zapier, which redirects it
	// to the Google Docs Product document
	// https://docs.google.com/a/sourcegraph.com/spreadsheets/d/1yt_bgb-lfGP7ugWerOF7xwxZbrbnXD0ghh9rks8fViQ/edit?usp=sharing
	private static WEBHOOK_URL: string = "https://hooks.zapier.com/hooks/catch/2112210/13py77/";

	public submitFeedback(feedback: IFeedback): void {
		const sentimentEmoji = feedback.sentiment === 1 ? ":heart_eyes:" : ":rage:";
		const emails = context.emails && context.emails.EmailAddrs || null;
		const primaryEmail = (emails && emails.filter(e => e.Primary).map(e => e.Email)[0]) || null;
		const email = feedback.email ? feedback.email : (primaryEmail ? primaryEmail : "Unknown");
		const userId = context.user ? context.user.Login : "Anonymous user";

		doSubmitFeedback(ZapierFeedbackService.WEBHOOK_URL, {
			email: email,
			user_id: userId,
			emotion: sentimentEmoji,
			feedback: feedback.feedback,
			feedback_url: window.location.href,
		});
	}
}

function doSubmitFeedback(url: string, body: any): Promise<Response> {
	// Use global fetch, not defaultFetch from
	// sourcegraph/util/xhr, because we are POSTing cross-domain
	// and do not want to include our auth headers.
	return fetch(url, {
		method: "POST",
		body: JSON.stringify(body),
	})
		.then(checkStatus)
		.catch(err => {
			console.error("Error submitting feedback:", err);
		});
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
				feedbackServices: [
					this.instantiationService.createInstance(SlackFeedbackService),
					this.instantiationService.createInstance(ZapierFeedbackService),
				]
			});
		}
		return { dispose(): void { /* noop */ } };
	}
}
