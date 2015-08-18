import "./client.less";

export default class ClientController {
	constructor(State) {
		this.token = State.get("token");
	}
}