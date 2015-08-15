routes.$inject = ["$stateProvider"];

export default function routes($stateProvider) {
	$stateProvider
		.state("public", {
			abstract:     true,
			template:     require("./public.html"),
			controller:   "PublicController",
			controllerAs: "public"
		})
		.state("public.home", {
			url:          "/",
			template:     require("./home/home.html"),
			controller:   "HomeController",
			controllerAs: "home"
		})
		.state("public.login", {
			url:          "/login",
			template:     require("./login/login.html"),
			controller:   "LoginController",
			controllerAs: "login"
		})
		.state("public.register", {
			url:          "/register",
			template:     require("./register/register.html"),
			controller:   "RegisterController",
			controllerAs: "register"
		});
}