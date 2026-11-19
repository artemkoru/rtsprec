import { test, expect } from '@playwright/test';
import { login, openCamerasTab, uniqueCameraName } from './helpers.js';

// Fills the add-camera form. Name and RTSP URL are optional arguments so tests can leave them empty.
async function fillCameraForm(page, { name = '', rtspUrl = '' } = {}) {
  await page.getByPlaceholder('Camera Name').fill(name);
  await page.getByPlaceholder('RTSP URL').fill(rtspUrl);
}

test.beforeEach(async ({ page }) => {
  await login(page);
  await openCamerasTab(page);
});

test('adding a camera shows it in the list without a page reload', async ({ page }) => {
  const name = uniqueCameraName('add');
  await fillCameraForm(page, { name, rtspUrl: 'rtsp://127.0.0.1:8554/e2e' });

  const created = page.waitForResponse(
    (r) => r.url().endsWith('/api/cameras') && r.request().method() === 'POST',
  );
  await page.getByRole('button', { name: 'Add Camera' }).click();
  expect((await created).status()).toBe(201);

  await expect(page.locator('.camera-config h4', { hasText: name })).toBeVisible();
  // the form is cleared after a successful add
  await expect(page.getByPlaceholder('Camera Name')).toHaveValue('');
  await expect(page.getByPlaceholder('RTSP URL')).toHaveValue('');
});

test('empty name or RTSP URL shows the required message and sends no request', async ({ page }) => {
  let posted = false;
  page.on('request', (r) => {
    if (r.url().endsWith('/api/cameras') && r.method() === 'POST') posted = true;
  });

  await fillCameraForm(page, { name: '', rtspUrl: 'rtsp://127.0.0.1:8554/e2e' });
  await page.getByRole('button', { name: 'Add Camera' }).click();
  await expect(page.getByText('Name is required')).toBeVisible();
  await expect(page.getByPlaceholder('Camera Name')).toHaveAttribute('aria-invalid', 'true');

  await fillCameraForm(page, { name: 'E2E no url', rtspUrl: '   ' });
  await page.getByRole('button', { name: 'Add Camera' }).click();
  await expect(page.getByText('RTSP URL is required')).toBeVisible();

  // give any stray request a moment to appear before asserting none was sent
  await page.waitForTimeout(300);
  expect(posted).toBe(false);
});

test('a server error keeps the entered values and shows the error next to the form', async ({ page }) => {
  await page.route('**/api/cameras', (route) => {
    if (route.request().method() === 'POST') {
      return route.fulfill({ status: 400, contentType: 'text/plain', body: 'rtsp_url is not valid' });
    }
    return route.continue();
  });

  const name = uniqueCameraName('error');
  const rtspUrl = 'not-a-url';
  await fillCameraForm(page, { name, rtspUrl });
  await page.getByRole('button', { name: 'Add Camera' }).click();

  await expect(page.getByRole('alert')).toHaveText('rtsp_url is not valid');
  await expect(page.getByPlaceholder('Camera Name')).toHaveValue(name);
  await expect(page.getByPlaceholder('RTSP URL')).toHaveValue(rtspUrl);
  await expect(page.locator('.camera-config h4', { hasText: name })).toHaveCount(0);
});
