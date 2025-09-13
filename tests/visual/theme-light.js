// Script to set light theme before visual regression test
module.exports = async (page, scenario, vp) => {
  await page.evaluate(() => {
    localStorage.setItem('theme', 'light');
    document.documentElement.setAttribute('data-theme', 'light');
  });
  await page.waitForTimeout(1000); // Wait for theme transition
};