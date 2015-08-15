import angular  from "angular";
import uirouter from "angular-ui-router";

import publicRoutes     from "./routes";
import publicController from "./public";

import homeController     from "./home/home";
import loginController    from "./login/login";
import registerController from "./register/register";

export default angular.module("pgpst.public", [uirouter])
	.config(publicRoutes)
	.controller("PublicController", publicController)
	.controller("HomeController", homeController)
	.controller("LoginController", loginController)
	.controller("RegisterController", registerController)
	.name;
