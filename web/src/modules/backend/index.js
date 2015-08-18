import angular from "angular";
import "restangular";

import API    from "./api";
import Crypto from "./crypto";

export default angular.module("pgpst.backend", ["restangular"])
	.factory("API", API.factory)
	.factory("Crypto", Crypto.factory)
	.name;