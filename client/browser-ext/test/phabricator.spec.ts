import { expect } from "chai";
import "fetch-mock";
import * as jsdom from "jsdom";
import * as utils from "../app/utils";
import * as annotations from "../app/utils/annotations";
import * as phabricator from "../app/utils/phabricator";
import { CodeCell, PhabDifferentialUrl, PhabDiffusionUrl, PhabRevisionUrl, PhabricatorMode } from "../app/utils/types";

function setupDOM(url: string): (done: any) => void {
	return (done) => jsdom.env(url, (err, window) => {
		if (err) {
			done(err);
		}
		global.window = window;
		global.document = window.document;
		global.navigator = window.navigator;
		global.Node = (window as any).Node;
		done();
	});
}

describe("Phabricator DOM", () => {
	describe("diffusion view", () => {
		const url = "https://secure.phabricator.com/diffusion/ARC/browse/master/src/__phutil_library_init__.php";
		phabricator.setPhabricatorInstance(phabricator.securePhabricatorInstance);
		before(setupDOM(url));

		it("diffusion state variables", () => {
			const phabUrlState = phabricator.getPhabricatorState({ href: url } as Location);
			expect(phabUrlState).to.be.not.null;
			expect(phabUrlState!.mode).to.be.eql(PhabricatorMode.Diffusion);
			const phabDiffusionUrlState = phabUrlState as PhabDiffusionUrl;
			expect(phabDiffusionUrlState.branch).to.be.eql("master");
			expect(phabDiffusionUrlState.path).to.be.eql("src/__phutil_library_init__.php");
			expect(phabDiffusionUrlState.rev.length).to.be.eql(40);
		});
	});

	describe("differential view", () => {
		const url = "https://secure.phabricator.com/rARCc13e5a629535f460ca1a16d0bfe6d95f43b70b78";
		phabricator.setPhabricatorInstance(phabricator.securePhabricatorInstance);
		before(setupDOM(url));

		it("diffusion state variables", () => {
			const phabUrlState = phabricator.getPhabricatorState({ href: url } as Location);
			expect(phabUrlState).to.be.not.null;
			expect(phabUrlState!.mode).to.be.eql(PhabricatorMode.Revision);
			const phabDifferentialUrlState = phabUrlState as PhabRevisionUrl;
			expect(phabDifferentialUrlState.childRev).to.be.eql("c13e5a629535f460ca1a16d0bfe6d95f43b70b78");
			expect(phabDifferentialUrlState.parentRev).to.be.eql("c75b671b221add586eb4e61ff453357bbe140e8e");
			expect(phabDifferentialUrlState.repoUri).to.be.eql("github.com/phacility/arcanist");
		});
		it("the right stuff", () => {
			const files = document.getElementsByClassName("differential-changeset") as HTMLCollectionOf<HTMLElement>;
			expect(files.length).to.be.eql(3);
		});
	});
});
