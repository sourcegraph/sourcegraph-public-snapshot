const boxShadow = "0 3px 12px rgba(27,31,35,0.15)";
const borderRadius = "3px";
const normalFontFamily = `-apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol"`;
const codeFontFamily = `"SFMono-Regular", Consolas, "Liberation Mono", Menlo, Courier, monospace`;
const normalFontColor = "rgb(36, 41, 69)";
const emphasisFontColor = "#666";
const modalBorderColor = "#afb2b7";
const dividerColor = "#e1e4e8";
const lightContrastBackground = "#fafbfc";

export const tooltip = {
	backgroundColor: lightContrastBackground,
	maxWidth: "500px",
	border: `solid 1px ${modalBorderColor}`,
	fontFamily: normalFontFamily,
	color: normalFontColor,
	fontSize: "12px",
	zIndex: 100,
	position: "absolute",
	overflow: "auto",
	borderRadius,
	boxShadow,
};

export const tooltipTitle = {
	fontFamily: codeFontFamily,
	wordWrap: "break-word",
	marginLeft: "24px",
	marginRight: "32px",
	padding: "0px",
};

export const divider = {
	borderBottom: `1px solid ${dividerColor}`,
	padding: "16px",
	lineHeight: "16px",
};

export const tooltipDoc = {
	paddingTop: "16px",
	paddingLeft: "16px",
	paddingRight: "16px",
	paddingBottom: "6px",
	maxHeight: "150px",
	overflow: "auto",
	marginBottom: "0px",
	fontFamily: normalFontFamily,
	borderBottom: `1px solid ${dividerColor}`,
};

export const tooltipActions = {
	display: "flex",
	textAlign: "center",
	padding: "16px",
};

export const tooltipAction = {
	flex: 1,
	cursor: "pointer",
};

export const tooltipMoreActions = {
	fontStyle: "italic",
	fontWeight: "bold",
	color: emphasisFontColor,
	padding: "16px",
};

export const fileNavButton = {
	borderTopLeftRadius: 0,
	borderBottomLeftRadius: 0,
	color: "black",
	textDecoration: "none",
};

export const sourcegraphIcon = {
	width: "16px",
	position: "absolute",
	left: "16px",
	marginBottom: "2px",
	verticalAlign: "middle",
};

export const closeIcon = {
	width: "12px",
	position: "absolute",
	top: "18px",
	right: "16px",
	verticalAlign: "middle",
	cursor: "pointer",
};

export const loadingTooltip = {
	padding: "16px",
};
