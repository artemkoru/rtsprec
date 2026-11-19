import { useEffect, useState } from 'react'
import {
  BrowserRouter,
  NavLink,
  Navigate,
  Outlet,
  Route,
  Routes,
  useNavigate,
} from 'react-router-dom'
import { checkSession, logout } from './api.js'
import Archive from './pages/Archive.jsx'
import Live from './pages/Live.jsx'
import Login from './pages/Login.jsx'
import Settings from './pages/Settings.jsx'
import './App.css'

// RequireSession renders its children only when the browser has a valid session.
// Otherwise it sends the user to the login page.
function RequireSession({ children }) {
  const [state, setState] = useState('checking')

  useEffect(() => {
    let active = true
    checkSession()
      .then((ok) => {
        if (active) setState(ok ? 'ok' : 'anonymous')
      })
      .catch(() => {
        if (active) setState('anonymous')
      })
    return () => {
      active = false
    }
  }, [])

  if (state === 'checking') {
    return <div className="container">Checking session...</div>
  }
  if (state === 'anonymous') {
    return <Navigate to="/login" replace />
  }
  return children
}

function Shell() {
  const navigate = useNavigate()

  const handleLogout = async () => {
    await logout()
    navigate('/login', { replace: true })
  }

  return (
    <div className="app-shell">
      <nav className="app-nav">
        <NavLink to="/live">Live</NavLink>
        <NavLink to="/archive">Archive</NavLink>
        <NavLink to="/settings">Settings</NavLink>
        <button type="button" onClick={handleLogout}>
          Logout
        </button>
      </nav>
      <main>
        <Outlet />
      </main>
    </div>
  )
}

function ProtectedLayout() {
  return (
    <RequireSession>
      <Shell />
    </RequireSession>
  )
}

function NotFound() {
  return (
    <div className="container">
      <h1>Page not found</h1>
    </div>
  )
}

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route element={<ProtectedLayout />}>
          <Route path="/" element={<Navigate to="/live" replace />} />
          <Route path="/live" element={<Live />} />
          <Route path="/archive" element={<Archive />} />
          <Route path="/settings" element={<Settings />} />
          <Route path="*" element={<NotFound />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}

export default App
