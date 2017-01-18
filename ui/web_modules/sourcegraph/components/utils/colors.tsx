// This file defines our color palette with RGB values.
// Each color is called as a function and can be passed an opacity value.
// Default opacity value is 1
// colors.blue(0.2) => rgba(15, 182, 242, 0.2)

// Opacity can be determined in the style
const rgb = (r, g, b) => (a = 1) => `rgba(${r},${g},${b},${a})`;

// Color definitions
export const blueD2 = rgb(0, 63, 102);
export const blueD1 = rgb(0, 113, 184);
export const blue = rgb(0, 145, 234);
export const blueL1 = rgb(92, 192, 255);
export const blueL2 = rgb(173, 224, 255);
export const blueL3 = rgb(235, 247, 255);

export const purpleD2 = rgb(68, 0, 102);
export const purpleD1 = rgb(122, 0, 184);
export const purple = rgb(170, 0, 255);
export const purpleL1 = rgb(201, 92, 255);
export const purpleL2 = rgb(228, 173, 255);
export const purpleL3 = rgb(248, 235, 255);

export const orangeD2 = rgb(102, 24, 0);
export const orangeD1 = rgb(184, 43, 0);
export const orange = rgb(255, 61, 0);
export const orangeL1 = rgb(255, 130, 92);
export const orangeL2 = rgb(255, 192, 173);
export const orangeL3 = rgb(255, 239, 235);

export const greenD2 = rgb(8, 94, 83);
export const greenD1 = rgb(13, 156, 138);
export const green = rgb(0, 191, 165);
export const greenL1 = rgb(104, 243, 224);
export const greenL2 = rgb(180, 249, 240);
export const greenL3 = rgb(236, 253, 251);

export const yellowD2 = rgb(102, 85, 0);
export const yellowD1 = rgb(184, 153, 0);
export const yellow = rgb(255, 214, 0);
export const yellowL1 = rgb(255, 228, 92);
export const yellowL2 = rgb(255, 241, 173);
export const yellowL3 = rgb(255, 252, 235);

export const redD2 = rgb(102, 0, 20);
export const redD1 = rgb(184, 0, 37);
export const red = rgb(255, 10, 59);
export const redL1 = rgb(255, 133, 157);
export const redL2 = rgb(255, 173, 190);
export const redL3 = rgb(255, 235, 239);

export const blueGrayD2 = rgb(35, 48, 67);
export const blueGrayD1 = rgb(62, 87, 121);
export const blueGray = rgb(93, 126, 172);
export const blueGrayL1 = rgb(147, 169, 200);
export const blueGrayL2 = rgb(201, 212, 227);
export const blueGrayL3 = rgb(242, 244, 248);

export const black = rgb(0, 0, 0);
export const white = rgb(255, 255, 255);

export const text = blueGrayD1;
export const gray = blueGray;
