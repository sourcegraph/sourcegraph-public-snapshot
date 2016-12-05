import * as tsfmt from "typescript-formatter";

tsfmt
.processFiles(["./index.ts"], {
	dryRun: true,
replace: false,
verify: false,
  tsconfig: true,
	tslint: true,
editorconfig: true,
	tsfmt: true
})
	.then(result =>{
		console.log(result["./index.ts"].dest);
});
