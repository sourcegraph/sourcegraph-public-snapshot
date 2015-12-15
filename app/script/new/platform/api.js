import React from "react";
import ReactDOM from "react-dom";

import EmojiMenu from "../reactions/EmojiMenu";
import ReactionList from "../reactions/ReactionList";

window.Sourcegraph = {
	Components: {
		emojiMenu(element, options) {
			ReactDOM.render(<EmojiMenu x={options.x} y={options.y} onSelect={options.onSelect} onClose={options.onClose}/>, element);
			return () => ReactDOM.unmountComponentAtNode(element);
		},

		reactionList(element, options) {
			ReactDOM.render(<ReactionList reactions={options.reactions} currentUser={options.currentUser} onSelect={options.onSelect} />, element);
			return () => ReactDOM.unmountComponentAtNode(element);
		},
	},
};
