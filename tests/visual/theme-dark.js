// Script to set dark theme before visual regression test
module.exports = async (page, scenario, vp) => {
  await page.evaluate(() => {
    localStorage.setItem('theme', 'dark');
    document.documentElement.setAttribute('data-theme', 'dark');
  });
  await page.waitForTimeout(1000); // Wait for theme transition
};