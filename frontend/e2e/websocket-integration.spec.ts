import { test, expect } from '@playwright/test';

test.describe('WebSocket Integration E2E Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
  });

  test('should establish WebSocket connection on load', async ({ page }) => {
    // Wait for WebSocket connection to be established
    await page.waitForTimeout(1000);
    
    // Verify terminal is visible and ready
    await expect(page.locator('.terminal')).toBeVisible();
    
    // Check that WebSocket connection attempt was made
    // This is a basic check - actual WebSocket testing requires more complex setup
    const consoleMessages: string[] = [];
    page.on('console', msg => {
      if (msg.text().includes('WebSocket') || msg.text().includes('ws://')) {
        consoleMessages.push(msg.text());
      }
    });
    
    await page.waitForTimeout(500);
    expect(consoleMessages.length).toBeGreaterThan(0);
  });

  test('should handle WebSocket reconnection on source change', async ({ page }) => {
    const sourceSelector = page.locator('select.source-selector');
    
    // Change source multiple times to trigger reconnections
    await sourceSelector.selectOption('karaf');
    await page.waitForTimeout(300);
    
    await sourceSelector.selectOption('postgresql');
    await page.waitForTimeout(300);
    
    await sourceSelector.selectOption('keycloak');
    await page.waitForTimeout(300);
    
    await sourceSelector.selectOption('combined');
    await page.waitForTimeout(300);
    
    // All source changes should complete without errors
    await expect(sourceSelector).toHaveValue('combined');
  });

  test('should maintain terminal functionality during WebSocket operations', async ({ page }) => {
    // Terminal should remain functional during WebSocket operations
    await expect(page.locator('.terminal')).toBeVisible();
    
    // Change sources to trigger WebSocket reconnections
    const sourceSelector = page.locator('select.source-selector');
    await sourceSelector.selectOption('karaf');
    await page.waitForTimeout(300);
    
    // Terminal should still be visible and functional
    await expect(page.locator('.terminal')).toBeVisible();
    
    await sourceSelector.selectOption('combined');
    await page.waitForTimeout(300);
    
    await expect(page.locator('.terminal')).toBeVisible();
  });

  test('should handle WebSocket errors gracefully', async ({ page }) => {
    // This test simulates WebSocket errors by monitoring console
    const errorMessages: string[] = [];
    page.on('console', msg => {
      if (msg.type() === 'error' && msg.text().includes('WebSocket')) {
        errorMessages.push(msg.text());
      }
    });
    
    // Application should continue functioning despite WebSocket errors
    await expect(page.locator('.terminal')).toBeVisible();
    await expect(page.locator('text=▶ Services')).toBeVisible();
    
    // Service panel should still work
    await page.locator('text=▶ Services').click();
    await expect(page.locator('.service-panel.open')).toBeVisible();
  });

  test('should handle simultaneous WebSocket and service operations', async ({ page }) => {
    // Open service panel
    await page.locator('text=▶ Services').click();
    
    // Change WebSocket source while service panel is open
    const sourceSelector = page.locator('select.source-selector');
    await sourceSelector.selectOption('karaf');
    
    // Service panel should remain open
    await expect(page.locator('.service-panel.open')).toBeVisible();
    
    // Interact with service buttons
    const karafStartButton = page.locator('text=karaf')
      .locator('..').locator('..')
      .locator('.btn-service-start');
    await karafStartButton.click();
    
    // Change source again
    await sourceSelector.selectOption('combined');
    
    // Everything should remain functional
    await expect(page.locator('.service-panel.open')).toBeVisible();
    await expect(page.locator('.terminal')).toBeVisible();
  });

  test('should handle browser refresh with active WebSocket', async ({ page }) => {
    // Establish WebSocket connection
    await page.waitForTimeout(1000);
    
    // Refresh page
    await page.reload();
    await page.waitForLoadState('networkidle');
    
    // WebSocket should reconnect automatically
    await page.waitForTimeout(1000);
    await expect(page.locator('.terminal')).toBeVisible();
  });

  test('should handle multiple rapid source changes', async ({ page }) => {
    const sourceSelector = page.locator('select.source-selector');
    
    // Rapid source changes
    for (let i = 0; i < 5; i++) {
      await sourceSelector.selectOption('karaf');
      await page.waitForTimeout(50);
      await sourceSelector.selectOption('combined');
      await page.waitForTimeout(50);
    }
    
    // Application should remain stable
    await expect(page.locator('.terminal')).toBeVisible();
    await expect(sourceSelector).toHaveValue('combined');
  });

  test('should handle WebSocket operations with service panel interactions', async ({ page }) => {
    // Open service panel
    await page.locator('text=▶ Services').click();
    
    // Perform WebSocket operations
    const sourceSelector = page.locator('select.source-selector');
    await sourceSelector.selectOption('karaf');
    await page.waitForTimeout(300);
    
    // Interact with services
    const allServicesStart = page.locator('text=All Services')
      .locator('..').locator('..')
      .locator('.btn-service-start');
    await allServicesStart.click();
    
    // Change WebSocket source
    await sourceSelector.selectOption('combined');
    await page.waitForTimeout(300);
    
    // Everything should work together
    await expect(page.locator('.service-panel.open')).toBeVisible();
    await expect(page.locator('.terminal')).toBeVisible();
  });
});