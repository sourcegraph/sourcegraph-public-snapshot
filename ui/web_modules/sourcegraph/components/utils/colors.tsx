// This file defines our color palette with RGB values.
// Each color is called as a function and can be passed an opacity value.
// Default opacity value is 1
// colors.blue(0.2) => rgba(15, 182, 242, 0.2)

// Opacity can be determined in the style
const rgb = (r, g, b) => (a = 1) => `rgba(${r},${g},${b},${a})`;

// Color definitions
const blue1			= rgb(0, 82, 140);
const blue2			= rgb(0, 146, 214);
const blue3			 = rgb(15, 182, 242);
const blue4			= rgb(119, 209, 242);
const blue5			= rgb(192, 235, 250);

const purple1		= rgb(123, 13, 172);
const purple2		= rgb(177, 20, 247);
const purple3		= rgb(206, 122, 255);
const purple4		= rgb(232, 195, 250);

const orange1		 = rgb(230, 69, 0);
const orange2		= rgb(248, 96, 18);
const orange3		= rgb(255, 165, 120);

const green1		 = rgb(5, 128, 128);
const green2		 = rgb(10, 166, 148);
const green3		 = rgb(13, 206, 183);

const yellow		 = rgb(265, 199, 112);

const red1			 = rgb(246, 106, 106);
const red2			 = rgb(242, 73, 73);

const coolGray1	= rgb(26, 48, 77);
const coolGray2	= rgb(52, 73, 102);
const coolGray3	= rgb(119, 147, 174);
const coolGray4	= rgb(213, 229, 242);

const black			= rgb(0, 0, 0);
const white			= rgb(255, 255, 255);

export const colors = {
	blue1,
	blue2,
	blue3,
	blue4,
	blue5,

	purple1,
	purple2,
	purple3,
	purple4,

	orange1,
	orange2,
	orange3,

	green1,
	green2,
	green3,

	yellow,

	red1,
	red2,

	coolGray1,
	coolGray2,
	coolGray3,
	coolGray4,

	black,
	white,

	// Nice names and aliases
	blueText		: blue1,
	blue				: blue3,
	purple			: purple2,
	orange			: orange2,
	green			 : green3,
	greenText	 : green2,
	redText		 : red2,
	text				: coolGray1,
};
