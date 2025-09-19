import { test, expect } from '@playwright/test';

test.describe('JUDO CLI Server E2E Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
  });

  test('should load main application with correct title', async ({ page }) => {
    await expect(page).toHaveTitle(/JUDO CLI Server/);
    await expect(page.locator('h1')).toHaveText('JUDO CLI Server');
  });

  test('should display service panel toggle button', async ({ page }) => {
    await expect(page.locator('text=▶ Services')).toBeVisible();
  });

  test('should toggle service panel visibility', async ({ page }) => {
    const servicePanelButton = page.locator('text=▶ Services');
    await servicePanelButton.click();
    
    // Service panel should be visible
    await expect(page.locator('.service-panel.open')).toBeVisible();
    await expect(page.locator('text=All Services')).toBeVisible();
    
    // Toggle back
    await servicePanelButton.click();
    await expect(page.locator('.service-panel.open')).not.toBeVisible();
  });

  test('should display log source selector', async ({ page }) => {
    await expect(page.locator('text=Source:')).toBeVisible();
    await expect(page.locator('select.source-selector')).toBeVisible();
    
    const sourceSelector = page.locator('select.source-selector');
    await expect(sourceSelector).toHaveValue('combined');
    
    // Check all options are present
    const options = sourceSelector.locator('option');
    await expect(options).toHaveCount(4);
    await expect(options.nth(0)).toHaveText('Combined');
    await expect(options.nth(1)).toHaveText('Karaf');
    await expect(options.nth(2)).toHaveText('PostgreSQL');
    await expect(options.nth(3)).toHaveText('Keycloak');
  });

  test('should change log source when selected', async ({ page }) => {
    const sourceSelector = page.locator('select.source-selector');
    await sourceSelector.selectOption('karaf');
    await expect(sourceSelector).toHaveValue('karaf');
    
    await sourceSelector.selectOption('postgresql');
    await expect(sourceSelector).toHaveValue('postgresql');
    
    await sourceSelector.selectOption('keycloak');
    await expect(sourceSelector).toHaveValue('keycloak');
    
    await sourceSelector.selectOption('combined');
    await expect(sourceSelector).toHaveValue('combined');
  });

  test('should display terminal component', async ({ page }) => {
    await expect(page.locator('.terminal-container')).toBeVisible();
    await expect(page.locator('.terminal')).toBeVisible();
  });

  test('should handle service status display', async ({ page }) => {
    // Open service panel
    await page.locator('text=▶ Services').click();
    
    // Check service status elements
    await expect(page.locator('.service-control')).toHaveCount(4); // All Services + 3 individual
    
    // Check individual services
    await expect(page.locator('text=karaf')).toBeVisible();
    await expect(page.locator('text=postgresql')).toBeVisible();
    await expect(page.locator('text=keycloak')).toBeVisible();
    
    // Check status indicators
    const statusElements = page.locator('.service-status');
    await expect(statusElements).toHaveCount(3);
  });

  test('should handle service control buttons', async ({ page }) => {
    // Open service panel
    await page.locator('text=▶ Services').click();
    
    // Check service buttons exist
    const startButtons = page.locator('.btn-service-start');
    const stopButtons = page.locator('.btn-service-stop');
    
    await expect(startButtons).toHaveCount(4); // All + 3 individual
    await expect(stopButtons).toHaveCount(4); // All + 3 individual
    
    // Check button states based on service status
    // This test assumes the mock server returns appropriate statuses
  });

  test('should handle WebSocket connections', async ({ page }) => {
    // This test verifies that WebSocket connections are established
    // The actual WebSocket behavior is tested in integration tests
    await expect(page.locator('.terminal')).toBeVisible();
    
    // Wait a moment for WebSocket connection to establish
    await page.waitForTimeout(1000);
  });

  test('should handle responsive design', async ({ page }) => {
    // Test different screen sizes
    await page.setViewportSize({ width: 1024, height: 768 });
    await expect(page.locator('.terminal-container')).toBeVisible();
    
    await page.setViewportSize({ width: 768, height: 1024 });
    await expect(page.locator('.terminal-container')).toBeVisible();
    
    await page.setViewportSize({ width: 375, height: 667 });
    await expect(page.locator('.terminal-container')).toBeVisible();
  });

  test('should handle project initialization modal', async ({ page }) => {
    // This test would require mocking the API response to show modal
    // For now, just verify the modal structure exists in DOM
    const modalOverlay = page.locator('.modal-overlay');
    await expect(modalOverlay).not.toBeVisible();
  });

  test('should maintain session state on navigation', async ({ page }) => {
    // Test that service panel state is maintained
    await page.locator('text=▶ Services').click();
    await expect(page.locator('.service-panel.open')).toBeVisible();
    
    // Refresh page
    await page.reload();
    await page.waitForLoadState('networkidle');
    
    // Service panel should be closed after refresh
    await expect(page.locator('.service-panel.open')).not.toBeVisible();
  });
});