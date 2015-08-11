import React from "react";
import BrowserHistory from "react-router/lib/BrowserHistory";
import HashHistory from "react-router/lib/HashHistory";
import Root from "./Root";

const rootEl = document.getElementById('root');
/*const history = process.env.NODE_ENV === 'production' ?
	new HashHistory() :
	new BrowserHistory();*/
const history = new BrowserHistory();

React.render(<Root history={history} />, rootEl);