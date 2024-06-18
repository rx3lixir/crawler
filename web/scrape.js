const puppeteer = require("puppeteer");

let browser;
let pagePool = [];

async function initializeBrowser() {
  if (!browser) {
    browser = await puppeteer.launch();
  }
}

async function getPage() {
  if (pagePool.length > 0) {
    return pagePool.pop();
  }
  return await browser.newPage();
}

async function releasePage(page) {
  pagePool.push(page);
}

async function scrapePage(url, selector) {
  const page = await getPage();

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
    await releasePage(page);
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
  await initializeBrowser();
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
