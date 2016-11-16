import {ReducerState, rootReducer} from "../reducers";
import {Store, applyMiddleware, compose, createStore} from "redux";

export function configureStore(): Store<ReducerState> {
	return createStore(rootReducer, compose(applyMiddleware(require("redux-thunk").default)));
}
