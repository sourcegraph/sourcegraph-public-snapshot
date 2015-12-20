import React from "react";
import ReactDOM from "react-dom";

import EmojiMenu from "sourcegraph/reactions/EmojiMenu";
import ReactionList from "sourcegraph/reactions/ReactionList";

let sourcegraph = {
	Components: {
		/**
		 * Display an emoji menu at the provided screen position.
		 *
		 * If there isn't an active logged in user, the menu will propmt the user
		 * to log in.
		 *
		 * @param {HTMLElement} element DOM element to render the menu inside.
		 * @param {Object} options Configuration options.
		 * @returns {function} A function that can be called to remove the rendered element from the DOM.
		 */
		emojiMenu(element, options) {
			ReactDOM.render(<EmojiMenu x={options.x} y={options.y} onSelect={options.onSelect} onClose={options.onClose}/>, element);
			return () => ReactDOM.unmountComponentAtNode(element);
		},

		/**
		 * Display a list of reactions.
		 *
		 * @param {HTMLElement} element DOM element to render the list inside.
		 * @param {Object} options Configuration options.
		 * @returns {function} A function that can be called to remove the rendered element from the DOM.
		 */
		reactionList(element, options) {
			ReactDOM.render(<ReactionList reactions={options.reactions} onSelect={options.onSelect} />, element);
			return () => ReactDOM.unmountComponentAtNode(element);
		},
	},
};

export default sourcegraph;
