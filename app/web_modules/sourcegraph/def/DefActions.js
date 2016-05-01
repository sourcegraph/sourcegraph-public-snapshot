// @flow

import type {Def, Ref, RefLocationsKey} from "sourcegraph/def";
import toQuery from "sourcegraph/util/toQuery";

export class WantDef {
	repo: string;
	rev: ?string;
	def: string;

	constructor(repo: string, rev: ?string, def: string) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
	}
}

export class DefFetched {
	repo: string;
	rev: ?string;
	def: string;
	defObj: Def;

	constructor(repo: string, rev: ?string, def: string, defObj: Def) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
		this.defObj = defObj;
	}
}

export class WantDefAuthors {
	repo: string;
	rev: ?string;
	def: string;

	constructor(repo: string, rev: ?string, def: string) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
	}
}

export class DefAuthorsFetched {
	repo: string;
	rev: ?string;
	def: string;
	authors: Object;

	constructor(repo: string, rev: ?string, def: string, authors: Object) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
		this.authors = authors;
	}
}

export class WantDefs {
	repo: string;
	rev: ?string;
	query: string;
	filePathPrefix: ?string;
	overlay: boolean;

	constructor(repo: string, rev: ?string, query: string, filePathPrefix?: string, overlay: boolean) {
		this.repo = repo;
		this.rev = rev;
		this.query = query;
		this.filePathPrefix = filePathPrefix || null;
		this.overlay = overlay; // For metrics purposes
	}
}

export class DefsFetched {
	repo: string;
	rev: ?string;
	query: string;
	defs: Array<Def>;
	filePathPrefix: ?string;
	overlay: boolean;

	constructor(repo: string, rev: ?string, query: string, filePathPrefix: ?string, defs: Array<Def>, overlay: boolean) {
		this.repo = repo;
		this.rev = rev;
		this.query = query;
		this.filePathPrefix = filePathPrefix;
		this.defs = defs;
		this.overlay = overlay;
	}
}

export class SelectDef {
	repo: string;
	rev: ?string;
	def: string;
	eventName: string;

	constructor(repo: string, rev: ?string, def: string) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
		this.eventName = "SelectDef";
	}
}

export class HighlightDef {
	url: ?string;
	eventName: string;

	constructor(url: ?string) {
		this.url = url;
		this.eventName = "HighlightDef";
	}
}

export class WantRefLocations {
	resource: RefLocationsKey;
	perPage: ?number;
	page: ?number;

	constructor(r: RefLocationsKey, p?: {perPage?: number, page?: number}) {
		this.resource = r;
		if (p) {
			this.perPage = p.perPage;
			this.page = p.page;
		}
	}

	url(): string {
		let q = toQuery({
			ReposOnly: this.resource.reposOnly,
			Query: this.resource.repos,
			Page: this.page,
			PerPage: this.perPage,
		});
		if (q) {
			q = `?${q}`;
		}
		return `/.api/repos/${this.resource.repo}${this.resource.rev ? `@${this.resource.rev}` : ""}/-/def/${this.resource.def}/-/ref-locations${q}`;
	}
}

export class RefLocationsFetched {
	request: WantRefLocations;
	locations: Array<Object>;

	constructor(request: WantRefLocations, locations: Array<Object>) {
		this.request = request;
		this.locations = locations;
	}
}

export class WantRefs {
	repo: string;
	rev: ?string;
	def: string;
	refRepo: string; // return refs from files in this repo
	refFile: ?string; // only return refs in this file

	constructor(repo: string, rev: ?string, def: string, refRepo: string, refFile: ?string) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
		this.refRepo = refRepo;
		this.refFile = refFile || null;
	}
}

export class RefsFetched {
	repo: string;
	rev: ?string;
	def: string;
	refRepo: string;
	refFile: ?string;
	refs: Array<Ref>;

	constructor(repo: string, rev: ?string, def: string, refRepo: string, refFile: ?string, refs: Array<Ref>) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
		this.refRepo = refRepo;
		this.refFile = refFile || null;
		this.refs = refs;
	}
}
