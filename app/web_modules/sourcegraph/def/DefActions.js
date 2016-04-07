// @flow

import type {Def, Ref} from "sourcegraph/def";

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
	eventName: string;

	constructor(repo: string, rev: ?string, def: string, defObj: Def) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
		this.defObj = defObj;
		this.eventName = "DefFetched";
	}
}

export class WantDefs {
	repo: string;
	rev: ?string;
	query: string;

	constructor(repo: string, rev: ?string, query: string) {
		this.repo = repo;
		this.rev = rev;
		this.query = query;
	}
}

export class DefsFetched {
	repo: string;
	rev: ?string;
	query: string;
	defs: Array<Def>;
	eventName: string;

	constructor(repo: string, rev: ?string, query: string, defs: Array<Def>) {
		this.repo = repo;
		this.rev = rev;
		this.query = query;
		this.defs = defs;
		this.eventName = "DefsFetched";
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

export class WantRefs {
	repo: string;
	rev: ?string;
	def: string;
	file: ?string; // only return refs in this file

	constructor(repo: string, rev: ?string, def: string, file: ?string) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
		this.file = file || null;
	}
}

export class RefsFetched {
	repo: string;
	rev: ?string;
	def: string;
	file: ?string;
	refs: Array<Ref>;
	eventName: string;

	constructor(repo: string, rev: ?string, def: string, file: ?string, refs: Array<Ref>) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
		this.file = file || null;
		this.refs = refs;
		this.eventName = "RefsFetched";
	}
}
