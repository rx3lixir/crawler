const puppeteer = require("puppeteer");

let browser;

async function scrapePage(url, selector) {
  if (!browser) {
    browser = await puppeteer.launch();
  }

  const page = await browser.newPage();

  try {
    console.log(`Navigating to ${url}`);
    await page.goto(url, { waitUntil: "networkidle2", timeout: 60000 });

    console.log(`Waiting for selector ${selector}`);
    await page.waitForSelector(selector, { timeout: 60000 });

    const html = await page.content();
    console.log(`Page content fetched`);
    return html;
  } catch (err) {
    console.error(`Error: ${err.message}`);
    return null;
  } finally {
    await page.close();
  }
}

async function closeBrowser() {
  if (browser) {
    await browser.close();
  }
}

process.on("exit", closeBrowser);
process.on("SIGINT", closeBrowser);

(async () => {
  const url = process.argv[2];
  const selector = process.argv[3];

  if (!url || !selector) {
    console.error("URL and selector must be provided as arguments.");
    process.exit(1);
  }

  const html = await scrapePage(url, selector);
  if (html) {
    console.log(html);
  } else {
    console.error("Failed to fetch page content.");
  }

  await closeBrowser();
})();
