import { whitespace } from "sourcegraph/components/utils";

type Direction = "y" | "x" | "all";
type Scale = 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7;

export function padding(direction: Direction, scale: Scale): React.CSSProperties {
	const size = whitespace[scale];

	switch (direction) {
		case "y":
			return { paddingBottom: size, paddingTop: size };
		case "x":
			return { paddingLeft: size, paddingRight: size };
		case "all":
			return { padding: size };
	}
};

export function margin(direction: Direction, scale: Scale): React.CSSProperties {
	const size = whitespace[scale];

	switch (direction) {
		case "y":
			return { marginBottom: size, marginTop: size };
		case "x":
			return { marginLeft: size, marginRight: size };
		case "all":
			return { margin: size };
	}
};
