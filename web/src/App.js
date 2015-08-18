/* Load vendor libraries */
import angular  from "angular";
import "angular-ui-router";
import "angular-scroll";
import "lodash";
import "restangular";

/* Load the application */

// Global CSS rules
import "./app.less";

// Modules
import "./modules/state";
import "./modules/backend";
import "./modules/client";
import "./modules/public";

angular.module("pgpst.app", [
	// vendor 
	"ui.router",
	"duScroll",
	"restangular",

	// modules
	"pgpst.state",
	"pgpst.backend",
	"pgpst.client",
	"pgpst.public",
]).config(($urlRouterProvider, $locationProvider) => {
	$urlRouterProvider.otherwise('/');
	$locationProvider.html5Mode(true);
}).config((RestangularProvider) => {
	RestangularProvider.setBaseUrl("http://127.0.0.1:8000/v1");
});
