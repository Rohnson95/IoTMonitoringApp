import React from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import WeatherWarnings from './components/WeatherWarnings';
import Navbar from './components/NavBar';
import Hero from './components/Hero';
import Login from './components/Login';
import Register from './components/Register';
import Dashboard from './components/Dashboard';
import { AuthProvider } from './context/AuthContext';
import ProtectedRoute from './components/ProtectedRoute';

const App: React.FC = () => {
  return (
    <AuthProvider>
      <Router>
        <div className="min-h-screen bg-gray-50 text-gray-800 flex flex-col">
          {/* Navbar */}
          <Navbar />

          {/* Hero */}
          {/* <header className="bg-white shadow">
            <div className="container mx-auto px-4 py-3 flex justify-between items-center">
              <Link to="/" className="text-xl font-bold text-indigo-600">
                IoT Weather Risk Monitor
              </Link>
              <nav>
                <Link to="/weather-warnings" className="text-gray-700 hover:text-indigo-600 mr-4">
                  Weather Warnings
                </Link>
                <Link to="/login" className="text-gray-700 hover:text-indigo-600 mr-4">
                  Login
                </Link>
                <Link to="/register" className="text-gray-700 hover:text-indigo-600">
                  Register
                </Link>
                <Link to="/dashboard" className="text-gray-700 hover:text-indigo-600">
                  Dashboard
                </Link>
              </nav>
            </div>
          </header> */}

          {/* Main Content */}
          <main className="flex-1">
            <Routes>
              <Route path="/" element={<HomePage />} />
              <Route path="/weather-warnings" element={<WeatherWarnings />} />
              <Route path="/login" element={<Login />} />
              <Route path="/register" element={<Register />} />
              <Route path="/dashboard"
               element={<ProtectedRoute>
                  <Dashboard />
               </ProtectedRoute>} />
            </Routes>
          </main>

          {/* Footer */}
          <footer className="bg-white mt-auto">
            <div className="container mx-auto px-4 py-4 text-center text-gray-600">
              &copy; {new Date().getFullYear()} IoT Weather Risk Monitor
            </div>
          </footer>
        </div>
      </Router>
    </AuthProvider>
  );
};

const HomePage: React.FC = () => (
  <div className="bg-gradient-to-r from-blue-500 to-indigo-600 text-white py-16 px-4">
    <div className="container mx-auto max-w-3xl text-center">
      <h1 className="text-4xl font-bold mb-4">Protect Your IoT Sensors</h1>
      <p className="text-lg mb-6">
        Stay informed about potential dangers from weather and fire risks in real time.
        Keep your devices safe and ensure constant monitoring with our easy-to-use platform.
      </p>
      <Link
        to="/weather-warnings"
        className="inline-block bg-white text-indigo-600 font-medium px-6 py-3 rounded-lg hover:bg-gray-100 transition"
      >
        Explore Weather Warnings
      </Link>
    </div>
  </div>
);

export default App;
