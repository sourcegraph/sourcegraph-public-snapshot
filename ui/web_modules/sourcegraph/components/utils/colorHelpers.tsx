import * as _ from "lodash";

export function rgbStringToHex(str: string, alpha: boolean): string {
	const rgb = str.split(alpha ? "rgba(" : "rgb(")[1].split(")")[0].split(",");
	rgb.pop();
	const [r, g, b] = rgb.map(x => parseInt(x, 10));
	return rgbToHex(r, g, b);
};

export function rgbToHex(r: number, g: number, b: number): string {
	return `#${[r, g, b].map(x => _.padStart(x.toString(16), 2, "0")).join("")}`;
}
