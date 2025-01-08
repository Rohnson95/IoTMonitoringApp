import React from 'react';
import { Link } from 'react-router-dom';


const Hero: React.FC = () => (
  <div className="relative bg-gradient-to-r from-blue-500 to-indigo-600 text-white p-8 text-center">
    <div className="max-w-4xl mx-auto">
      <h1 className="text-4xl font-bold mb-4">IoT Weather Risk Monitor</h1>
      <p className="text-lg mb-6">
        Stay ahead of dangerous weather conditions that could affect your IoT sensors.
        Monitor potential threats and protect your environment with real-time updates.
      </p>
      <Link to='/weather-warnings'>
      
      <button
        className="bg-white text-indigo-600 font-medium px-6 py-3 rounded-lg hover:bg-gray-100 transition"
      >
        Explore Weather Warnings
      </button>
      </Link>
    </div>
  </div>
);

export default Hero;