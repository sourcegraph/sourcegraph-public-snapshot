const boxShadow = "0 3px 12px rgba(27,31,35,0.15)";
const borderRadius = "3px";
const normalFontFamily = `-apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol"`;
const codeFontFamily = `"SFMono-Regular", Consolas, "Liberation Mono", Menlo, Courier, monospace`;
const normalFontColor = "rgb(36, 41, 69)";
const emphasisFontColor = "#666";
const modalBorderColor = "#afb2b7";
const dividerColor = "#e1e4e8";
const lightContrastBackground = "#fafbfc";
const altLightContrastBackground = "#f6f8fa";

export const tooltip = {
	backgroundColor: lightContrastBackground,
	maxWidth: "500px",
	maxHeight: "250px",
	border: `solid 1px ${modalBorderColor}`,
	fontFamily: normalFontFamily,
	color: normalFontColor,
	fontSize: "12px",
	zIndex: 100,
	position: "absolute",
	overflow: "auto",
	padding: "5px 5px",
	borderRadius,
	boxShadow,
};

export const tooltipTitle = {
	fontFamily: codeFontFamily,
	wordWrap: "break-word",
	paddingBottom: "5px",
	borderBottom: `solid 1px ${dividerColor}`,
};

export const tooltipDoc = {
	paddingTop: "5px",
	paddingLeft: "5px",
	paddingRight: "10px",
	maxHeight: "150px",
	overflow: "auto",
	fontFamily: normalFontFamily,
	borderBottom: `1px solid ${dividerColor}`,
};

export const tooltipActions = {
	display: "flex",
	textAlign: "center",
	paddingTop: "5px",
};

export const tooltipAction = {
	flex: 1,
	cursor: "pointer",
};

export const tooltipMoreActions = {
	fontStyle: "italic",
	fontWeight: "bold",
	color: emphasisFontColor,
	paddingTop: "5px",
	paddingLeft: "5px",
	paddingRight: "5px",
};
