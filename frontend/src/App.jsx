// App.jsx - Main React application component
// Root component that sets up routing and global state

import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { Provider } from 'react-redux';
import { store } from './store/store';
import Dashboard from './pages/Dashboard';
import ContentDetail from './pages/ContentDetail';
import './index.css';

function App() {
  return (
    <Provider store={store}>
      <Router>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/content/:id" element={<ContentDetail />} />
        </Routes>
      </Router>
    </Provider>
  );
}

export default App;
