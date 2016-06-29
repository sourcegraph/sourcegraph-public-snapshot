import path from "path";
import webdriver from "selenium-webdriver";
import {expect} from "chai";
import {delay, startChromeDriver, buildWebDriver} from "../utils";

describe("inject app", function test() {
	let driver;
	this.timeout(10000); // whole test must complete in 10s

	before(async () => {
		await startChromeDriver();
		const extPath = path.resolve("build");
		driver = buildWebDriver(extPath);
		await driver.get("https://github.com/gorilla/mux");
	});

	after(async () => driver.quit());

	it("should open github.com/gorilla/mux", async () => {
		const title = await driver.getTitle();
		expect(title).to.equal("GitHub - gorilla/mux: A powerful URL router and dispatcher for golang.");
	});

	describe("#sourcegraph-app-bootstrap", function test() {

		it("should get added to the document", async () => {
			await driver.wait(
				() => driver.findElements(webdriver.By.id("sourcegraph-app-bootstrap"))
					.then((elems) => elems.length === 1),
				1000
			);
		});

		it("should be hidden", async () => {
			await driver.wait(
				() => driver.findElement(webdriver.By.id("sourcegraph-app-bootstrap"))
					.then((elem) => elem.getCssValue("display"))
					.then((val) => val === "none"),
				1000
			);
		});
	});

	describe("#sourcegraph-app-background", function test() {
		it("should get added to the document", async () => {
			await driver.wait(
				() => driver.findElements(webdriver.By.id("sourcegraph-app-background"))
					.then((elems) => elems.length === 1),
				1000
			);
		});
	});

	describe("#sourcegraph-search-button", function test() {

		it("should get added to the document", async () => {
			await driver.wait(
				() => driver.findElements(webdriver.By.id("sourcegraph-search-button"))
					.then((elems) => elems.length === 1),
				1000
			);
		});

		describe("sourcegraph-search-frame", function test() {

			it("should not be on document initially", async () => {
				await driver.wait(
					() => driver.findElements(webdriver.By.id("sourcegraph-search-frame"))
						.then((elems) => elems.length === 0),
					1000
				);
			});

			it("should be toggled when search button is clicked", async () => {
				await driver.wait(
					() => driver.findElement(webdriver.By.id("sourcegraph-search-button"))
						.then((elem) => elem.click())
						.then(() => driver.findElement(webdriver.By.id("sourcegraph-search-frame")))
						.then((elem) => elem.getCssValue("display"))
						.then((val) => val === "block"),
					1000
				);
			});

			it("should return results from the Sourcegraph API", async () => {
				await driver.wait(
					() => driver.findElement(webdriver.By.className("sg-input"))
						.then((elem) => elem.sendKeys("mux"))
						.then(() => driver.sleep(1000)) // wait for results to be returned
						.then(() => driver.findElement(webdriver.By.className("sg-search-result"))),
					1000
				);
			});

			it("should be toggled when escape is hit", async () => {
				await driver.wait(
					() => driver.findElement(webdriver.By.tagName("body"))
						.then((elem) => elem.sendKeys(webdriver.Key.ESCAPE))
						.then(() => driver.findElement(webdriver.By.id("sourcegraph-search-frame")))
						.then((elem) => elem.getCssValue("display"))
						.then((val) => val === "none"),
					1000
				);
			});
		});
	});
});
