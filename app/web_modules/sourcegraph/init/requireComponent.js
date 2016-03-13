const _requireComponentContext = require.context("sourcegraph/", true, /\.\/((?!testdata\/)[^/]+\/)*((?![^/.]+_test)[^/.])+$/);

// requireComponent requires the component exported (as default) from the named
// module (e.g., "sourcegraph/search/SearchBar").
export default function(module) {
	if (!module.startsWith("sourcegraph/")) {
		throw new Error(`Unable to load module '${module}' because only 'sourcegraph/**/*' modules may be statically resolved.`);
	}
	module = module.replace(/^sourcegraph\//, "");
	let mod = _requireComponentContext(`./${module}`);
	if (!mod.default) {
		throw new Error(`Unable to load module '${module}' because it does not have a default export.`);
	}
	return mod.default;
}
