// src/components/WeatherWarnings.tsx
import React, { useEffect, useState, useContext } from 'react';
import axios from '../utils/axiosConfig';
import { Warning, WarningArea } from '../Types/warningTypes';
import SwedenMap from './SwedenMap';
import { AuthContext } from '../context/AuthContext';

interface Sensor {
  id: number;
  name: string;
  latitude: number;
  longitude: number;
  status: string;
  description?: string;
}

const WeatherWarnings: React.FC = () => {
  const { token } = useContext(AuthContext);
  const [warnings, setWarnings] = useState<Warning[]>([]);
  const [sensors, setSensors] = useState<Sensor[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string>('');

  // Sök- och pagineringsstater
  const [areaName, setAreaName] = useState<string>('');
  const [eventType, setEventType] = useState<string>('');
  const [page, setPage] = useState<number>(1);
  const [pageSize, setPageSize] = useState<number>(10);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const warningsResponse = await axios.get('/api/weather-warnings', {
          params: { eventType, areaName, page, pageSize },
        });
        setWarnings(warningsResponse.data.warnings);

        const sensorsResponse = await axios.get('/api/sensors');
        setSensors(sensorsResponse.data);
      } catch (err: any) {
        setError('Failed to fetch data');
        console.error('Fetch data error:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [eventType, areaName, page, pageSize]);

  if (loading) return <div className="text-center py-8">Loading...</div>;
  if (error) return <div className="text-center py-8 text-red-500">{error}</div>;

  // Samla alla WarningAreas i en enda array för kartan
  const allWarningAreas: WarningArea[] = [];
  warnings.forEach((warning) => {
    if (warning.warningAreas) {
      allWarningAreas.push(...warning.warningAreas);
    }
  });

  return (
    <div className="flex flex-col md:flex-row min-h-screen bg-gray-50">
      {/* Vänsterkolumn: sökning, paginering, varningar lista */}
      <div className="md:w-1/2 p-4 overflow-y-auto" style={{ maxHeight: '100vh' }}>
        <h2 className="text-2xl font-bold text-gray-800 mb-4">Weather Warnings</h2>

        {/* Sök och Filter */}
        <div className="flex flex-col md:flex-row items-center gap-4 mb-6">
          {/* Sök efter areaName */}
          <input
            type="text"
            value={areaName}
            onChange={(e) => setAreaName(e.target.value)}
            placeholder="Search for city or län..."
            className="flex-1 p-2 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-300"
          />
          {/* pageSize */}
          <input
            type="number"
            min={1}
            value={pageSize}
            onChange={(e) => setPageSize(Math.max(Number(e.target.value), 1))}
            className="w-24 p-2 border border-gray-300 rounded-lg shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-300"
          />
        </div>

        {/* Varningslista */}
        {warnings.map((warning) => (
          <div key={warning.id} className="bg-white p-6 shadow rounded-lg mb-4">
            <h3 className="text-xl font-semibold text-gray-700 mb-2">{warning.event.sv}</h3>
            <p className="text-gray-500 mb-2">
              <span className="font-medium">Event Code:</span> {warning.event.code}
            </p>
            <div>
              <span className="font-medium text-gray-700">Affected Areas:</span>
              <ul className="list-disc list-inside ml-4">
                {warning.warningAreas?.map((wa) => (
                  <li key={wa.id}>{wa.areaName?.sv || 'Unknown Area'}</li>
                ))}
              </ul>
            </div>
          </div>
        ))}

        {/* Paginering */}
        <div className="flex items-center justify-center gap-4 mt-6">
          <button
            onClick={() => setPage((prev) => Math.max(prev - 1, 1))}
            disabled={page === 1}
            className="px-4 py-2 bg-blue-100 text-blue-700 rounded-lg disabled:opacity-50"
          >
            Previous
          </button>
          <span>Page {page}</span>
          <button
            onClick={() => setPage((prev) => prev + 1)}
            className="px-4 py-2 bg-blue-100 text-blue-700 rounded-lg"
          >
            Next
          </button>
        </div>
      </div>

      {/* Högerkolumn: Karta */}
      <div className="md:w-1/2 h-screen p-4">
        <SwedenMap
          warningAreas={allWarningAreas}
          sensors={sensors}
        />
      </div>
    </div>
  );
};

export default WeatherWarnings;
