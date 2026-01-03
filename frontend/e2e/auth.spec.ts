import { test, expect } from '@playwright/test';

test.describe('Authentication', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should navigate to login page', async ({ page }) => {
    await page.click('text=Login');
    await expect(page).toHaveURL(/.*login/);
  });

  test('should show validation errors on empty login form', async ({ page }) => {
    await page.goto('/login');
    await page.click('button[type="submit"]');
    
    // Проверяем наличие ошибок валидации
    const emailError = page.locator('text=/email/i');
    const passwordError = page.locator('text=/password/i');
    
    await expect(emailError.first()).toBeVisible();
    await expect(passwordError.first()).toBeVisible();
  });

  test('should register new user', async ({ page }) => {
    await page.goto('/register');
    
    const timestamp = Date.now();
    const email = `test${timestamp}@example.com`;
    const username = `testuser${timestamp}`;
    const password = 'password123';

    await page.fill('input[type="email"]', email);
    await page.fill('input[type="text"]', username);
    await page.fill('input[type="password"]', password);
    await page.fill('input[type="password"]:nth-of-type(2)', password);

    await page.click('button[type="submit"]');

    // После регистрации должен быть редирект
    await expect(page).toHaveURL('/');
  });

  test('should login with valid credentials', async ({ page }) => {
    await page.goto('/login');
    
    // Предполагаем, что пользователь уже зарегистрирован
    const email = 'test@example.com';
    const password = 'password123';

    await page.fill('input[type="email"]', email);
    await page.fill('input[type="password"]', password);
    await page.click('button[type="submit"]');

    // После входа должен быть редирект
    await expect(page).toHaveURL('/');
  });

  test('should show error on invalid credentials', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('input[type="email"]', 'invalid@example.com');
    await page.fill('input[type="password"]', 'wrongpassword');
    await page.click('button[type="submit"]');

    // Должно появиться сообщение об ошибке
    const errorMessage = page.locator('text=/invalid|error|неверн/i');
    await expect(errorMessage.first()).toBeVisible();
  });

  test('should logout user', async ({ page }) => {
    // Предполагаем, что пользователь уже залогинен
    // (можно добавить setup для логина)
    
    await page.goto('/');
    await page.click('text=Logout');
    
    // После выхода должен быть редирект на главную или login
    await expect(page).toHaveURL(/\/(login)?$/);
  });
});
