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

export class WantDefs {
	constructor(repo, rev, query) {
		this.repo = repo;
		this.rev = rev;
		this.query = query;
	}
}

export class DefsFetched {
	constructor(repo, rev, query, defs) {
		this.repo = repo;
		this.rev = rev;
		this.query = query;
		this.defs = defs;
	}
}

export class SelectDef {
	constructor(url) {
		this.url = url;
	}
}

export class SelectMultipleDefs {
	constructor(urls, left, top) {
		this.urls = urls;
		this.left = left;
		this.top = top;
	}
}

export class HighlightDef {
	constructor(url) {
		this.url = url;
	}
}

export class WantExamples {
	constructor(defURL, page) {
		this.defURL = defURL;
		this.page = page;
	}
}

export class ExamplesFetched {
	constructor(defURL, page, examples) {
		this.defURL = defURL;
		this.page = page;
		this.examples = examples;
	}
}

export class NoExamplesAvailable {
	constructor(defURL, page) {
		this.defURL = defURL;
		this.page = page;
	}
}

export class WantRefs {
	constructor(defURL) {
		this.defURL = defURL;
	}
}

export class RefsFetched {
	constructor(defURL, refs) {
		this.defURL = defURL;
		this.refs = refs;
	}
}
