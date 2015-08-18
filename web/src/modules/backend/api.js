class API {
	authenticatePassword({
		address,
		password,
		expiry_time,
		client_id,
	}) {
		return this.Restangular.all("oauth").post({
			"grant_type": "password",
			"address":     address,
			"password":    password,
			"expiry_time": expiry_time,
			"client_id":   client_id,
		});
	}
	
	setToken(token) {
		this.token = token;
		this.Restangular.setDefaultHeaders({
			"Authorization": "Bearer " + token
		});
	}

	static factory(Restangular) {
		return new API(Restangular);
	}

	constructor(Restangular) {
		this.Restangular = Restangular;
	}
}

API.factory.$inject = ["Restangular"];
export default API;