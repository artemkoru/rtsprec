export const E2E_USERNAME = process.env.E2E_USERNAME || 'admin';
export const E2E_PASSWORD = process.env.E2E_PASSWORD || 'change-me-local';

// Signs in through the login page, the same way a user does.
export async function login(page) {
  await page.goto('/login');
  await page.getByLabel('Username').fill(E2E_USERNAME);
  await page.getByLabel('Password').fill(E2E_PASSWORD);
  await page.getByRole('button', { name: 'Sign in' }).click();
  await page.waitForURL('**/live');
}

// Opens the Cameras tab of the Settings page.
export async function openCamerasTab(page) {
  await page.goto('/settings');
  await page.getByRole('button', { name: 'Cameras' }).click();
}

export function uniqueCameraName(label) {
  return `E2E ${label} ${Date.now()}`;
}
