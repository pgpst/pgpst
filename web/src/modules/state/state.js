import moment from "moment";

class State {
	static factory($localStorage) {
		return new State($localStorage);
	}

	constructor($localStorage) {
		this.$localStorage = $localStorage;

		// Check the token
		if (this.$localStorage.token) {
			let token = this.$localStorage.token;
			let expiry = moment(token.expiry_date);

			if (token.expiry_date && expiry.unix() > 0) {
				if (expiry.isBefore(moment())) {
					// Token expired
					this.$localStorage.token = null;
					this.$localStorage.logged_in = false;
				}
			}
		} else {
			if (this.$localStorage.logged_in) {
				this.$localStorage.logged_in = false;
			}
		}
	}

	get(key) {
		return this.$localStorage[key];
	}

	set(key, value) {
		this.$localStorage[key] = value;
	}
}

State.factory.$inject = ["$localStorage"];
export default State;