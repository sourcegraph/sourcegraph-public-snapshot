import {Def, Ref} from "sourcegraph/api";
import {ExamplesKey, RefLocationsKey} from "sourcegraph/def";

export type Action =
	WantDef |
	DefFetched |
	WantDefAuthors |
	DefAuthorsFetched |
	WantDefs |
	DefsFetched |
	SelectDef |
	WantJumpDef |
	JumpDefFetched |
	Hovering |
	WantHoverInfo |
	HoverInfoFetched |
	WantRefLocations |
	RefLocationsFetched |
	WantLocalRefLocations |
	LocalRefLocationsFetched |
	WantExamples |
	ExamplesFetched |
	WantRefs |
	RefsFetched;

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
	authors: any;

	constructor(repo: string, commitID: string, def: string, authors: any) {
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

	constructor(repo: string, commitID: string, query: string, filePathPrefix: string | null, overlay?: boolean) {
		this.repo = repo;
		this.commitID = commitID;
		this.query = query;
		this.filePathPrefix = filePathPrefix || null;
		this.overlay = overlay || false; // For metrics purposes
	}
}

export class DefsFetched {
	repo: string;
	commitID: string;
	query: string;
	defs: Def[];
	filePathPrefix: string | null;
	overlay: boolean;

	constructor(repo: string, commitID: string, query: string, filePathPrefix: string | null, defs: Def[], overlay?: boolean) {
		this.repo = repo;
		this.commitID = commitID;
		this.query = query;
		this.filePathPrefix = filePathPrefix;
		this.defs = defs;
		this.overlay = overlay || false;
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

export interface BlobPos {
	repo: string;
	commit: string;
	file: string;
	line: number;
	character: number;
};

export class WantJumpDef {
	pos: BlobPos;

	constructor(pos: BlobPos) {
		this.pos = pos;
	}
}

export class JumpDefFetched {
	pos: BlobPos;
	def: any;

	constructor(pos: BlobPos, def: any) {
		this.pos = pos;
		this.def = def;
	}
}

export class Hovering {
	pos: BlobPos | null;

	constructor(pos: BlobPos | null) {
		this.pos = pos;
	}
}

export class WantHoverInfo {
	pos: BlobPos;

	constructor(pos: BlobPos) {
		this.pos = pos;
	}
}

export class HoverInfoFetched {
	pos: BlobPos;
	info: any;

	constructor(pos: BlobPos, info: any) {
		this.pos = pos;
		this.info = info;
	}
}

export class WantRefLocations {
	resource: RefLocationsKey;

	constructor(r: RefLocationsKey) {
		this.resource = r;
	}
}

export class RefLocationsFetched {
	request: WantRefLocations;
	locations: any;

	constructor(request: WantRefLocations, locations: any) {
		this.request = request;
		this.locations = locations;
	}
}

export class WantLocalRefLocations {
	pos: BlobPos;
	resource: RefLocationsKey;

	// TODO: should only use BlobPos after we completely discarded the old way of showing refLocations.
	constructor(pos: BlobPos, r: RefLocationsKey) {
		this.pos = pos;
		this.resource = r;
	}
}

export class LocalRefLocationsFetched {
	request: WantLocalRefLocations;
	locations: Object;

	constructor(request: WantLocalRefLocations, locations: Object) {
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
	locations: any;

	constructor(request: WantExamples, locations: any) {
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

	constructor(repo: string, commitID: string, def: string, refRepo: string, refFile?: string | null) {
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
	refs: Ref[] | null;

	constructor(repo: string, commitID: string, def: string, refRepo: string, refFile: string | null, refs: Ref[] | null) {
		this.repo = repo;
		this.commitID = commitID;
		this.def = def;
		this.refRepo = refRepo;
		this.refFile = refFile || null;
		this.refs = refs;
	}
}
