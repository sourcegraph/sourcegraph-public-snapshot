import {rootReducer} from "../reducers";
import {saveAccessTokenMiddleware} from "../utils/middleware";
import {applyMiddleware, compose, createStore} from "redux";

// tslint:disable-next-line
const thunk = require("redux-thunk").default;

const middlewares = applyMiddleware(thunk);
const enhancer = compose(
	middlewares,
	saveAccessTokenMiddleware,
);

export function configureStore(initialState) {
	return createStore(rootReducer, initialState, enhancer);
}
