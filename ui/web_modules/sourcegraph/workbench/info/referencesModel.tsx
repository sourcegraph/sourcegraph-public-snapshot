import { Repo } from "sourcegraph/api/index";

import Event, { fromEventEmitter } from "vs/base/common/event";
import { EventEmitter } from "vs/base/common/eventEmitter";
import { defaultGenerator } from "vs/base/common/idGenerator";
import { IDisposable, IReference, dispose } from "vs/base/common/lifecycle";
import { basename, dirname } from "vs/base/common/paths";
import * as strings from "vs/base/common/strings";
import URI from "vs/base/common/uri";
import { TPromise } from "vs/base/common/winjs.base";
import { Range } from "vs/editor/common/core/range";
import { IModel, IPosition, IRange } from "vs/editor/common/editorCommon";
import { Location } from "vs/editor/common/modes";
import { ITextEditorModel, ITextModelResolverService } from "vs/editor/common/services/resolverService";

import * as _ from "lodash";

import { LocationWithCommitInfo, ReferenceCommitInfo } from "sourcegraph/util/RefsBackend";

export class OneReference implements IDisposable {

	private _id: string;
	private _commitInfo: ReferenceCommitInfo;
	private _preview: FilePreview;
	private _resolved: boolean;

	constructor(
		private _parent: FileReferences,
		private _range: IRange,
		private _eventBus: EventEmitter
	) {
		this._id = defaultGenerator.nextId();
	}

	public get id(): string {
		return this._id;
	}

	public get model(): FileReferences {
		return this._parent;
	}

	public get preview(): FilePreview {
		return this._preview;
	}

	public set parent(value: FileReferences) {
		this._parent = value;
	}

	public get parent(): FileReferences {
		return this._parent;
	}

	public get uri(): URI {
		return this._parent.uri;
	}

	public get name(): string {
		return this._parent.name;
	}

	public get directory(): string {
		return this._parent.directory;
	}

	public get range(): IRange {
		return this._range;
	}

	public set range(value: IRange) {
		this._range = value;
		this._eventBus.emit("ref/changed", this);
	}

	public get commitInfo(): ReferenceCommitInfo {
		return this._commitInfo;
	}

	public set commitInfo(value: ReferenceCommitInfo) {
		this._commitInfo = value;
	};

	public async resolve(textModelResolverService: ITextModelResolverService): Promise<OneReference> {
		if (this._resolved) {
			return TPromise.as(this);
		}

		let modelReference = await textModelResolverService.createModelReference(this.uri);
		if (!modelReference) {
			// something went wrong... rutro.
			this._resolved = true;
			return this;
		}

		const model = modelReference.object;
		if (!model) {
			modelReference.dispose();
			throw new Error();
		}

		this._preview = new FilePreview(modelReference);
		this._resolved = true;
		return this;
	}

	dispose(): void {
		if (this._preview) {
			this._preview.dispose();
			this._preview = (null as any);
		}
	}
}

export class FilePreview implements IDisposable {

	constructor(private _modelReference: IReference<ITextEditorModel>) {

	}

	private get _model(): IModel { return this._modelReference.object.textEditorModel; }

	public preview(range: IRange, n: number = 8): { before: string; inside: string; after: string } {
		const { startLineNumber, startColumn, endColumn } = range;
		const word = this._model.getWordUntilPosition({ lineNumber: startLineNumber, column: startColumn - n });
		const beforeRange = new Range(startLineNumber, word.startColumn, startLineNumber, startColumn);
		const afterRange = new Range(startLineNumber, endColumn, startLineNumber, Number.MAX_VALUE);

		const ret = {
			before: this._model.getValueInRange(beforeRange).replace(/^\s+/, strings.empty),
			inside: this._model.getValueInRange(range),
			after: this._model.getValueInRange(afterRange).replace(/\s+$/, strings.empty)
		};

		return ret;
	}

	dispose(): void {
		if (this._modelReference) {
			this._modelReference.dispose();
			this._modelReference = (null as any);
		}
	}
}

export class FileReferences implements IDisposable {

	private _children: OneReference[];
	private _preview: FilePreview;
	private _resolved: boolean;
	private _loadFailure: any;

	constructor(private _parent: ReferencesModel, private _uri: URI, private _range: IRange) {
		this._children = [];
	}

	public get id(): string {
		return `${this._uri.toString()}:${this._range.startLineNumber}${this._range.startColumn}`;
	}

	public get parent(): ReferencesModel {
		return this._parent;
	}

	public get range(): IRange {
		return this._range;
	}

	public set children(value: OneReference[]) {
		this._children = value;
	}

	public get children(): OneReference[] {
		return this._children;
	}

	public get uri(): URI {
		return this._uri;
	}

	public get name(): string {
		return basename(this.uri.fsPath);
	}

	public get directory(): string {
		return dirname(this.uri.fsPath);
	}

	public get preview(): FilePreview {
		return this._preview;
	}

	public get failure(): any {
		return this._loadFailure;
	}

	public async resolve(textModelResolverService: ITextModelResolverService): TPromise<FileReferences> {
		if (this._resolved) {
			return TPromise.as(this);
		}

		let modelReference = await textModelResolverService.createModelReference(this._uri);
		if (!modelReference) {
			// something wrong here
			this._children = [];
			this._resolved = true;
			this._loadFailure = new Error();
			return this;
		}

		const model = modelReference.object;
		if (!model) {
			modelReference.dispose();
			throw new Error();
		}

		this._preview = new FilePreview(modelReference);
		this._resolved = true;

		for (const oneRef of this._children) {
			await oneRef.resolve(textModelResolverService);
		}

		return this;
	}

	dispose(): void {
		if (this._preview) {
			this._preview.dispose();
			this._preview = (null as any);
		}
	}
}

export class ReferencesModel implements IDisposable {

	private _groups: FileReferences[] = [];
	private _references: OneReference[] = [];
	private _eventBus: EventEmitter = new EventEmitter();

	onDidChangeReferenceRange: Event<OneReference> = fromEventEmitter<OneReference>(this._eventBus, "ref/changed");

	constructor(references: LocationWithCommitInfo[], private _workspace: URI, private _tempFileReferences?: [Repo]) {
		let newArrayOfLocs: FileReferences[] = [];
		if (this._tempFileReferences) {
			newArrayOfLocs = _.flatten(this._tempFileReferences.map(repository => {
				let loc: Location = {
					uri: URI.from({
						scheme: this._workspace.scheme,
						authority: this._workspace.authority,
						path: (repository as any).URI.replace("github.com", ""),
						fragment: "",
						query: repository.DefaultBranch,
					}),
					range: {
						startLineNumber: 0,
						startColumn: 0,
						endColumn: 0,
						endLineNumber: 0,
					},
				};

				return new FileReferences(this, loc.uri, loc.range);
			}));
		}

		// grouping and sorting
		references.sort(ReferencesModel._compareReferences);

		let current: FileReferences | null = null;
		// Make the real groups again.
		let realGroups: FileReferences[] = [];
		for (let ref of references) {
			// We have a new repo! YAY!
			if (!current || current.uri.path !== ref.uri.path) {
				let temp = new FileReferences(this, ref.uri, ref.range);
				realGroups.push(temp);
			}
			// Make the correct file reference and generate a real preview!
			current = new FileReferences(this, ref.uri, ref.range);
			this.groups.push(current);

			// append, check for equality first!
			if (current.children.length === 0
				|| !Range.equalsRange(ref.range, current.children[current.children.length - 1].range)) {
				let oneRef = new OneReference(current, ref.range, this._eventBus);
				this._references.push(oneRef);
				current.children.push(oneRef);
			}
		}

		let arrayOfChildren: OneReference[] = [];
		for (let group of this._groups) {
			group.children.forEach(child => {
				arrayOfChildren.push(child);
			});
		}

		this._groups = [];
		for (let group of realGroups) {
			let tempGroup: OneReference[] = [];
			for (let reference of arrayOfChildren) {
				if (reference.uri.path === group.uri.path) {
					tempGroup.push(reference);
				}
			}

			group.children = tempGroup;
			if (this._workspace && group.uri.path === this._workspace.path) {
				this._groups.splice(0, 0, group);
			} else {
				this._groups.push(group);
			}
		}

		this._groups = this._groups.concat(newArrayOfLocs);

		for (let i = 0; i < references.length; ++i) {
			const commitInfo = references[i].commitInfo;
			if (commitInfo) {
				this.references[i].commitInfo = commitInfo;
			}
		}
	}

	public get empty(): boolean {
		return this._groups.length === 0;
	}

	public get references(): OneReference[] {
		return this._references;
	}

	public set groups(value: FileReferences[]) {
		this._groups = value;
	}

	public get groups(): FileReferences[] {
		return this._groups;
	}

	public nextReference(reference: OneReference): OneReference {

		let idx = reference.parent.children.indexOf(reference);
		let len = reference.parent.children.length;
		let totalLength = reference.parent.parent.groups.length;

		if (idx + 1 < len || totalLength === 1) {
			return reference.parent.children[(idx + 1) % len];
		}

		idx = reference.parent.parent.groups.indexOf(reference.parent);
		idx = (idx + 1) % totalLength;

		return reference.parent.parent.groups[idx].children[0];
	}

	public nearestReference(resource: URI, position: IPosition): OneReference | any {

		const nearest = this._references.map((ref, idx) => {
			return {
				idx,
				prefixLen: strings.commonPrefixLength(ref.uri.toString(), resource.toString()),
				offsetDist: Math.abs(ref.range.startLineNumber - position.lineNumber) * 100 + Math.abs(ref.range.startColumn - position.column)
			};
		}).sort((a, b) => {
			if (a.prefixLen > b.prefixLen) {
				return -1;
			} else if (a.prefixLen < b.prefixLen) {
				return 1;
			} else if (a.offsetDist < b.offsetDist) {
				return -1;
			} else if (a.offsetDist > b.offsetDist) {
				return 1;
			} else {
				return 0;
			}
		})[0];

		if (nearest) {
			return this._references[nearest.idx];
		}
	}

	dispose(): void {
		this._groups = dispose(this._groups);
	}

	private static _compareReferences(a: Location, b: Location): number {
		if (a.uri.toString() < b.uri.toString()) {
			return -1;
		} else if (a.uri.toString() > b.uri.toString()) {
			return 1;
		} else {
			return Range.compareRangesUsingStarts(a.range, b.range);
		}
	}
}
