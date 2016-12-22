/**
 * @module yarnshim provides a programmatic interface to the
 * buildhandler to call yarn. It is concurrency-safe (its functions
 * can be called from multiple processes referencing the same global
 * yarn directory and/or node_modules directories).
 */

import * as path from 'path';
import * as mkdirp from 'mkdirp';

import { FileSystem } from 'javascript-typescript-langserver/src/fs';
import { readFile } from './vfs';


const EventReporter = require('yarn/lib/reporters').EventReporter;
const Config = require('yarn/lib/config').default;
const Install = require('yarn/lib/cli/commands/install').Install;
const Lockfile = require('yarn/lib/lockfile/wrapper').default;
const normalizeManifest = require('yarn/lib/util/normalize-manifest').default;
const PackageRequest = require('yarn/lib/package-request').default;
const parsePackageName = require('yarn/lib/util/parse-package-name').default;
const registries = require('yarn/lib/registries').registries;

const semver = require('semver');
const lockfile = require('proper-lockfile');

/*
 * info mimics `yarn info` to return metadata about the specified package.
 */
export async function info(cwd: string, globaldir: string, overlaydir: string, packageName: string): Promise<Info> {
	const reporter = new EventReporter({
		emoji: false,
		verbose: false,
		noProgress: true,
	});
	const config = new Config(reporter);
	await config.init({
		cwd: cwd,
		globalFolder: path.join(globaldir, ".config"),
		linkFolder: path.join(globaldir, ".config", "link"),
		cacheFolder: path.join(globaldir, ".cache"),
		modulesFolder: path.join(overlaydir, "node_modules"),
		ignoreScripts: true,
	});

	// Handle the case when we are referencing a local package.
	if (packageName === '.') {
		packageName = (await config.readRootManifest()).name;
	}

	const packageInput = registries.npm.escapeName(packageName);
	const {name, version} = parsePackageName(packageInput);

	let result = await config.registries.npm.request(name);
	if (!result) {
		reporter.error(reporter.lang('infoFail'));
		return Promise.reject("failed to fetch metadata from NPM for package " + packageName);
	}

	result = clean(result);

	const versions = result.versions;
	result.versions = Object.keys(versions).sort(semver.compareLoose);
	result.version = version || result.versions[result.versions.length - 1];
	result = Object.assign(result, versions[result.version]);

	return result;
}

/*
 * infoAlt uses an alternative method of fetching package info in the
 * case that the standard info fails. This occurs, for example, when a
 * the package is a direct git dependency that doesn't exist in npm.
 */
export async function infoAlt(remoteFs: FileSystem, cwd: string, globaldir: string, overlaydir: string, packageName: string): Promise<Info> {
	const reporter = new EventReporter({
		emoji: false,
		verbose: false,
		noProgress: true,
	});
	const config = new Config(reporter);
	await config.init({
		cwd: cwd,
		globalFolder: path.join(globaldir, ".config"),
		linkFolder: path.join(globaldir, ".config", "link"),
		cacheFolder: path.join(globaldir, ".cache"),
		modulesFolder: path.join(overlaydir, "node_modules"),
		ignoreScripts: true,
	});
	const lf = await Lockfile.fromDirectory(config.cwd, reporter);
	const inst = new Install({}, config, reporter, lf);

	const { patterns: rawPatterns } = await fetchRequestFromRemoteFS(inst, [], remoteFs, overlaydir);

	for (const pattern of rawPatterns) {
		const pkginfo = parsePackageName(pattern);
		if (pkginfo.name === packageName) {
			const ghInfo = parseGitHubInfo(pkginfo.version);
			if (ghInfo) {
				return ghInfo;
			}
		}
	}

	return Promise.reject("could not resolve package," + packageName + ",through alternative means");
}

const ghURLParser = /^(?:https:\/\/|git\+https:\/\/|git:\/\/)(?:www\.)?(github\.com(?:\/[^\/#]+){2})(?:#([^\s]+))?$/;

export function parseGitHubInfo(cloneURL: string): Info | null {
	const matchInfo = cloneURL.match(ghURLParser);
	if (!matchInfo) {
		return null;
	}
	let match = matchInfo[0];
	if (!match) {
		return null;
	}
	let repoURI = matchInfo[1];
	let version = matchInfo[2];
	if (repoURI.endsWith(".git")) {
		repoURI = repoURI.substring(0, repoURI.length - ".git".length);
	}
	return {
		repository: {
			type: "git",
			url: "git://" + repoURI,
		},
		gitHead: version,
	};
}

/*
 * install mimics `yarn install --ignore-scripts`, installing
 * dependencies into a temporary directory on disk. cwd should specify
 * the directory in remoteFs from which the package.json should be
 * read.
 */
export async function install(remoteFs: FileSystem, cwd: string, globaldir: string, overlaydir: string): Promise<void> {
	await new Promise<void>((resolve, reject) => {
		mkdirp(overlaydir, (err) => {
			if (err) {
				return reject(err);
			} else {
				return resolve();
			}
		});
	});

	const reporter = new EventReporter({
		emoji: false,
		verbose: false,
		noProgress: true,
	});
	const config = new Config(reporter);
	await config.init({
		cwd: cwd,
		globalFolder: path.join(globaldir, ".config"),
		linkFolder: path.join(globaldir, ".config", "link"),
		cacheFolder: path.join(globaldir, ".cache"),
		modulesFolder: path.join(overlaydir, "node_modules"),
		ignoreScripts: true,
	});
	const globalLockfile = path.join(globaldir, ".lock");
	const overlayLockfile = path.join(overlaydir, ".lock");
	const lf = await Lockfile.fromDirectory(config.cwd, reporter);
	const inst = new Install({}, config, reporter, lf);

	const {
		requests: depRequests,
		ignorePatterns,
	} = await fetchRequestFromRemoteFS(inst, [], remoteFs, overlaydir);

	// filter out packages that are covered by @types/* packages
	const prunedDepRequests = [];
	{
		const typeDepNames = new Set<string>();
		for (const dep of depRequests) {
			const pkg = parsePackageName(dep.pattern);
			if (pkg.name.startsWith("@types/")) {
				typeDepNames.add(pkg.name.substr("@types/".length));
			}
		}
		for (const dep of depRequests) {
			const pkg = parsePackageName(dep.pattern);
			if (pkg.name.startsWith("@types/") || !typeDepNames.has(pkg.name)) {
				prunedDepRequests.push(dep);
			}
		}
	}

	// resolve
	const resolveStart = new Date().getTime();
	const deps: DependencyRequestPattern[] = inst.prepareRequests(prunedDepRequests);
	inst.resolver.flat = inst.flags.flat;
	const resolvedPatterns: string[] = [];

	await Promise.all(deps.map(async (req): Promise<void> => {
		try {
			await inst.resolver.find(req);
			resolvedPatterns.push(req.pattern);
		} catch (e) {
			console.error("warning: could not resolve dep: ", req);
		}
	}));

	// Note: if `--flat` is set, a yarn install will try to flatten
	// the patterns here via `inst.flatten`. We do not do so here,
	// because it may require manual conflict resolution and it also
	// attempts to re-read the manifest from disk at config.cwd (which
	// will fail, the files only exist in the VFS, not on local disk).
	const patterns = resolvedPatterns;

	const resolveEnd = new Date().getTime();
	console.error("resolve", patterns.length, (resolveEnd - resolveStart) / 1000.0);

	// fetch
	await runWithLockfile(globalLockfile, async () => {
		const fetchStart = new Date().getTime();
		inst.markIgnored(ignorePatterns);
		await inst.fetcher.init();
		const fetchEnd = new Date().getTime();
		console.error("fetch", resolvedPatterns.length, (fetchEnd - fetchStart) / 1000.0)
	});

	// link
	await runWithLockfile(overlayLockfile, async () => {
		const linkStart = new Date().getTime();
		inst.linker.resolvePeerModules();
		await inst.linker.copyModules(patterns);
		const linkEnd = new Date().getTime();
		console.error("link", patterns.length, (linkEnd - linkStart) / 1000.0)
	});

	return Promise.resolve();
}

async function runWithLockfile(lf: string, run: () => Promise<void>): Promise<void> {
	await new Promise<void>((resolve, reject) => {
		lockfile.lock(lf, { realpath: false }, async (err: any, release: () => void) => {
			try {
				if (err) {
					await new Promise<void>((resolve, reject) => {
						setTimeout(() => {
							return resolve();
						}, 200);
					});
					await runWithLockfile(lf, run);
				} else {
					await run();
					release();
				}
			} catch (e) {
				release();
				return reject(e);
			}
			return resolve();
		});
	});
}

/*
 * clean is a non-exported function copied over from src/cli/commands/info.js in the yarn repository.
 */
function clean(object: { [key: string]: any }): any {
	if (Array.isArray(object)) {
		const result: { [key: string]: any }[] = [];
		object.forEach((item) => {
			item = clean(item);
			if (item) {
				result.push(item);
			}
		});
		return result;
	} else if (typeof object === 'object') {
		const result: { [key: string]: any } = {};
		for (const key in object) {
			if (key.startsWith('_')) {
				continue;
			}

			const item = clean(object[key]);
			if (item) {
				result[key] = item;
			}
		}
		return result;
	} else if (object) {
		return object;
	} else {
		return null;
	}
}

// fetchRequestFromRemoteFS replicates functionality of
// yarn.cli.commands.Install.fetchRequestFromCwd using the VFS instead
// of the current working directory of the local filesystem.
async function fetchRequestFromRemoteFS(inst: Install, excludePatterns: string[] = [], fs: FileSystem, overlaydir: string): Promise<InstallCwdRequest> {
	const patterns: string[] = [];
	const deps: any[] = [];
	const manifest: Manifest = {};

	const ignorePatterns: string[] = [];
	const usedPatterns: string[] = [];

	// exclude package names that are in install args
	const excludeNames: string[] = [];
	for (const pattern of excludePatterns) {
		// can't extract a package name from inst
		if (PackageRequest.getExoticResolver(pattern)) {
			continue;
		}

		// extract the name
		const parts = PackageRequest.normalizePattern(pattern);
		excludeNames.push(parts.name);
	}

	for (const registry of Object.keys(registries)) {
		const {filename} = registries[registry];
		const loc = path.join(inst.config.cwd, filename);

		let jsonRaw: string;
		try {
			jsonRaw = await readFile(fs, loc)
		} catch (e) {
			continue;
		}

		inst.rootManifestRegistries.push(registry);
		const json = JSON.parse(jsonRaw);
		await normalizeManifest(json, overlaydir, inst.config, true);

		Object.assign(inst.resolutions, json.resolutions);
		Object.assign(manifest, json);

		const pushDeps = (depType: string, {hint, optional}: { hint: string | null, optional: boolean }, isUsed: boolean) => {
			const depMap = json[depType];
			for (const name in depMap) {
				if (excludeNames.indexOf(name) >= 0) {
					continue;
				}

				let pattern = name;
				if (!inst.lockfile.getLocked(pattern, true)) {
					// when we use --save we save the dependency to the lockfile with just the name rather than the
					// version combo
					pattern += '@' + depMap[name];
				}

				if (isUsed) {
					usedPatterns.push(pattern);
				} else {
					ignorePatterns.push(pattern);
				}

				inst.rootPatternsToOrigin[pattern] = depType;
				patterns.push(pattern);
				deps.push({ pattern, registry, hint, optional });
			}
		};

		pushDeps('dependencies', { hint: null, optional: false }, true);
		pushDeps('devDependencies', { hint: 'dev', optional: false }, !inst.config.production);
		pushDeps('optionalDependencies', { hint: 'optional', optional: true }, !inst.flags.ignoreOptional);

		break;
	}

	// inherit root flat flag
	if (manifest.flat) {
		inst.flags.flat = true;
	}

	return {
		requests: deps,
		patterns,
		manifest,
		usedPatterns,
		ignorePatterns,
	};
}

/*
 * The following types mirror those defined in Flow in the Yarn repository.
 */

export interface Info {
	repository?: {
		type: string;
		url: string;
	};
	gitHead?: string;
}


interface Config {
	cwd: string;
	production?: boolean;
}

interface PackageFetcher {
	init: any;
}

interface PackageResolver {
	init: any;
}

interface Install {
	lockfile: {
		getLocked: any;
	};
	flags: Flags;
	rootPatternsToOrigin: { [key: string]: string; };
	rootManifestRegistries: any[];
	config: Config;
	resolutions: any;

	fetcher: PackageFetcher;
	resolver: PackageResolver;

	init: () => Promise<Array<string>>;
}

interface Flags {
	// install
	ignorePlatform: boolean;
	ignoreEngines: boolean;
	ignoreScripts: boolean;
	ignoreOptional: boolean;
	har: boolean;
	force: boolean;
	flat: boolean;
	lockfile: boolean;
	pureLockfile: boolean;
	skipIntegrity: boolean;

	// add
	peer: boolean;
	dev: boolean;
	optional: boolean;
	exact: boolean;
	tilde: boolean;
}

interface InstallCwdRequest {
	requests: any;
	patterns: Array<string>;
	ignorePatterns: Array<string>;
	usedPatterns: Array<string>;
	manifest: Object;
}

interface Manifest {
	flat?: boolean;
}

interface DependencyRequestPattern {
	pattern: string;
	registry: any;
	optional: boolean;
	hint?: string | null;
	parentRequest?: any;
}
