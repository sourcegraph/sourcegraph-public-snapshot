export class AddAlert {
	constructor(autoDismiss, html) {
		this.autoDismiss = autoDismiss;
		this.html = html;
	}
}

export class RemoveAlert {
	constructor(id) {
		this.id = id;
	}
}
