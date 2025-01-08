// src/components/Dashboard.tsx
import React, { useContext, useEffect, useState } from 'react';
// import axios from 'axios';
import axios from '../utils/axiosConfig';
import { AuthContext } from '../context/AuthContext';

interface Sensor {
  id: number;
  name: string;
  latitude: number;
  longitude: number;
  status: string;
  description?: string;
}

interface Webhook {
  id: number;
  url: string;
}

const Dashboard: React.FC = () => {
  const { token } = useContext(AuthContext);
  const [sensors, setSensors] = useState<Sensor[]>([]);
  const [webhooks, setWebhooks] = useState<Webhook[]>([]);
  const [sensorForm, setSensorForm] = useState({ name: '', latitude: '', longitude: '', description: '' });
  const [webhookURL, setWebhookURL] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    if (token) {
      fetchSensors();
      fetchWebhooks();
    }
  }, [token]);

  const fetchSensors = async () => {
    try {
      const response = await axios.get('/api/sensors');
      console.log('Fetched sensors:', response.data); 
      setSensors(response.data);
    } catch (err: any) {
      setError('Error fetching sensors');
      console.error('Fetch sensors error:', err);
    }
  };

  const fetchWebhooks = async () => {
    try {
      const response = await axios.get('/api/webhooks'); // Interceptorn hanterar Authorization
      setWebhooks(response.data);
    } catch (err: any) {
      setError('Error fetching webhooks');
      console.error('Fetch webhooks error:', err);
    }
  };

  const handleAddSensor = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
  
    // Convert strings to numbers
    const latitude = parseFloat(sensorForm.latitude);
    const longitude = parseFloat(sensorForm.longitude);
  
    // Validate the numbers
    if (isNaN(latitude) || isNaN(longitude)) {
      setError('Latitude and Longitude must be valid numbers.');
      return;
    }
  
    try {
      await axios.post('/api/sensors', { 
        name: sensorForm.name, 
        latitude, 
        longitude, 
        description: sensorForm.description 
      });
      setSensorForm({ name: '', latitude: '', longitude: '', description: '' });
      fetchSensors();
    } catch (err: any) {
      setError('Error adding sensor');
      console.error('Add sensor error:', err);
    }
  };
  
  

  const handleDeleteSensor = async (id: number) => {
    console.log(`Deleting sensor with ID: ${id}`);
    setError('');
    try {
      await axios.delete(`/api/sensors?id=${id}`);
      fetchSensors();
    } catch (err: any) {
      setError('Error deleting sensor');
      console.error('Delete sensor error:', err);
    }
  };

  const handleAddWebhook = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    try {
      await axios.post('/api/webhooks', { url: webhookURL });
      setWebhookURL('');
      fetchWebhooks();
    } catch (err: any) {
      setError('Error adding webhook');
      console.error('Add webhook error:', err);
    }
  };

  const handleDeleteWebhook = async (id: number) => {
    setError('');
    try {
      await axios.delete(`/api/webhooks?id=${id}`);
      fetchWebhooks();
    } catch (err: any) {
      setError('Error deleting webhook');
      console.error('Delete webhook error:', err);
    }
  };

  return (
    <div className="container mx-auto p-4">
      <h2 className="text-2xl font-bold mb-4">Dashboard</h2>
      {error && <div className="text-red-500 mb-4">{error}</div>}

      {/* Sensors Management */}
      <div className="mb-8">
        <h3 className="text-xl font-semibold mb-2">Manage Sensors</h3>
        <form onSubmit={handleAddSensor} className="mb-4">
  <div className="flex flex-col md:flex-row gap-4">
    <input
      type="text"
      required
      placeholder="Sensor Name"
      value={sensorForm.name}
      onChange={(e) => setSensorForm({ ...sensorForm, name: e.target.value })}
      className="flex-1 p-2 border border-gray-300 rounded"
    />
    <input
        type="number"
        required
        placeholder="Latitude"
        value={sensorForm.latitude}
        onChange={(e) => setSensorForm({ ...sensorForm, latitude: e.target.value })}
        className="flex-1 p-2 border border-gray-300 rounded"
    />
    <input
  type="number"
  required
  placeholder="Longitude"
  value={sensorForm.longitude}
  onChange={(e) => setSensorForm({ ...sensorForm, longitude: e.target.value })}
  className="flex-1 p-2 border border-gray-300 rounded"
/>
  </div>
  <div className="mt-2">
    <textarea
      placeholder="Description (optional)"
      value={sensorForm.description}
      onChange={(e) => setSensorForm({ ...sensorForm, description: e.target.value })}
      className="w-full p-2 border border-gray-300 rounded"
    />
  </div>
  <button
    type="submit"
    className="mt-2 bg-indigo-600 text-white px-4 py-2 rounded hover:bg-indigo-700 transition"
  >
    Add Sensor
  </button>
</form>

        <div>
          {sensors.length === 0 ? (
            <p>No sensors found.</p>
          ) : (
            <ul>
              {sensors.map((sensor, index) => (
                <li key={sensor.id || index} className="flex justify-between items-center bg-gray-100 p-2 mb-2 rounded">
                  <div>
                    <strong>{sensor.name}</strong> - {sensor.status ||"Unknown"}
                  </div>
                  <button
                    onClick={() => handleDeleteSensor(sensor.id)}
                    className="text-red-500 hover:text-red-700"
                  >
                    Delete
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      {/* Webhooks Management */}
      <div className="mb-8">
        <h3 className="text-xl font-semibold mb-2">Manage Webhooks</h3>
        <form onSubmit={handleAddWebhook} className="mb-4">
          <div className="flex flex-col md:flex-row gap-4">
            <input
              type="url"
              required
              placeholder="Webhook URL"
              value={webhookURL}
              onChange={(e) => setWebhookURL(e.target.value)}
              className="flex-1 p-2 border border-gray-300 rounded"
            />
            <button
              type="submit"
              className="bg-indigo-600 text-white px-4 py-2 rounded hover:bg-indigo-700 transition"
            >
              Add Webhook
            </button>
          </div>
        </form>

        <div>
          {webhooks.length === 0 ? (
            <p>No webhooks found.</p>
          ) : (
            <ul>
              {webhooks.map((webhook) => (
                <li key={webhook.id} className="flex justify-between items-center bg-gray-100 p-2 mb-2 rounded">
                  <div>{webhook.url}</div>
                  <button
                    onClick={() => handleDeleteWebhook(webhook.id)}
                    className="text-red-500 hover:text-red-700"
                  >
                    Delete
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
