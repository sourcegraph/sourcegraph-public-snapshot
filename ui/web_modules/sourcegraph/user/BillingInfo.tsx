import * as React from "react";

import { context } from "sourcegraph/app/context";

interface Props {
	closed: () => void;
	amount: number;
	submitLabel?: string;
	description?: string;
	submit: (token: any) => void;
}

export class PaymentInfo extends React.Component<Props, {}> {

	handler: any = null;

	domRef = (ref: HTMLDivElement) => {
		if (!ref) {
			return;
		}
		const script = document.createElement("script");
		script.src = "https://checkout.stripe.com/checkout.js";
		script.onload = this.initialize;
		ref.appendChild(script);
	}

	initialize = () => {
		const StripeCheckout = (window as any).StripeCheckout;
		this.handler = StripeCheckout.configure({
			key: context.stripePublicKey,
			image: `${context.assetsRoot}/img/sourcegraph-stripe.png`,
			locale: "auto",
			token: this.props.submit,
			closed: this.props.closed,
			panelLabel: this.props.submitLabel,
		});
		this.handler.open({
			name: "Sourcegraph",
			amount: this.props.amount,
			description: this.props.description,
		});
	}

	componentWillUnmount(): void {
		if (this.handler) {
			this.handler.close();
		}
	}

	render(): JSX.Element {
		return <div style={{ display: "none" }} ref={this.domRef} />;
	}

}

interface State {
	show: boolean;
}

export class ChangeBillingInfo extends React.Component<{}, State> {
	state: State = { show: false };

	showModal = () => this.setState({ show: true });

	close = () => this.setState({ show: false });

	submit = () => {
		// TODO
	}

	render(): JSX.Element {
		return <div>
			<a onClick={this.showModal}>Update payment details</a>
			{this.state.show && <PaymentInfo closed={this.close} submitLabel="Update card details" amount={0} submit={this.submit} />}
		</div>;
	}
}
