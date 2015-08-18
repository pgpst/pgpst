import "./public.less";

export default class PublicController {
	constructor($location) {
		this.isActive = (location) => {
			return $location.path().indexOf(location) == 0;
		}
	}
}