import * as React from "react";
import {context} from "sourcegraph/app/context";
import {Button} from "sourcegraph/components";
import {GitHubIcon} from "sourcegraph/components/Icons";
import * as typography from "sourcegraph/components/styles/_typography.css";
import {whitespace} from "sourcegraph/components/utils";
import {Location} from "sourcegraph/Location";
import * as AnalyticsConstants from "sourcegraph/util/constants/AnalyticsConstants";
import {EventLogger} from "sourcegraph/util/EventLogger";
import {urlToOAuth} from "sourcegraph/util/urlTo";

interface Props {
	scopes?: string;
	returnTo?: string | Location | null;
	color?: string;
	outline?: boolean;
	block?: boolean;
	size?: string;
	children?: any;
	className?: string;
	tabIndex?: number;
	pageName?: string;
	img?: boolean;
}

type State = any;

export class GitHubAuthButton extends React.Component<Props, State> {
	static defaultProps: Props = {
		scopes: "",
		returnTo: null,
		color: "blue",
		outline: false,
		block: false,
		img: true,
	};

	render(): JSX.Element | null {
		const {scopes, returnTo, outline, color, block, children, size, pageName, className, tabIndex} = this.props;
		const url = urlToOAuth("github", scopes || "read:org,user:email", returnTo || null);
		return (
			<form method="POST" action={url}>
				<input type="hidden" name="gorilla.csrf.Token" value={context.csrfToken} />
				<Button
						type="submit"
						outline={outline}
						formNoValidate={true}
						color={color}
						block={block}
						size={size}
						className={className}
						tabIndex={tabIndex}
						onClick={() => {
							EventLogger.logEventForCategory(AnalyticsConstants.CATEGORY_AUTH, AnalyticsConstants.ACTION_CLICK, "InitiateGitHubOAuth2Flow", {page_name: pageName});
						}}>
					{this.props.img && <span style={{marginRight: whitespace[2]}}><GitHubIcon className={size === "small" ? typography.f5 : typography.f4}/></span>}
					{children}
				</Button>
			</form>
		);
	}
}
