import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import { LocalizationProvider } from './i18n'
import { ErrorBoundary } from './ErrorBoundary'
ReactDOM.createRoot(document.getElementById('root')!).render(<React.StrictMode><ErrorBoundary><LocalizationProvider><App /></LocalizationProvider></ErrorBoundary></React.StrictMode>)
