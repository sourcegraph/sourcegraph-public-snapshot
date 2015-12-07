import React from "react";
import ReactDOM from "react-dom";

import EmojiMenu from "../reactions/EmojiMenu";

window.Sourcegraph = {
	Components: {
		emojiMenu(element, options) {
			ReactDOM.render(<EmojiMenu x={options.x} y={options.y} onSelect={options.onSelect} onClose={options.onClose}/>, element);
			return () => ReactDOM.unmountComponentAtNode(element);
		},
	},
};
