// font size in unit, line height
const scale = [
	[3.35, 1.45],
	[2.5, 1.45],
	[2.0, 1.45],
	[1.5, 1.45],
	[1.25, 1.45],
	[1, 1.45],
	[0.9, 1.45],
];

const unit = "rem";

const baseSizeDesktop = "15px";
const baseSizeMobile = "14px";

export const fontStack = {
	sansSerif: "'circular-std', -apple-system, BlinkMacSystemFont, '.SFNSText-Regular', 'San Francisco', 'SFNS Display', 'Roboto', 'Lucida Grande', 'avenir next', avenir, 'Helvetica', Ubuntu, 'segoe ui', arial, sans-serif;",
	code: "Menlo, Consolas, Monaco, monospace;",
};

export const weight = [ null, 100, 800 ];

export const codeSize = [
	{
		fontSize: 1 + unit,
		lineHeight: 1.45,
	},
	{
		fontSize: 0.9 + unit,
		lineHeight: 1.45,
	},
];

export const size = [
	null,
	{
		fontSize: scale[0][0] + unit,
		lineHeight: scale[0][1],
	},
	{
		fontSize: scale[1][0] + unit,
		lineHeight: scale[1][1],
	},
	{
		fontSize: scale[2][0] + unit,
		lineHeight: scale[2][1],
	},
	{
		fontSize: scale[3][0] + unit,
		lineHeight: scale[3][1],
	},
	{
		fontSize: scale[4][0] + unit,
		lineHeight: scale[4][1],
	},
	{
		fontSize: scale[5][0] + unit,
		lineHeight: scale[5][1],
	},
	{
		fontSize: scale[6][0] + unit,
		lineHeight: scale[6][1],
	},
];

export const small = size[0];
export const large = size[2];

export const typography = {
	size,
	small,
	large,
	codeSize,
	weight,
	baseSizeMobile,
	baseSizeDesktop,
};
