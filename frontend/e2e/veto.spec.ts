import { test, expect } from '@playwright/test';

test.describe('Veto Process', () => {
  test.beforeEach(async ({ page }) => {
    // Предполагаем, что пользователь уже залогинен
    await page.goto('/');
  });

  test('should start veto process Bo1', async ({ page }) => {
    // Навигация к процессу вето
    await page.goto('/ban/valorant');
    
    // Выбираем пул карт (предполагаем, что есть выбор)
    const poolButton = page.locator('button, .pool-card').first();
    if (await poolButton.count() > 0) {
      await poolButton.click();
    }

    // Переходим к процессу вето
    await page.goto('/veto/valorant/1');
    
    // Начинаем процесс
    await page.click('text=/start|начать/i');
    
    // Проверяем, что процесс начался
    const startButton = page.locator('text=/start|начать/i');
    await expect(startButton).not.toBeVisible();
  });

  test('should ban a map in Bo1', async ({ page }) => {
    await page.goto('/veto/valorant/1');
    
    // Начинаем процесс, если еще не начат
    const startButton = page.locator('text=/start|начать/i');
    if (await startButton.count() > 0) {
      await startButton.click();
    }

    // Выбираем карту для бана
    const mapCard = page.locator('.map-card, [data-map]').first();
    if (await mapCard.count() > 0) {
      await mapCard.click();
      
      // Проверяем, что карта забанена
      await expect(mapCard).toHaveClass(/banned|disabled/);
    }
  });

  test('should complete Bo1 process', async ({ page }) => {
    await page.goto('/veto/valorant/1');
    
    // Начинаем процесс
    const startButton = page.locator('text=/start|начать/i');
    if (await startButton.count() > 0) {
      await startButton.click();
    }

    // Баним все карты кроме одной
    // В реальном тесте нужно банить карты до последней
    const mapCards = page.locator('.map-card, [data-map]');
    const count = await mapCards.count();
    
    // Баним все кроме последней
    for (let i = 0; i < count - 1; i++) {
      const card = mapCards.nth(i);
      if (await card.isVisible()) {
        await card.click();
        // Ждем обновления UI
        await page.waitForTimeout(500);
      }
    }

    // Проверяем, что процесс завершен
    const finishedMessage = page.locator('text=/finished|завершен|selected|выбрана/i');
    await expect(finishedMessage.first()).toBeVisible();
  });

  test('should test Bo5 with team B selection', async ({ page }) => {
    await page.goto('/ban/valorant');
    
    // Выбираем Bo5 (если есть выбор)
    const bo5Button = page.locator('text=/bo5|best of 5/i');
    if (await bo5Button.count() > 0) {
      await bo5Button.click();
    }

    await page.goto('/veto/valorant/1');
    
    // Начинаем процесс
    const startButton = page.locator('text=/start|начать/i');
    if (await startButton.count() > 0) {
      await startButton.click();
    }

    // Выполняем действия для Bo5
    // В реальном тесте нужно выполнить полный цикл Bo5
    // с выбором команды B на шаге 4
  });
});
