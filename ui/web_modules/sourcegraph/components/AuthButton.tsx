import * as React from "react";

import { Button, SplitButton } from "sourcegraph/components";
import { ButtonProps } from "sourcegraph/components/Button";
import { GitHubLogo } from "sourcegraph/components/symbols";
import { colors, typography, whitespace } from "sourcegraph/components/utils";
import { ComponentWithRouter } from "sourcegraph/core/ComponentWithRouter";
import { githubAuthAction } from "sourcegraph/user/Auth";

interface Props extends ButtonProps {
	secondaryText?: string;
	privateCode: boolean;
}

export class AuthButton extends ComponentWithRouter<Props, {}> {
	render(): JSX.Element {
		const {
			secondaryText,
			children,
			privateCode,
			...btnProps
		} = this.props;

		const iconSx = this.props.size === "small" ? typography.size[5] : typography.size[4];
		const iconSize = 24;

		const { submit, form } = githubAuthAction(this.context.router, privateCode);

		if (secondaryText) {
			return <SplitButton onClick={submit} {...btnProps} secondaryText={secondaryText}>
				{form}
				<GitHubLogo width={iconSize} style={iconSx} />
				{children}
			</SplitButton>;
		}

		return <Button {...btnProps} onClick={submit} style={{
			padding: 0,
			textAlign: "left",
		}}>
			{form}
			<span style={{
				backgroundColor: colors.black(0.25),
				display: "inline-block",
				padding: whitespace[2],
			}}>
				<GitHubLogo width={iconSize} style={iconSx} />
			</span>
			<span style={{
				marginLeft: whitespace[3],
				marginRight: whitespace[3],
			}}>{children}</span>
		</Button>;
	}
}
