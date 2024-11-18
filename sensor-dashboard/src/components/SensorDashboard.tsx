'use client';

import { useState, useEffect } from 'react';
import { AlertCircle } from 'lucide-react';

interface SensorData {
  accX: number; accY: number; accZ: number;
  magX: number; magY: number; magZ: number;
  gyrX: number; gyrY: number; gyrZ: number;
  liaX: number; liaY: number; liaZ: number;
  grvX: number; grvY: number; grvZ: number;
  eulHeading: number; eulRoll: number; eulPitch: number;
  quaW: number; quaX: number; quaY: number; quaZ: number;
}

const defaultSensorData: SensorData = {
  accX: 0, accY: 0, accZ: 0,
  magX: 0, magY: 0, magZ: 0,
  gyrX: 0, gyrY: 0, gyrZ: 0,
  liaX: 0, liaY: 0, liaZ: 0,
  grvX: 0, grvY: 0, grvZ: 0,
  eulHeading: 0, eulRoll: 0, eulPitch: 0,
  quaW: 0, quaX: 0, quaY: 0, quaZ: 0
};

// Safe number formatter
const formatNumber = (value: number | undefined, decimals: number = 2): string => {
  if (value === undefined || value === null) return '0.00';
  return value.toFixed(decimals);
};

const SensorGroup = ({ title, x, y, z, unit }: { title: string; x: number; y: number; z: number; unit: string }) => (
  <div className="p-4 bg-white rounded-lg shadow-sm">
    <h3 className="text-lg font-semibold mb-2">{title}</h3>
    <div className="grid grid-cols-3 gap-4">
      <div>
        <span className="text-gray-600">X:</span>
        <span className="ml-2 font-mono">{formatNumber(x)}</span>
      </div>
      <div>
        <span className="text-gray-600">Y:</span>
        <span className="ml-2 font-mono">{formatNumber(y)}</span>
      </div>
      <div>
        <span className="text-gray-600">Z:</span>
        <span className="ml-2 font-mono">{formatNumber(z)}</span>
      </div>
    </div>
    <div className="text-xs text-gray-500 mt-1">Unit: {unit}</div>
  </div>
);

const SensorDashboard = () => {
  const [data, setData] = useState<SensorData>(defaultSensorData);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let ws: WebSocket;
    let reconnectTimeout: NodeJS.Timeout;

    const connectWebSocket = () => {
      ws = new WebSocket('ws://localhost:8080/ws');

      ws.onopen = () => {
        setConnected(true);
        setError(null);
        console.log('Connected to sensor server');
      };

      ws.onmessage = (event) => {
        try {
          const newData = JSON.parse(event.data);
          setData(newData);
        } catch (e) {
          console.error('Error parsing sensor data:', e);
        }
      };

      ws.onerror = (event) => {
        setError('Failed to connect to sensor server');
        setConnected(false);
        console.error('WebSocket error:', event);
      };

      ws.onclose = () => {
        setConnected(false);
        console.log('Disconnected from sensor server');
        
        // Try to reconnect after 5 seconds
        reconnectTimeout = setTimeout(connectWebSocket, 5000);
      };
    };

    connectWebSocket();

    return () => {
      if (ws) {
        ws.close();
      }
      if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
      }
    };
  }, []);

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">IMU Sensor Dashboard</h1>
      
      <div className="mb-4 flex items-center">
        <div className={`h-2 w-2 rounded-full ${connected ? 'bg-green-500' : 'bg-red-500'} mr-2`}></div>
        <span className={`text-sm ${connected ? 'text-green-700' : 'text-red-700'}`}>
          {connected ? 'Connected to sensor' : 'Disconnected'}
        </span>
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <SensorGroup
          title="Accelerometer"
          x={data.accX}
          y={data.accY}
          z={data.accZ}
          unit="mg"
        />
        
        <SensorGroup
          title="Magnetometer"
          x={data.magX}
          y={data.magY}
          z={data.magZ}
          unit="µT"
        />
        
        <SensorGroup
          title="Gyroscope"
          x={data.gyrX}
          y={data.gyrY}
          z={data.gyrZ}
          unit="dps"
        />
        
        <SensorGroup
          title="Linear Acceleration"
          x={data.liaX}
          y={data.liaY}
          z={data.liaZ}
          unit="mg"
        />
        
        <SensorGroup
          title="Gravity Vector"
          x={data.grvX}
          y={data.grvY}
          z={data.grvZ}
          unit="mg"
        />

        <div className="p-4 bg-white rounded-lg shadow-sm">
          <h3 className="text-lg font-semibold mb-2">Euler Angles</h3>
          <div className="grid grid-cols-1 gap-4">
            <div>
              <span className="text-gray-600">Heading:</span>
              <span className="ml-2 font-mono">{formatNumber(data.eulHeading)}°</span>
            </div>
            <div>
              <span className="text-gray-600">Roll:</span>
              <span className="ml-2 font-mono">{formatNumber(data.eulRoll)}°</span>
            </div>
            <div>
              <span className="text-gray-600">Pitch:</span>
              <span className="ml-2 font-mono">{formatNumber(data.eulPitch)}°</span>
            </div>
          </div>
        </div>

        <div className="p-4 bg-white rounded-lg shadow-sm">
          <h3 className="text-lg font-semibold mb-2">Quaternion</h3>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <span className="text-gray-600">W:</span>
              <span className="ml-2 font-mono">{formatNumber(data.quaW, 3)}</span>
            </div>
            <div>
              <span className="text-gray-600">X:</span>
              <span className="ml-2 font-mono">{formatNumber(data.quaX, 3)}</span>
            </div>
            <div>
              <span className="text-gray-600">Y:</span>
              <span className="ml-2 font-mono">{formatNumber(data.quaY, 3)}</span>
            </div>
            <div>
              <span className="text-gray-600">Z:</span>
              <span className="ml-2 font-mono">{formatNumber(data.quaZ, 3)}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SensorDashboard;