import angular from "angular";
import "ngstorage";


import State from "./state";

export default angular.module("pgpst.state", ["ngStorage"])
	.factory("State", State.factory)
	.name;