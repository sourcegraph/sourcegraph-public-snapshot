import {expect} from "chai";
import * as annotations from "../../app/utils/annotations";

describe("annotations", () => {

	// const annsByStartByte = {
	// 	0: {URL: "url", StartByte: 0, EndByte: 3},
	// 	4: {URL: "url", StartByte: 4, EndByte: 9}
	// };

	// const offsetByLine = {
	// 	1: 0,
	// 	2: 4,
	// };

	// const text = "fmt"
	// const textNode = document.createTextNode(text);

	// const openingTag = `<span class="pd-s">`;
	// const elementNode = document.createElement("span");
	// elementNode.appendChild(textNode);
	// elementNode.className = "pd-s";

	// const quotedStringNode = document.createElement("span");
	// const quoteNode = document.createElement("span");
	// quoteNode.className = ".pl-pds";
	// quotedStringNode.appendChild(quoteNode);
	// quotedStringNode.appendChild(textNode);
	// quotedStringNode.appendChild(quoteNode);

	// const codeCell1 = document.createElement("td");
	// codeCell1.appendChild(elementNode);
	// const codeCell2 = document.createElement("td");
	// codeCell2.appendChild(quotedStringNode);

	// describe("getOpeningTag", () => {
	// 	it("should return for tag with attribute", () => {
	// 		expect(annotations.getOpeningTag(elementNode)).to.eql(openingTag);
	// 	});
	// });

	// describe("convertTextNode", () => {
	// 	it.only("should convert text node with annotation", () => {
	// 		const {result, bytesConsumed} = annotations.convertTextNode(textNode, annsByStartByte, 0);
	// 		console.log(result)
	// 		expect(textNode).to.eql("fmt");
	// 		expect(result.startsWith("<a")).to.eql(true);
	// 		expect(result.contains(`>${text}</a>`)).to.eql(true);
	// 		expect(bytesConsumed).to.eql(3);
	// 	});

	// 	it("should convert text node without annotation", () => {
	// 		const {result, bytesConsumed} = annotations.convertTextNode(textNode, annsByStartByte, 1);
	// 		expect(result).to.eql(text);
	// 		expect(bytesConsumed).to.eql(3);
	// 	});
	// });

	// describe("convertElementNode", () => {
	// 	it("should convert element node with annotation", () => {
	// 		const {result, bytesConsumed} = annotations.convertElementNode(elementNode, annsByStartByte, 0);
	// 		expect(result.startsWith(openingTag)).to.eql(true);
	// 		expect(result.contains("<a")).to.eql(true);
	// 		expect(result.contains(`>${text}</a>`)).to.eql(true);
	// 		expect(result.endsWith("</span>")).to.eql(true);
	// 		expect(bytesConsumed).to.eql(3);
	// 	});

	// 	it("should convert element node without annotation", () => {
	// 		const {result, bytesConsumed} = annotations.convertElementNode(elementNode, annsByStartByte, 1);
	// 		expect(result).to.eql(`<span class="pd-s">${text}</span>`);
	// 		expect(bytesConsumed).to.eql(3);
	// 	});
	// });

	// describe("isQuotedStringNode", () => {
	// 	it("should return false for element node", () => {
	// 		expect(annotations.isQuotedStringNode(elementNode)).to.eql(false);
	// 	});

	// 	it("should return false for text node", () => {
	// 		expect(annotations.isQuotedStringNode(textNode)).to.eql(false);
	// 	});

	// 	it("should return true for quoted string node", () => {
	// 		expect(annotations.getOpeningTag(quotedStringNode)).to.eql(true);
	// 	});
	// });

	// describe("convertQuotedStringNode", () => {
	// 	const quotedStringInnerHTML = `<span><span class=".pl-pds">"</span>${text}<span class=".pl-pds">"</span></span>`;

	// 	it("should convert quoted string node with annotation", () => {
	// 		const {result, bytesConsumed} = annotations.convertQuotedStringNode(quotedStringNode, annsByStartByte, 0);
	// 		expect(result.startsWith("<a")).to.eql(true);
	// 		expect(result.contains(openingTag)).to.eql(true);
	// 		expect(result.contains(quotedStringInnerHTML)).to.eql(true);
	// 		expect(result.endsWith("</a>")).to.eql(true);
	// 		expect(bytesConsumed).to.eql(5);
	// 	});

	// 	it("should convert quoted string node without annotation", () => {
	// 		const {result, bytesConsumed} = annotations.convertQuotedStringNode(quotedStringNode, annsByStartByte, 1);
	// 		expect(result).to.eql(quotedStringInnerHTML);
	// 		expect(bytesConsumed).to.eql(3);
	// 	});
	// });

	// describe("convertNode", () => {
	// 	it("should return for tag with attribute", () => {
	// 		expect(annotations.getOpeningTag(elementNode)).to.eql(`<span class="pd-s">`);
	// 	});
	// });
});
