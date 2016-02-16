import autotest from "sourcegraph/util/autotest";

import React from "react";

import FileDiff from "sourcegraph/delta/FileDiff";

import testdataAdded from "sourcegraph/delta/testdata/FileDiff-added.json";
import testdataChanged from "sourcegraph/delta/testdata/FileDiff-changed.json";
import testdataRenamed from "sourcegraph/delta/testdata/FileDiff-renamed.json";
import testdataDeleted from "sourcegraph/delta/testdata/FileDiff-deleted.json";

const sampleStats = {Added: 5, Changed: 6, Deleted: 7};

const sampleAdded = {
	OrigName: "/dev/null",
	NewName: "b",
	Hunks: [{Body: "a\nb"}],
	Stats: sampleStats,
};
const sampleChanged = {
	OrigName: "a",
	NewName: "b",
	Hunks: [{Body: "a\nb"}, {Body: "a\nb"}],
	Stats: sampleStats,
};
const sampleRenamed = {
	OrigName: "a",
	NewName: "b",
	Hunks: [{Body: "a\nb"}],
	Stats: sampleStats,
};
const sampleDeleted = {
	OrigName: "a",
	NewName: "/dev/null",
	Hunks: [{Body: "a\nb"}],
	Stats: sampleStats,
};

describe("FileDiff", () => {
	it("should render added file", () => {
		autotest(testdataAdded, `${__dirname}/testdata/FileDiff-added.json`,
			<FileDiff diff={sampleAdded} baseRepo="br" baseRev="bv" headRepo="hr" headRev="hv" id="myid" annotations={{get() { return null; }}} />
		);
	});

	it("should render changed file", () => {
		autotest(testdataChanged, `${__dirname}/testdata/FileDiff-changed.json`,
			<FileDiff diff={sampleChanged} baseRepo="br" baseRev="bv" headRepo="hr" headRev="hv" id="myid" annotations={{get() { return null; }}} />
		);
	});


	it("should render renamed file", () => {
		autotest(testdataRenamed, `${__dirname}/testdata/FileDiff-renamed.json`,
			<FileDiff diff={sampleRenamed} baseRepo="br" baseRev="bv" headRepo="hr" headRev="hv" id="myid" annotations={{get() { return null; }}} />
		);
	});


	it("should render deleted file", () => {
		autotest(testdataDeleted, `${__dirname}/testdata/FileDiff-deleted.json`,
			<FileDiff diff={sampleDeleted} baseRepo="br" baseRev="bv" headRepo="hr" headRev="hv" id="myid" annotations={{get() { return null; }}} />
		);
	});
});
