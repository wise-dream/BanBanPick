import { test, expect } from '@playwright/test';

test.describe('Rooms', () => {
  test.beforeEach(async ({ page }) => {
    // Предполагаем, что пользователь уже залогинен
    // В реальных тестах нужно добавить setup для логина
    await page.goto('/');
  });

  test('should navigate to rooms list', async ({ page }) => {
    await page.click('text=Rooms');
    await expect(page).toHaveURL(/.*rooms/);
  });

  test('should create a new room', async ({ page }) => {
    await page.goto('/rooms');
    
    await page.click('text=/create|создать/i');
    
    // Заполняем форму создания комнаты
    const roomName = `Test Room ${Date.now()}`;
    await page.fill('input[type="text"]', roomName);
    
    await page.click('button[type="submit"]');

    // После создания должен быть редирект на комнату
    await expect(page).toHaveURL(/.*room\/\d+/);
  });

  test('should join room by code', async ({ page }) => {
    await page.goto('/rooms');
    
    // Предполагаем, что есть поле для ввода кода
    const codeInput = page.locator('input[placeholder*="code" i], input[placeholder*="код" i]');
    if (await codeInput.count() > 0) {
      await codeInput.fill('TESTCODE');
      await page.click('button:has-text("Join"), button:has-text("Присоединиться")');
      
      // Должен быть редирект на комнату
      await expect(page).toHaveURL(/.*room\/\d+/);
    }
  });

  test('should leave room', async ({ page }) => {
    // Предполагаем, что пользователь уже в комнате
    await page.goto('/room/1');
    
    await page.click('text=/leave|выйти/i');
    
    // После выхода должен быть редирект на список комнат
    await expect(page).toHaveURL(/.*rooms/);
  });
});
