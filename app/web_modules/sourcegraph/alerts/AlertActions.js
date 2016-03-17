export class AddAlert {
	constructor(autoDismiss, component) {
		this.autoDismiss = autoDismiss;
		this.component = component;
	}
}

export class RemoveAlert {
	constructor(id) {
		this.id = id;
	}
}
