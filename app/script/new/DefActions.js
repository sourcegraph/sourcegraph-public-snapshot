export class WantDef {
	constructor(url) {
		this.url = url;
	}
}

export class DefFetched {
	constructor(url, def) {
		this.url = url;
		this.def = def;
	}
}

export class SelectDef {
	constructor(url) {
		this.url = url;
	}
}

export class HighlightDef {
	constructor(url) {
		this.url = url;
	}
}

export class GoToDef {
	constructor(url) {
		this.url = url;
	}
}
