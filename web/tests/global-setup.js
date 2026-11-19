// Fails fast with a clear message when the backend behind the dev proxy is not running.
// A running backend answers GET /api/cameras with 401 when no session is sent.
export default async function globalSetup(config) {
  const baseURL = config.projects[0].use.baseURL || process.env.E2E_BASE_URL || 'http://localhost:5173';
  let status;
  try {
    const response = await fetch(new URL('/api/cameras', baseURL));
    status = response.status;
  } catch (error) {
    throw new Error(`E2E: cannot reach ${baseURL}/api/cameras. Start the backend (make docker-up) and retry.`);
  }
  if (status !== 401) {
    throw new Error(`E2E: ${baseURL}/api/cameras answered ${status}, expected 401 from the backend. Check the dev proxy and the backend.`);
  }
}
