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

// Safe number formatter
const formatValue = (value: number | undefined | null) => {
  if (value === undefined || value === null) return '0.00';
  return value.toFixed(2);
};

// Component to display a sensor group (3 values)
const SensorGroup = ({ title, x = 0, y = 0, z = 0, unit }: { 
  title: string; 
  x?: number; 
  y?: number; 
  z?: number; 
  unit: string 
}) => (
  <div className="p-4 bg-white rounded-lg shadow-sm">
    <h3 className="text-lg font-semibold mb-2">{title}</h3>
    <div className="grid grid-cols-3 gap-4">
      <div>
        <span className="text-gray-600">X:</span>
        <span className="ml-2 font-mono">{formatValue(x)}</span>
      </div>
      <div>
        <span className="text-gray-600">Y:</span>
        <span className="ml-2 font-mono">{formatValue(y)}</span>
      </div>
      <div>
        <span className="text-gray-600">Z:</span>
        <span className="ml-2 font-mono">{formatValue(z)}</span>
      </div>
    </div>
    <div className="text-xs text-gray-500 mt-1">Unit: {unit}</div>
  </div>
);

export default function SensorDashboard() {
  const [data, setData] = useState<Partial<SensorData>>({});
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let ws: WebSocket | null = null;
    
    const connect = () => {
      try {
        ws = new WebSocket('ws://localhost:8080/sensor');

        ws.onopen = () => {
          console.log('Connected to sensor');
          setConnected(true);
          setError(null);
        };

        ws.onmessage = (event) => {
          try {
            const newData = JSON.parse(event.data);
            setData(newData);
          } catch (e) {
            console.error('Error parsing data:', e);
          }
        };

        ws.onerror = () => {
          setError('Connection error');
          setConnected(false);
        };

        ws.onclose = () => {
          setConnected(false);
          // Try to reconnect after 2 seconds
          setTimeout(connect, 2000);
        };
      } catch (e) {
        console.error('WebSocket creation error:', e);
        setError('Failed to create connection');
        setConnected(false);
      }
    };

    connect();

    return () => {
      if (ws) {
        ws.close();
      }
    };
  }, []);

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <h1 className="text-2xl font-bold mb-6">Solar Car IMU Sensor Dashboard</h1>
      
      <div className="mb-4 flex items-center">
        <div className={`h-2 w-2 rounded-full ${connected ? 'bg-green-500' : 'bg-red-500'} mr-2`}></div>
        <span className={`text-sm ${connected ? 'text-green-700' : 'text-red-700'}`}>
          {connected ? 'Connected to sensor' : 'Disconnected from sensor'}
        </span>
        {error && (
          <div className="ml-4 text-sm text-red-600">
            <AlertCircle className="inline-block h-4 w-4 mr-1" />
            {error}
          </div>
        )}
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
              <span className="ml-2 font-mono">{formatValue(data.eulHeading)}°</span>
            </div>
            <div>
              <span className="text-gray-600">Roll:</span>
              <span className="ml-2 font-mono">{formatValue(data.eulRoll)}°</span>
            </div>
            <div>
              <span className="text-gray-600">Pitch:</span>
              <span className="ml-2 font-mono">{formatValue(data.eulPitch)}°</span>
            </div>
          </div>
        </div>

        <div className="p-4 bg-white rounded-lg shadow-sm">
          <h3 className="text-lg font-semibold mb-2">Quaternion</h3>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <span className="text-gray-600">W:</span>
              <span className="ml-2 font-mono">{formatValue(data.quaW)}</span>
            </div>
            <div>
              <span className="text-gray-600">X:</span>
              <span className="ml-2 font-mono">{formatValue(data.quaX)}</span>
            </div>
            <div>
              <span className="text-gray-600">Y:</span>
              <span className="ml-2 font-mono">{formatValue(data.quaY)}</span>
            </div>
            <div>
              <span className="text-gray-600">Z:</span>
              <span className="ml-2 font-mono">{formatValue(data.quaZ)}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}