/* Load vendor libraries */
import angular  from "angular";
import "angular-ui-router";
import "angular-scroll";

/* Load the application */

// Global CSS rules
import "./app.less";

// Modules
//import clientApp from "./modules/client";
import publicApp from "./modules/public";

angular.module("pgpst.app", [
	// vendor 
	"ui.router",
	"duScroll",

	// modules
	//clientApp,
	publicApp,
	/*"pgpst.client",
	"pgpst.public",*/
]).config(($urlRouterProvider, $locationProvider) => {
	$urlRouterProvider.otherwise('/');
	$locationProvider.html5Mode(true);
});
