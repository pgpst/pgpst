import { Dispatcher } from "flux";

const flux = new Dispatcher();

export function register(callback) {
	return flux.register(callback);
}

export function waitFor(ids) {
	return flux.waitFor(ids);
}

/**
 * Dispatches a single action.
 */
export function dispatch(type, action = {}) {
	if (!type) {
		throw new Error("You forgot to specify type.");
	}

	if (process.env.NODE_ENV !== "production") {
		if (action.error) {
			console.error(type, action);
		} else {
			console.log(type, action);
		}
	}

	flux.dispatch({ type, ...action });
}

/**
* Dispatches three actions for an async operation represented by promise.
*/
export function dispatchAsync(promise, types, action = {}) {
	const { request, success, failure } = types;

	dispatch(request, action);
	promise.then(
		response => dispatch(success, { ...action, response }),
		error => dispatch(failure, { ...action, error })
	);
}