import angular from "angular";
import "angular-ui-router";

import clientRoutes     from "./routes";
import clientController from "./client";

import demoController from "./demo/demo";

export default angular.module("pgpst.client", ["ui.router"])
	.config(clientRoutes)
	.controller("ClientController", clientController)
	.controller("DemoController", demoController)
	.name;