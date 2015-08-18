import openpgp from "openpgp";
import "./login.less";

export default class LoginController {
	constructor($scope, $state, API, State) {
		this.$scope = $scope;
		this.$state = $state;
		this.API = API;
		this.State = State;

		this.invalid = {};
	}

	resetAddressValid() {
		this.invalid.address = false;
	}

	resetPasswordValid() {
		this.invalid.password = false;
	}

	async onSubmit() {
		console.log(this.address);
		console.log(this.password);
		let shapwd = openpgp.util.hexidump(openpgp.crypto.hash.sha256(this.password));
		console.log(shapwd);
		console.log(this.remember);

		try {
			let token = await this.API.authenticatePassword({
				"address":     this.address,
				"password":    shapwd,
				"expiry_date": 60 * 60 * 24,
				"client_id":   process.env.CLIENT_ID
			});

			this.invalid = {};
			this.API.setToken(token.id);

			this.State.set("token", token);

			console.log(token);

			this.$state.go("client.demo");
		} catch (error) {
			if (error.data && error.code && (error.data.code == 2005 || error.data.code == 2006)) {
				switch(error.data.code) {
					case 2005:
						// Invalid address
						this.invalid["address"] = true;
						break;
					case 2006:
						// Invalid password
						this.invalid["password"] = true;
						break;
				}

				this.$scope.$digest();
			} else {
				throw error;
			}
		}
	}
}