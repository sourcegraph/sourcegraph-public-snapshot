import {applyMiddleware, createStore, combineReducers, compose} from "redux";
import rootReducer from "../reducers";
import thunk from "redux-thunk";
import storage from "../utils/storage";

const middlewares = applyMiddleware(thunk);
const enhancer = compose(
	middlewares,
	storage()
);

export default function(initialState) {
	return createStore(rootReducer, initialState, enhancer);
}
