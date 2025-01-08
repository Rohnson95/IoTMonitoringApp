import React from 'react';
import { Link } from 'react-router-dom';
import { AuthContext } from '../context/AuthContext';

const Navbar: React.FC = () => {
    const { token, logout } = React.useContext(AuthContext); // Access token and logout function

    return (
        <header className="bg-white shadow">
            <div className="container mx-auto px-4 py-3 flex justify-between items-center">
                <Link to="/" className="text-xl font-bold text-indigo-600">
                    IoT Weather Risk Monitor
                </Link>
                <nav>
                    <Link to="/weather-warnings" className="text-gray-700 hover:text-indigo-600 mr-4">
                        Weather Warnings
                    </Link>
                    {token ? ( // Conditionally render links based on login status
                        <>
                            <Link to="/dashboard" className="text-gray-700 hover:text-indigo-600 mr-4">
                                Dashboard
                            </Link>
                            <button
                                onClick={logout}
                                className="text-gray-700 hover:text-red-600"
                            >
                                Log Out
                            </button>
                        </>
                    ) : (
                        <>
                            <Link to="/login" className="text-gray-700 hover:text-indigo-600 mr-4">
                                Login
                            </Link>
                            <Link to="/register" className="text-gray-700 hover:text-indigo-600">
                                Register
                            </Link>
                        </>
                    )}
                </nav>
            </div>
        </header>
    );
};

export default Navbar;
