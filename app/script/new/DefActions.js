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
	constructor(def) {
		this.def = def;
	}
}

export class HighlightDef {
	constructor(def) {
		this.def = def;
	}
}

export class GoToDef {
	constructor(url) {
		this.url = url;
	}
}
