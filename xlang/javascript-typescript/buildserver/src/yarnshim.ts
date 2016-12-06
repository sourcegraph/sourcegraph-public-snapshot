import * as path from 'path';
import * as fs from 'fs';
import * as mkdirp from 'mkdirp';

import { FileSystem, FileInfo } from 'javascript-typescript-langserver/src/fs';
import { readFile, readDir } from './vfs';


const ConsoleReporter = require('yarn/lib/reporters').ConsoleReporter;
const Config = require('yarn/lib/config').default;
const Install = require('yarn/lib/cli/commands/install').Install;
const Lockfile = require('yarn/lib/lockfile/wrapper').default;
const normalizeManifest = require('yarn/lib/util/normalize-manifest').default;
const PackageRequest = require('yarn/lib/package-request').default;
const parsePackageName = require('yarn/lib/util/parse-package-name').default;
const registries = require('yarn/lib/registries').registries;
const parsePacakgeName = require('yarn/lib/util/parse-package-name').default;

const semver = require('semver');

/*
 * info mimics `yarn info` to return metadata about the specified package.
 */
export async function info(cwd: string, globaldir: string, overlaydir: string, packageName: string): Promise<Info> {
	const reporter = new ConsoleReporter({
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
		return;
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
	const reporter = new ConsoleReporter({
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
	const lockfile = await Lockfile.fromDirectory(config.cwd, reporter);
	const inst = new Install({}, config, reporter, lockfile);

	const {
		requests: depRequests,
		patterns: rawPatterns,
		ignorePatterns,
		usedPatterns,
	} = await fetchRequestFromRemoteFS(inst, [], remoteFs, overlaydir);

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

const ghURLParser = /^(?:https:\/\/|git\+https:\/\/|git:\/\/)(github\.com(?:\/[^\/#]+){2})(?:\.git)?(?:#([^\s]+))?$/;

function parseGitHubInfo(cloneURL: string): Info | null {
	const [match, repoURI, version] = cloneURL.match(ghURLParser);
	if (!match) {
		return null;
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

	const reporter = new ConsoleReporter({
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
	const lockfile = await Lockfile.fromDirectory(config.cwd, reporter);
	const inst = new Install({}, config, reporter, lockfile);

	const {
		requests: depRequests,
		patterns: rawPatterns,
		ignorePatterns,
		usedPatterns,
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
	const seedPatterns = deps.map((dep): string => dep.pattern);
	const resolvedPatterns = [];
	await Promise.all(deps.map(async (req): Promise<void> => {
		try {
			await inst.resolver.find(req);
			resolvedPatterns.push(req.pattern);
		} catch (e) {
			console.error("warning: could not resolve dep: ", req);
		}
	}));
	const patterns: any[] = await inst.flatten(inst.preparePatterns(resolvedPatterns));
	const resolveEnd = new Date().getTime();
	console.error("resolve", patterns.length, (resolveEnd - resolveStart) / 1000.0);

	// fetch
	const fetchStart = new Date().getTime();
	inst.markIgnored(ignorePatterns);
	await inst.fetcher.init();
	const fetchEnd = new Date().getTime();
	console.error("fetch", resolvedPatterns.length, (fetchEnd - fetchStart) / 1000.0)

	// link
	const linkStart = new Date().getTime();
	inst.linker.resolvePeerModules();
	await inst.linker.copyModules(patterns);
	const linkEnd = new Date().getTime();
	console.error("link", patterns.length, (linkEnd - linkStart) / 1000.0)

	return Promise.resolve();
}

/*
 * clean is a non-exported function copied over from src/cli/commands/info.js in the yarn repository.
 */
function clean(object: any): any {
	if (Array.isArray(object)) {
		const result = [];
		object.forEach((item) => {
			item = clean(item);
			if (item) {
				result.push(item);
			}
		});
		return result;
	} else if (typeof object === 'object') {
		const result = {};
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
	const patterns = [];
	const deps = [];
	const manifest: Manifest = {};

	const ignorePatterns = [];
	const usedPatterns = [];

	// exclude package names that are in install args
	const excludeNames = [];
	for (const pattern of excludePatterns) {
		// can't extract a package name from this
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

		const pushDeps = (depType, {hint, optional}, isUsed) => {
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
		this.flags.flat = true;
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
	flags: {
		ignoreOptional?: boolean;
	};
	rootPatternsToOrigin: { [key: string]: number; };
	rootManifestRegistries: any[];
	config: Config;
	resolutions: any;

	fetcher: PackageFetcher;
	resolver: PackageResolver;

	init: () => Promise<Array<string>>;
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
