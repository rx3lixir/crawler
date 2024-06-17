const puppeteer = require("puppeteer");
const fs = require("fs");

(async () => {
  let browser;
  try {
    const url = process.argv[2];
    const selector = process.argv[3];
    const filePath = process.argv[4];

    browser = await puppeteer.launch();
    const page = await browser.newPage();

    console.log(`Navigating to ${url}`);
    await page.goto(url, { waitUntil: "networkidle2", timeout: 60000 });

    console.log(`Waiting for selector ${selector}`);
    await page.waitForSelector(selector, { timeout: 60000 });

    const html = await page.content();
    fs.writeFileSync(filePath, html);

    console.log(`Page saved to ${filePath}`);
  } catch (err) {
    console.error(`Error: ${err.message}`);
    process.exit(1);
  } finally {
    if (browser) {
      await browser.close();
    }
  }
})();
