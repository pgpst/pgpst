import $ from "jquery";
import "./login.less";

export default class LoginController {
	constructor($rootScope) {
		[
			"reserve-menu",
			"features-menu",
 			"pricing-menu",
			"roadmap-menu"
		].forEach((element) => {
			$rootScope.$broadcast('duScrollspy:becameInactive', $("#" + element));
		});
	}

	changeName() {
		this.name = "pgp.st";
	}
}