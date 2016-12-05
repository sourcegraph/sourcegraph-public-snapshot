import { keyframes, style as gStyle } from "glamor";
import * as React from "react";
import { colors } from "sourcegraph/components/utils";

const fillColors = [
	colors.blue(),
	colors.red1(),
	colors.green(),
	colors.purple(),
	colors.orange(),
	colors.yellow(),
];

function getRandomInt(min: number, max: number): number {
	return Math.floor(Math.random() * (max - min) + min);
};

function getRandomColor(): string {
	return fillColors[getRandomInt(0, fillColors.length)];
}

function generateParticles(count: number): JSX.Element[] {
	let particles = [];
	for (let i = 0; i < count; i++) {
		let startY = getRandomInt(-5, 5);
		let startX = getRandomInt(-5, 5);
		let endY = startY + getRandomInt(getRandomInt(200, 250) * -1, getRandomInt(200, 250));
		let endX = startX + getRandomInt(getRandomInt(200, 250) * -1, getRandomInt(200, 250));

		let size = getRandomInt(15, 80);
		let color = getRandomColor();

		let particleMovement = keyframes({
			"15%": {
				transform: "scale(1.1)",
				left: endX / 2,
				top: endY / 2,
			},
			"100%": {
				transform: "scale(0)",
				left: endX,
				top: endY,
			},
		});

		(particles as JSX.Element[]).push(
			<div key={i} {...gStyle({
				backgroundColor: color,
				position: "absolute",
				left: startX,
				top: startY,
				width: size,
				height: size,
				borderRadius: "50%",
				boxShadow: `0 0 2px 1px ${colors.black(0.1)}`,
				animation: `${particleMovement} 850ms cubic-bezier(0.000, 0.735, 1.000, 1.010) 1 forwards`,
			}) }></div>
		);
	}
	return particles;
}

interface Props { style: React.CSSProperties; }

export function Boom({style}: Props): JSX.Element {
	const particles = generateParticles(getRandomInt(20, 25));
	const sx = Object.assign(
		{ position: "relative" },
		style,
	);
	return <div style={sx}>{particles}</div>;
}
