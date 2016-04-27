import {applyMiddleware, createStore, combineReducers, compose} from "redux";
import rootReducer from "../reducers";
import thunk from "redux-thunk";
import createLogger from "redux-logger";
import storage from "../utils/storage";

const enhancer = compose(
	applyMiddleware(thunk, createLogger()),
	storage(),
	window.devToolsExtension ? window.devToolsExtension() : (nope) => nope
);

export default function (initialState) {
	const store = createStore(rootReducer, initialState, enhancer);

	if (module.hot) {
		module.hot.accept("../reducers", () => {
			const nextRootReducer = require("../reducers");
			store.replaceReducer(nextRootReducer);
		});
	}
	return store;
}
