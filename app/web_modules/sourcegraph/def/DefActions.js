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

export class WantRefLocations {
	constructor(defURL) {
		this.defURL = defURL;
	}
}

export class WantRefs {
	constructor(defURL, repo, file) {
		this.defURL = defURL;
		this.repo = repo || null;
		this.file = file || null;
	}
}

export class RefLocationsFetched {
	constructor(defURL, locations) {
		this.defURL = defURL;
		this.locations = locations || null;
	}
}

export class RefsFetched {
	constructor(defURL, repo, file, refs) {
		this.defURL = defURL;
		this.repo = repo || null;
		this.file = file || null;
		this.refs = refs;
	}
}
