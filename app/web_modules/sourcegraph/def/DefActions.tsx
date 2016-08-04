import {Def, ExamplesKey, Ref, RefLocationsKey, refLocsPerPage} from "sourcegraph/def/index";
import toQuery from "sourcegraph/util/toQuery";

export class WantDef {
	repo: string;
	rev: string | null;
	def: string;

	constructor(repo: string, rev: string | null, def: string) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
	}
}

export class DefFetched {
	repo: string;
	rev: string | null;
	def: string;
	defObj: Def;

	constructor(repo: string, rev: string | null, def: string, defObj: Def) {
		this.repo = repo;
		this.rev = rev;
		this.def = def;
		this.defObj = defObj;
	}
}

export class WantDefAuthors {
	repo: string;
	commitID: string;
	def: string;

	constructor(repo: string, commitID: string, def: string) {
		this.repo = repo;
		this.commitID = commitID;
		this.def = def;
	}
}

export class DefAuthorsFetched {
	repo: string;
	commitID: string;
	def: string;
	authors: Object;

	constructor(repo: string, commitID: string, def: string, authors: Object) {
		this.repo = repo;
		this.commitID = commitID;
		this.def = def;
		this.authors = authors;
	}
}

export class WantDefs {
	repo: string;
	commitID: string;
	query: string;
	filePathPrefix: string | null;
	overlay: boolean;

	constructor(repo: string, commitID: string, query: string, filePathPrefix: string | null, overlay: boolean) {
		this.repo = repo;
		this.commitID = commitID;
		this.query = query;
		this.filePathPrefix = filePathPrefix || null;
		this.overlay = overlay; // For metrics purposes
	}
}

export class DefsFetched {
	repo: string;
	commitID: string;
	query: string;
	defs: Array<Def>;
	filePathPrefix: string | null;
	overlay: boolean;

	constructor(repo: string, commitID: string, query: string, filePathPrefix: string | null, defs: Array<Def>, overlay: boolean) {
		this.repo = repo;
		this.commitID = commitID;
		this.query = query;
		this.filePathPrefix = filePathPrefix;
		this.defs = defs;
		this.overlay = overlay;
	}
}

export class SelectDef {
	repo: string;
	commitID: string;
	def: string;
	eventName: string;

	constructor(repo: string, commitID: string, def: string) {
		this.repo = repo;
		this.commitID = commitID;
		this.def = def;
		this.eventName = "SelectDef";
	}
}

export type BlobPos = {repo: string, commit: string, file: string, line: number, character: number};

export class Hovering {
	pos: BlobPos | null;

	constructor(pos: BlobPos | null) {
		this.pos = pos;
	}
}

export class WantHoverInfo {
	pos: BlobPos | null;

	constructor(pos: BlobPos | null) {
		this.pos = pos;
	}
}

export class HoverInfoFetched {
	pos: BlobPos | null;
	info: any;

	constructor(pos: BlobPos | null, info: any) {
		this.pos = pos;
		this.info = info;
	}
}

export class WantRefLocations {
	resource: RefLocationsKey;

	constructor(r: RefLocationsKey) {
		this.resource = r;
	}

	url(): string {
		let q = toQuery({
			Repos: this.resource.repos,
			Page: this.resource.page,
			PerPage: refLocsPerPage,
		});
		if (q) {
			q = `?${q}`;
		}
		return `/.api/repos/${this.resource.repo}${this.resource.commitID ? `@${this.resource.commitID}` : ""}/-/def/${this.resource.def}/-/ref-locations${q}`;
	}
}

export class RefLocationsFetched {
	request: WantRefLocations;
	locations: Object;

	constructor(request: WantRefLocations, locations: Object) {
		this.request = request;
		this.locations = locations;
	}
}

export class WantExamples {
	resource: ExamplesKey;

	constructor(r: ExamplesKey) {
		this.resource = r;
	}

	url(): string {
		return `/.api/repos/${this.resource.repo}${this.resource.commitID ? `@${this.resource.commitID}` : ""}/-/def/${this.resource.def}/-/examples`;
	}
}

export class ExamplesFetched {
	request: WantExamples;
	locations: Object;

	constructor(request: WantExamples, locations: Object) {
		this.request = request;
		this.locations = locations;
	}
}

export class WantRefs {
	repo: string;
	commitID: string;
	def: string;
	refRepo: string; // return refs from files in this repo
	refFile: string | null; // only return refs in this file

	constructor(repo: string, commitID: string, def: string, refRepo: string, refFile: string | null) {
		this.repo = repo;
		this.commitID = commitID;
		this.def = def;
		this.refRepo = refRepo;
		this.refFile = refFile || null;
	}
}

export class RefsFetched {
	repo: string;
	commitID: string;
	def: string;
	refRepo: string;
	refFile: string | null;
	refs: Array<Ref>;

	constructor(repo: string, commitID: string, def: string, refRepo: string, refFile: string | null, refs: Array<Ref>) {
		this.repo = repo;
		this.commitID = commitID;
		this.def = def;
		this.refRepo = refRepo;
		this.refFile = refFile || null;
		this.refs = refs;
	}
}
