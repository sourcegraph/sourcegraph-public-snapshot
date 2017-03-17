import { $, Builder } from "vs/base/browser/builder";
import { Dropdown } from "vs/base/browser/ui/dropdown/dropdown";
import { IDisposable } from "vs/base/common/lifecycle";
import * as nls from "vs/nls";
import { IContextViewService } from "vs/platform/contextview/browser/contextView";
import { ITelemetryService } from "vs/platform/telemetry/common/telemetry";
import "vs/workbench/parts/feedback/electron-browser/media/feedback.css";

export interface IFeedback {
	feedback: string;
	sentiment: number;
}

export interface IFeedbackService {
	submitFeedback(feedback: IFeedback): void;
}

export interface IFeedbackDropdownOptions {
	contextViewProvider: IContextViewService;
	feedbackService: IFeedbackService;
}

enum FormEvent {
	SENDING,
	SENT,
	SEND_ERROR
}

export class FeedbackDropdown extends Dropdown {
	protected feedback: string;
	protected sentiment: number;
	protected isSendingFeedback: boolean;
	protected autoHideTimeout: number | null;

	protected feedbackService: IFeedbackService;

	protected feedbackForm: HTMLFormElement | null;
	protected feedbackDescriptionInput: HTMLTextAreaElement | null;
	protected smileyInput: Builder | null;
	protected frownyInput: Builder | null;
	protected sendButton: Builder | null;

	protected requestFeatureLink: string;
	protected reportIssueLink: string;

	constructor(
		container: HTMLElement,
		options: IFeedbackDropdownOptions,
		@ITelemetryService protected telemetryService: ITelemetryService,
	) {
		super(container, {
			contextViewProvider: options.contextViewProvider,
			labelRenderer: (container: HTMLElement): IDisposable => { // tslint:disable-line no-shadowed-variable
				$(container).addClass("send-feedback");
				return { dispose(): void {/* noop */ } };
			}
		});

		this.$el.addClass("send-feedback");
		this.$el.title("Send Feedback");

		this.feedbackService = options.feedbackService;

		this.feedback = "";
		this.sentiment = 1;

		this.feedbackForm = null;
		this.feedbackDescriptionInput = null;

		this.smileyInput = null;
		this.frownyInput = null;

		this.sendButton = null;

		this.reportIssueLink = "abc";
		this.requestFeatureLink = "xyz";
	}

	public renderContents(container: HTMLElement): IDisposable {
		const $form = $("form.feedback-form").attr({
			action: "javascript:void(0);",
			tabIndex: "-1"
		}).appendTo(container);

		$(container).addClass("monaco-menu-container");

		this.feedbackForm = $form.getHTMLElement() as HTMLFormElement;

		$("h2.title").text("Share your feedback").appendTo($form);

		this.invoke($("div.cancel").attr("tabindex", "0"), () => {
			this.hide();
		}).appendTo($form);

		const $content = $("div.content").appendTo($form);

		const $sentimentContainer = $("div").appendTo($content);
		$("span").text(nls.localize("sentiment", "How was your experience?")).appendTo($sentimentContainer);

		const $feedbackSentiment = $("div.feedback-sentiment").appendTo($sentimentContainer);

		this.smileyInput = $("div").addClass("sentiment smile").attr({
			"aria-checked": "false",
			"aria-label": nls.localize("smileCaption", "Happy"),
			"tabindex": 0,
			"role": "checkbox"
		});
		this.invoke(this.smileyInput, () => { this.setSentiment(true); }).appendTo($feedbackSentiment);

		this.frownyInput = $("div").addClass("sentiment frown").attr({
			"aria-checked": "false",
			"aria-label": nls.localize("frownCaption", "Sad"),
			"tabindex": 0,
			"role": "checkbox"
		});

		this.invoke(this.frownyInput, () => { this.setSentiment(false); }).appendTo($feedbackSentiment);

		if (this.sentiment === 1) {
			this.smileyInput.addClass("checked").attr("aria-checked", "true");
		} else {
			this.frownyInput.addClass("checked").attr("aria-checked", "true");
		}

		$("h3").text("Tell us why:")
			.appendTo($form);

		this.feedbackDescriptionInput = $("textarea.feedback-description").attr({
			rows: 3,
			"aria-label": nls.localize("commentsHeader", "Comments")
		})
			.text(this.feedback).attr("required", "required")
			.on("keyup", () => {
				this.feedbackDescriptionInput!.value ? this.sendButton!.removeAttribute("disabled") : this.sendButton!.attr("disabled", "");
			})
			.appendTo($form).domFocus().getHTMLElement() as HTMLTextAreaElement;

		const $buttons = $("div.form-buttons").appendTo($form);

		this.sendButton = this.invoke($("input.send").type("submit").style("background-image", "none").style("padding-left", "12px").style("width", "auto").attr("disabled", "").value("Send feedback").appendTo($buttons), () => {
			if (this.isSendingFeedback) {
				return;
			}
			this.onSubmit();
		});

		return {
			dispose: () => {
				this.feedbackForm = null;
				this.feedbackDescriptionInput = null;
				this.smileyInput = null;
				this.frownyInput = null;
			}
		};
	}

	protected setSentiment(smile: boolean): void {
		if (smile) {
			this.smileyInput!.addClass("checked");
			this.smileyInput!.attr("aria-checked", "true");
			this.frownyInput!.removeClass("checked");
			this.frownyInput!.attr("aria-checked", "false");
		} else {
			this.frownyInput!.addClass("checked");
			this.frownyInput!.attr("aria-checked", "true");
			this.smileyInput!.removeClass("checked");
			this.smileyInput!.attr("aria-checked", "false");
		}
		this.sentiment = smile ? 1 : 0;
	}

	protected invoke(element: Builder, callback: () => void): Builder {
		element.on("click", callback);
		element.on("keypress", (e) => {
			if (e instanceof KeyboardEvent) {
				const keyboardEvent = e as KeyboardEvent;
				if (keyboardEvent.keyCode === 13 || keyboardEvent.keyCode === 32) { // Enter or Spacebar
					callback();
				}
			}
		});
		return element;
	}

	public hide(): void {
		if (this.feedbackDescriptionInput) {
			this.feedback = this.feedbackDescriptionInput.value;
		}

		if (this.autoHideTimeout) {
			clearTimeout(this.autoHideTimeout);
			this.autoHideTimeout = null;
		}

		super.hide();
	}

	public onEvent(e: Event, activeElement: HTMLElement): void {
		if (e instanceof KeyboardEvent) {
			const keyboardEvent = e as KeyboardEvent;
			if (keyboardEvent.keyCode === 27) { // Escape
				this.hide();
			}
		}
	}

	protected onSubmit(): void {
		if ((this.feedbackForm!.checkValidity && !this.feedbackForm!.checkValidity())) {
			return;
		}

		this.changeFormStatus(FormEvent.SENDING);

		this.feedbackService.submitFeedback({
			feedback: this.feedbackDescriptionInput!.value,
			sentiment: this.sentiment
		});

		this.changeFormStatus(FormEvent.SENT);
	}

	private changeFormStatus(event: FormEvent): void {
		switch (event) {
			case FormEvent.SENDING:
				this.isSendingFeedback = true;
				this.sendButton!.setClass("send in-progress");
				this.sendButton!.value(nls.localize("feedbackSending", "Sending"));
				break;
			case FormEvent.SENT:
				this.isSendingFeedback = false;
				this.sendButton!.setClass("send success").value(nls.localize("feedbackSent", "Thanks"));
				this.resetForm();
				this.autoHideTimeout = setTimeout(() => {
					this.hide();
				}, 1000);
				this.sendButton!.off(["click", "keypress"]);
				this.invoke(this.sendButton!, () => {
					this.hide();
					this.sendButton!.off(["click", "keypress"]);
				});
				break;
			case FormEvent.SEND_ERROR:
				this.isSendingFeedback = false;
				this.sendButton!.setClass("send error").value(nls.localize("feedbackSendingError", "Try again"));
				break;
		}
	}

	protected resetForm(): void {
		if (this.feedbackDescriptionInput) {
			this.feedbackDescriptionInput.value = "";
		}
		this.sentiment = 1;
	}
}
