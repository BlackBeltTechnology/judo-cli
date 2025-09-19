import { test, expect } from '@playwright/test';

test.describe('Service Controls E2E Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    
    // Open service panel
    await page.locator('text=▶ Services').click();
    await expect(page.locator('.service-panel.open')).toBeVisible();
  });

  test('should display all service controls', async ({ page }) => {
    // Check all service controls are present
    const serviceControls = page.locator('.service-control');
    await expect(serviceControls).toHaveCount(4); // All Services + 3 individual
    
    // Check individual services
    await expect(page.locator('text=karaf')).toBeVisible();
    await expect(page.locator('text=postgresql')).toBeVisible();
    await expect(page.locator('text=keycloak')).toBeVisible();
    await expect(page.locator('text=All Services')).toBeVisible();
    
    // Check status indicators
    const statusElements = page.locator('.service-status');
    await expect(statusElements).toHaveCount(3);
    
    // Check button groups
    const buttonGroups = page.locator('.service-buttons');
    await expect(buttonGroups).toHaveCount(4);
  });

  test('should handle service start/stop button states', async ({ page }) => {
    // Test button states based on service status
    // This assumes the mock server returns appropriate statuses
    
    const karafControl = page.locator('text=karaf').locator('..').locator('..');
    const postgresqlControl = page.locator('text=postgresql').locator('..').locator('..');
    
    // Check button visibility and states
    const karafStartButton = karafControl.locator('.btn-service-start');
    const karafStopButton = karafControl.locator('.btn-service-stop');
    
    const postgresqlStartButton = postgresqlControl.locator('.btn-service-start');
    const postgresqlStopButton = postgresqlControl.locator('.btn-service-stop');
    
    await expect(karafStartButton).toBeVisible();
    await expect(karafStopButton).toBeVisible();
    await expect(postgresqlStartButton).toBeVisible();
    await expect(postgresqlStopButton).toBeVisible();
    
    // Buttons should not be disabled initially (mock dependent)
    // This test verifies the UI structure, actual state depends on API response
  });

  test('should handle all services control', async ({ page }) => {
    const allServicesControl = page.locator('text=All Services').locator('..').locator('..');
    
    const startAllButton = allServicesControl.locator('.btn-service-start');
    const stopAllButton = allServicesControl.locator('.btn-service-stop');
    
    await expect(startAllButton).toBeVisible();
    await expect(stopAllButton).toBeVisible();
    await expect(startAllButton).toHaveText('Start All');
    await expect(stopAllButton).toHaveText('Stop All');
  });

  test('should maintain service panel state during interactions', async ({ page }) => {
    // Service panel should remain open during interactions
    await expect(page.locator('.service-panel.open')).toBeVisible();
    
    // Interact with some elements
    await page.locator('select.source-selector').selectOption('karaf');
    await expect(page.locator('.service-panel.open')).toBeVisible();
    
    // Click a service button
    const karafStartButton = page.locator('text=karaf')
      .locator('..').locator('..')
      .locator('.btn-service-start');
    await karafStartButton.click();
    
    // Panel should still be open
    await expect(page.locator('.service-panel.open')).toBeVisible();
  });

  test('should handle service panel responsiveness', async ({ page }) => {
    // Test different screen sizes with service panel open
    await page.setViewportSize({ width: 1024, height: 768 });
    await expect(page.locator('.service-panel.open')).toBeVisible();
    await expect(page.locator('.terminal-container')).toBeVisible();
    
    await page.setViewportSize({ width: 768, height: 1024 });
    await expect(page.locator('.service-panel.open')).toBeVisible();
    await expect(page.locator('.terminal-container')).toBeVisible();
    
    await page.setViewportSize({ width: 375, height: 667 });
    await expect(page.locator('.service-panel.open')).toBeVisible();
    await expect(page.locator('.terminal-container')).toBeVisible();
  });

  test('should handle service status updates', async ({ page }) => {
    // This test would verify that service status updates are reflected in the UI
    // For now, just verify the status elements exist
    const statusElements = page.locator('.service-status');
    await expect(statusElements).toHaveCount(3);
    
    // Check that status elements have valid status text
    const statusTexts = await statusElements.allTextContents();
    statusTexts.forEach(text => {
      expect(['running', 'stopped', 'starting', 'stopping']).toContain(text.toLowerCase());
    });
  });

  test('should handle loading states for service buttons', async ({ page }) => {
    // Click a service button to trigger loading state
    const karafStartButton = page.locator('text=karaf')
      .locator('..').locator('..')
      .locator('.btn-service-start');
    
    await karafStartButton.click();
    
    // Button should show loading state (text change)
    // This depends on the actual implementation - may need to wait for API response
    await page.waitForTimeout(500);
    
    // Verify button is still interactive or in loading state
    await expect(karafStartButton).toBeVisible();
  });
});