routes.$inject = ["$stateProvider"];

export default function routes($stateProvider) {
	$stateProvider
		.state("client", {
			abstract:     true,
			template:     require("./client.html"),
			controller:   "ClientController",
			controllerAs: "client"
		})
		.state("client.demo", {
			url:          "/client",
			template:     require("./demo/demo.html"),
			controller:   "DemoController",
			controllerAs: "demo"
		})
}