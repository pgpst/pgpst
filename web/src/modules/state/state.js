class State {
	static factory($localStorage) {
		return new State($localStorage);
	}

	constructor($localStorage) {
		this.$localStorage = $localStorage;
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