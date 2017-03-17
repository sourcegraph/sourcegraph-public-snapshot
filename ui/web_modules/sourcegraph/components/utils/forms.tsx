import { css } from "glamor";
import * as colors from "sourcegraph/components/utils/colors";

const backgroundColor = "white";
const errorColor = colors.orange();

const borderColor = colors.blueGrayL2();
const focusBorderColor = colors.blueGrayL1();

const textColor = colors.blueGrayD1();
const placeholderColor = colors.blueGray(0.7);

export const focus = `inset 0 0 0 1px ${focusBorderColor} !important`;
export const error = `inset 0 0 0 1px ${errorColor}`;

export const style = css({
	appearance: "none",
	borderRadius: 3,
	backgroundColor: backgroundColor,
	boxShadow: `inset 0 0 0 1px ${borderColor}`,
	border: "none",
	boxSizing: "border-box",
	color: textColor,
	outline: "none",
	transition: "all 0.25s ease-in-out",
	":focus": { boxShadow: focus },
	"::-webkit-input-placeholder": { color: placeholderColor },
	"::-moz-placeholder": { color: placeholderColor },
	":-moz-placeholder": { color: placeholderColor },
	":-ms-input-placeholder": { color: placeholderColor },
});
