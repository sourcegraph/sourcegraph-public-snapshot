import * as React from "react";

interface Props {
	children?: React.ReactNode[];
	img: string;
	position?: string;
	repeat?: string;
	size?: string;
	style?: React.CSSProperties;
}

export function BGContainer(props: Props): JSX.Element {
	return <div style={Object.assign({},
		{
			backgroundImage: `url('${props.img}')`,
			backgroundPosition: props.position,
			backgroundRepeat: props.repeat,
			backgroundSize: props.size,
		},
		props.style,
	)}>{props.children}</div>;
}
