// Thin wrapper around fetch for the backend API.
// Requests are same-origin, so the session cookie set by login is sent automatically.
// A 401 means there is no valid session: the user is sent to the login page.

export class ApiError extends Error {
  constructor(status, message) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

export async function apiFetch(path, options = {}) {
  const response = await fetch(path, {
    credentials: 'same-origin',
    ...options,
  })

  if (response.status === 401) {
    if (window.location.pathname !== '/login') {
      window.location.assign('/login')
    }
    throw new ApiError(401, 'session expired or missing')
  }

  if (!response.ok) {
    throw new ApiError(response.status, `request failed with status ${response.status}`)
  }

  return response.json()
}

// checkSession reports whether the browser has a valid session. It uses the
// cameras endpoint, which is protected: 200 means logged in, 401 means not.
export async function checkSession() {
  const response = await fetch('/api/cameras', { credentials: 'same-origin' })
  return response.status !== 401
}

// login posts credentials; the backend sets the session cookie on success.
// Returns true on success and false for wrong credentials.
export async function login(username, password) {
  const response = await fetch('/api/auth/login', {
    method: 'POST',
    credentials: 'same-origin',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  })
  return response.ok
}

// logout asks the backend to end the session. The backend answers with a fixed
// response and does not invalidate the token, so the caller only leaves the UI.
export async function logout() {
  await fetch('/api/auth/logout', { method: 'POST', credentials: 'same-origin' })
}

// postJSON sends a JSON body with POST. It follows the same rules as apiFetch:
// 401 redirects to the login page, other failures throw ApiError with the status
// and the server's message when the body has one.
export async function postJSON(path, body) {
  const response = await fetch(path, {
    method: 'POST',
    credentials: 'same-origin',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })

  if (response.status === 401) {
    if (window.location.pathname !== '/login') {
      window.location.assign('/login')
    }
    throw new ApiError(401, 'session expired or missing')
  }

  if (!response.ok) {
    const text = (await response.text().catch(() => '')).trim()
    throw new ApiError(response.status, text || `request failed with status ${response.status}`)
  }

  const text = await response.text()
  return text ? JSON.parse(text) : null
}
