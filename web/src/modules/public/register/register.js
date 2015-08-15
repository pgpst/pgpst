import $ from "jquery";
import "./register.less";

export default class RegisterController {
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